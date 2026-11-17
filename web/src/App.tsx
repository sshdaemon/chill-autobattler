import { Canvas } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import { useState } from 'react';
import { runBattle, type SimResult } from './sim/client.ts';
import { battleApi } from './api.ts';

// Placeholder scene: the six toy units at their start positions.
function Units() {
  return (
    <>
      {[0, 1].flatMap((team) => [-3, 0, 3].map((x) => (
        <mesh key={`${team}${x}`} position={[x, 0.9, team ? 10 : -10]} castShadow>
          <capsuleGeometry args={[0.4, 1]} />
          <meshStandardMaterial color={team ? '#e0524a' : '#3a8cff'} />
        </mesh>
      )))}
    </>
  );
}

export function App() {
  const [run, setRun] = useState<{ seed: bigint; version: string; result: SimResult }>();
  const [verdict, setVerdict] = useState<string>();

  const play = async () => {
    const seed = BigInt(Math.floor(Math.random() * 2 ** 48)); // TODO: the server issues seeds
    setVerdict(undefined);
    setRun({ seed, ...(await runBattle(seed)) });
  };
  const verify = async () => {
    if (!run) return;
    try {
      const r = await battleApi.verifyBattle({
        simVersion: run.version, seed: run.seed, maxTicks: 4000,
        claimedWinner: run.result.winner, claimedChecksum: BigInt(run.result.checksum),
      });
      setVerdict(r.fair ? `✅ fair (server got ${r.checksum})` : `❌ ${r.reason}`);
    } catch (e) {
      setVerdict(`⚠️ server unreachable: ${(e as Error).message}`);
    }
  };

  return (
    <>
      <Canvas shadows camera={{ position: [14, 16, 22], fov: 45 }}>
        <hemisphereLight args={['#dfefff', '#445533', 1.5]} />
        <directionalLight position={[10, 20, 8]} intensity={2.6} castShadow />
        <mesh rotation-x={-Math.PI / 2} receiveShadow>
          <planeGeometry args={[40, 40]} />
          <meshStandardMaterial color="#7a9a5a" />
        </mesh>
        <Units />
        <OrbitControls />
      </Canvas>
      <div style={{ position: 'fixed', top: 12, left: 12, background: '#fffd', padding: 12, borderRadius: 8, maxWidth: 360 }}>
        <b>Chill Autobattler</b> — pipeline check
        <div style={{ display: 'flex', gap: 8, margin: '8px 0' }}>
          <button onClick={play}>Run battle (WASM)</button>
          <button onClick={verify} disabled={!run}>Verify on server</button>
        </div>
        {run && <div data-testid="result">sim {run.version} · seed {run.seed.toString()}<br />winner {run.result.winner ? 'Red' : 'Blue'} after {run.result.ticks} ticks · checksum {run.result.checksum}</div>}
        {verdict && <div data-testid="verdict">{verdict}</div>}
      </div>
    </>
  );
}
