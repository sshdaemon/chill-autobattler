// Every three.js loader (glTF, its buffers and textures, textures) asks for
// versioned URLs (assets.ts). Imported first by scene/BattleView.tsx: scene
// modules start loads when they are evaluated (Nature.tsx preloads its
// models), and ES modules run in import order.
import * as THREE from 'three';
import { assetUrl } from '../assets.ts';

THREE.DefaultLoadingManager.setURLModifier(assetUrl);
