import type { JetStreamClient, JetStreamManager, ConsumerMessages } from "@nats-io/jetstream";
import { AckPolicy } from "@nats-io/jetstream";
import { createLogger } from "@corekg/logger";

const logger = createLogger("nats-consumer");

export interface TaskHandler {
  (taskId: string, payload: unknown): Promise<{ status: "success" | "fail"; result?: unknown; error?: string }>;
}

export interface TaskConsumerOptions {
  stream: string;
  subject: string;
  durableName: string;
  maxAckPending?: number;
  ackWaitMs?: number;
}

export class TaskConsumer {
  private stopped = false;
  private activeIter: ConsumerMessages | null = null;

  constructor(
    private js: JetStreamClient,
    private jsm: JetStreamManager,
    private options: TaskConsumerOptions,
    private handler: TaskHandler,
  ) {}

  async start(): Promise<void> {
    const ackWaitNanos = (this.options.ackWaitMs ?? 300_000) * 1_000_000;

    try {
      await this.jsm.consumers.info(this.options.stream, this.options.durableName);
    } catch {
      await this.jsm.consumers.add(this.options.stream, {
        durable_name: this.options.durableName,
        filter_subject: this.options.subject,
        ack_policy: AckPolicy.Explicit,
        ack_wait: ackWaitNanos,
        max_ack_pending: this.options.maxAckPending ?? 2,
      });
    }

    const consumer = await this.js.consumers.get(this.options.stream, this.options.durableName);

    logger.info({ subject: this.options.subject }, "consumer started");

    while (!this.stopped) {
      try {
        const iter = await consumer.consume({ expires: 30_000, max_messages: 1 });
        this.activeIter = iter;
        for await (const msg of iter) {
          if (this.stopped) {
            msg.nak();
            await iter.close().catch(() => {});
            return;
          }
          try {
            const task = msg.json<{ task_id: string; payload: unknown }>();
            const result = await this.handler(task.task_id, task.payload);
            msg.ack();
            logger.info({ taskId: task.task_id, status: result.status }, "task completed");
          } catch (err) {
            logger.error({ err }, "task failed, nak");
            msg.nak();
          }
        }
        this.activeIter = null;
      } catch (err) {
        this.activeIter = null;
        if (this.stopped) break;
        logger.error({ err }, "consume error, retrying");
        await new Promise((r) => setTimeout(r, 1000));
      }
    }
  }

  async stop(): Promise<void> {
    this.stopped = true;
    if (this.activeIter) {
      await this.activeIter.close().catch(() => {});
    }
  }
}
