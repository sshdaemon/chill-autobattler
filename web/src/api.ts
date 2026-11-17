import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { BattleService } from './gen/chill/v1/battle_pb.ts';
import { RunService } from './gen/chill/v1/run_pb.ts';
import { withToken } from './session.ts';

const transport = createConnectTransport({ baseUrl: location.origin, interceptors: [withToken] });
export const battleApi = createClient(BattleService, transport);
export const runApi = createClient(RunService, transport);
