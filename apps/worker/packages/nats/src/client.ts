import { connect, type NatsConnection } from "@nats-io/transport-node";
import { jetstream, jetstreamManager, RetentionPolicy, type JetStreamClient, type JetStreamManager } from "@nats-io/jetstream";
import type { AppConfig } from "@corekg/config";

export interface NATSClient {
  nc: NatsConnection;
  js: JetStreamClient;
  jsm: JetStreamManager;
}

export async function createNATSClient(config: AppConfig): Promise<NATSClient> {
  const nc = await connect({ servers: config.nats.url });
  const js = jetstream(nc);
  const jsm = await jetstreamManager(nc);

  try {
    await jsm.streams.info(config.nats.stream);
  } catch {
    await jsm.streams.add({
      name: config.nats.stream,
      subjects: ["core.task.>"],
      retention: RetentionPolicy.Workqueue,
    });
  }

  return { nc, js, jsm };
}
