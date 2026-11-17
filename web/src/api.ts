import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { BattleService } from './gen/chill/v1/battle_pb.ts';

export const battleApi = createClient(BattleService, createConnectTransport({ baseUrl: location.origin }));
