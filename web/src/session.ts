// The anonymous player token: issued by the server on first visit, kept in
// localStorage, sent as a Bearer header on every RPC.
import { Code, ConnectError, createClient, type Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { RunService } from './gen/chill/v1/run_pb.ts';

const KEY = 'playerToken';
let token: string | null = null;
try { token = localStorage.getItem(KEY); } catch { /* storage blocked: a fresh player each load */ }

const bare = createClient(RunService, createConnectTransport({ baseUrl: location.origin }));

export async function ensurePlayer(force = false) {
  if (token && !force) return;
  token = (await bare.createPlayer({})).playerToken;
  try { localStorage.setItem(KEY, token); } catch { /* see above */ }
}

export const withToken: Interceptor = (next) => async (req) => {
  if (token) req.header.set('Authorization', `Bearer ${token}`);
  return next(req);
};

// Runs fn; if the server no longer knows our token (say, a wiped dev DB),
// registers a new player once and retries.
export async function authed<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn();
  } catch (e) {
    if (ConnectError.from(e).code !== Code.Unauthenticated) throw e;
    await ensurePlayer(true);
    return fn();
  }
}
