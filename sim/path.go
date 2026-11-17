package sim

import "container/heap"

// Integer path costs: a straight step is 10, a diagonal 14; wading costs
// 1/0.7 ≈ 1.43× so units go round ponds when the detour is about as quick.
const (
	costStraight = 10
	costDiagonal = 14
)

func wetCost(c int) int { return (c*10 + 3) / 7 } // ×1/0.7, rounded

type step struct{ cell, cost int }

// neighbours lists the walkable 8-neighbours of a cell, without cutting corners.
func (m *Map) neighbours(c int) []step {
	ci, cj := c%GridN, c/GridN
	out := make([]step, 0, 8)
	for dj := -1; dj <= 1; dj++ {
		for di := -1; di <= 1; di++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := ci+di, cj+dj
			if !m.Free(ni, nj) || (di != 0 && dj != 0 && (!m.Free(ci+di, cj) || !m.Free(ci, cj+dj))) {
				continue
			}
			cost := costStraight
			if di != 0 && dj != 0 {
				cost = costDiagonal
			}
			if m.Wet(ni, nj) {
				cost = wetCost(cost)
			}
			out = append(out, step{nj*GridN + ni, cost})
		}
	}
	return out
}

// octile distance heuristic in path-cost units
func octile(a, b int) int {
	dx, dz := a%GridN-b%GridN, a/GridN-b/GridN
	if dx < 0 {
		dx = -dx
	}
	if dz < 0 {
		dz = -dz
	}
	if dx < dz {
		dx, dz = dz, dx
	}
	return costStraight*dx + (costDiagonal-costStraight)*dz
}

type openItem struct{ f, h, cell int }
type openHeap []openItem

func (h openHeap) Len() int { return len(h) }
func (h openHeap) Less(i, j int) bool {
	a, b := h[i], h[j]
	if a.f != b.f {
		return a.f < b.f
	}
	if a.h != b.h {
		return a.h < b.h
	}
	return a.cell < b.cell // stable ties: same path everywhere
}
func (h openHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *openHeap) Push(x any)   { *h = append(*h, x.(openItem)) }
func (h *openHeap) Pop() any {
	old := *h
	it := old[len(old)-1]
	*h = old[:len(old)-1]
	return it
}

// FindPath returns waypoints from a to b (ending exactly at b), string-pulled
// so only needed corners remain, or nil if b can't be reached.
func (m *Map) FindPath(a, b Vec) []Vec {
	si, sj := m.NearestFree(cellOf(a))
	gi, gj := m.NearestFree(cellOf(b))
	start, goal := sj*GridN+si, gj*GridN+gi

	const unset = -1
	var g [GridN * GridN]int
	var came [GridN * GridN]int
	var closed [GridN * GridN]bool
	for i := range g {
		g[i], came[i] = 1<<30, unset
	}
	g[start] = 0
	open := &openHeap{{octile(start, goal), octile(start, goal), start}}
	for open.Len() > 0 {
		it := heap.Pop(open).(openItem)
		c := it.cell
		if closed[c] {
			continue
		}
		if c == goal {
			break
		}
		closed[c] = true
		for _, n := range m.neighbours(c) {
			if closed[n.cell] {
				continue
			}
			if ng := g[c] + n.cost; ng < g[n.cell] {
				g[n.cell], came[n.cell] = ng, c
				h := octile(n.cell, goal)
				heap.Push(open, openItem{ng + h, h, n.cell})
			}
		}
	}
	if came[goal] == unset && goal != start {
		return nil
	}
	var cells []int
	for c := goal; c != unset; c = came[c] {
		cells = append(cells, c)
	}
	for i, j := 0, len(cells)-1; i < j; i, j = i+1, j-1 {
		cells[i], cells[j] = cells[j], cells[i]
	}

	// string-pulling: keep only the corners we need; a shortcut may not wade
	// through more water than the grid path did
	pts := make([]Vec, len(cells))
	wetSum := make([]Fx, len(cells)+1)
	for k, c := range cells {
		pts[k] = V(CellCenter(c%GridN), CellCenter(c/GridN))
		wetSum[k+1] = wetSum[k]
		if m.wet[c] {
			wetSum[k+1] += Cell
		}
	}
	pts[len(pts)-1] = b
	okShortcut := func(from Vec, k, j int) bool {
		return m.LineOfSight(from, pts[j]) && m.WetLength(from, pts[j]) <= wetSum[j+1]-wetSum[k]+Cell
	}
	var out []Vec
	from, k := a, 0
	for k < len(pts)-1 {
		far := k + 1
		for j := len(pts) - 1; j > k; j-- {
			if okShortcut(from, k, j) {
				far = j
				break
			}
		}
		out = append(out, pts[far])
		from, k = pts[far], far
	}
	if len(out) == 0 {
		return []Vec{b}
	}
	return out
}
