// The arena's files, without three.js, so the startup bundle can list them
// for prefetching (assets.ts).
export const NATURE_GLB = '/assets/nature/nature.glb';
export const TREES_GLB = '/assets/trees/trees.glb';

const TR = '/assets/trees/';
export const TREE_TEXTURES = {
  bark_oak: TR + 'bark_oak_color.jpg', bark_oak_n: TR + 'bark_oak_normal.jpg',
  bark_birch: TR + 'bark_birch_color.jpg', bark_birch_n: TR + 'bark_birch_normal.jpg',
  bark_pine: TR + 'bark_pine_color.jpg', bark_pine_n: TR + 'bark_pine_normal.jpg',
  leaves_oak: TR + 'leaves_oak.png', leaves_ash: TR + 'leaves_ash.png',
  leaves_aspen: TR + 'leaves_aspen.png', leaves_pine: TR + 'leaves_pine.png',
};

const T = '/assets/nature/';
export const GROUND_TEXTURES = {
  grass: T + 'Grass004_color.jpg', dirt: T + 'Ground037_color.jpg', mud: T + 'Ground054_color.jpg',
  grassN: T + 'Grass004_normal.jpg', dirtN: T + 'Ground037_normal.jpg', mudN: T + 'Ground054_normal.jpg',
};

export const ARENA_FILES = [NATURE_GLB, TREES_GLB, ...Object.values(TREE_TEXTURES), ...Object.values(GROUND_TEXTURES)];
