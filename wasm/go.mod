module chill/wasm

go 1.27.4

require (
	chill/pb v0.0.0
	chill/sim v0.0.0
	google.golang.org/protobuf v1.36.12
)

replace chill/sim => ../sim

replace chill/pb => ../pb
