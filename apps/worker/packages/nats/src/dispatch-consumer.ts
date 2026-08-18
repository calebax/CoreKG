import type { JetStreamClient, JetStreamManager, ConsumerMessages } from "@nats-io/jetstream";
import { AckPolicy, RetentionPolicy } from "@nats-io/jetstream";
import { createLogger } from "@corekg/logger";
import { TaskPayloadSchema } from "./types.js";
import type { ResultPublisher } from "./result-publisher.js";
import { STREAM_NAMES, SUBJECT_WILDCARDS } from "@corekg/rpc";

const logger = createLogger("dispatch-consumer");

export interface DispatchConsumerOptions {
  handlerName: string;
  dispatchSubject: string;
  resultSubject: string;
}

export interface DispatchHandlerResult {
  status: "success" | "fail";
  result?: unknown;
  error?: string;
}

export type DispatchHandlerFn<TContext = unknown> = (ctx: TContext, payload: ReturnType<typeof TaskPayloadSchema.parse>) => Promise<DispatchHandlerResult>;

export interface DispatchContext {
  [key: string]: unknown;
}

export class DispatchConsumer<TContext = unknown> {
  private stopped = false;
  private activeIter: ConsumerMessages | null = null;

  constructor(
    private js: JetStreamClient,
    private jsm: JetStreamManager,
    private options: DispatchConsumerOptions,
    private handler: DispatchHandlerFn<TContext>,
    private ctx: TContext,
    private resultPublisher: ResultPublisher,
    private workerId: string,
  ) {}

  async start(): Promise<void> {
    const streamName = STREAM_NAMES.dispatch;
    const durableName = `worker-${this.options.handlerName}`;
    const ackWaitNanos = 300_000 * 1_000_000;

    try {
      await this.jsm.streams.info(streamName);
    } catch {
      await this.jsm.streams.add({
        name: streamName,
        subjects: [SUBJECT_WILDCARDS.dispatchAll],
        retention: RetentionPolicy.Workqueue,
      });
    }

    try {
      await this.jsm.consumers.info(streamName, durableName);
    } catch {
      await this.jsm.consumers.add(streamName, {
        durable_name: durableName,
        filter_subject: this.options.dispatchSubject,
        ack_policy: AckPolicy.Explicit,
        ack_wait: ackWaitNanos,
        max_ack_pending: 2,
      });
    }

    const consumer = await this.js.consumers.get(streamName, durableName);

    logger.info(
      { handler: this.options.handlerName, subject: this.options.dispatchSubject },
      "dispatch consumer started",
    );

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
            const raw = msg.json<unknown>();
            const payload = TaskPayloadSchema.parse(raw);
            const taskId = Number(payload.task_id ?? 0);
            const taskType = payload.task_type;

            const result = await this.handler(this.ctx, payload);

            await this.resultPublisher.publish(this.options.resultSubject, taskId, taskType, result);
            msg.ack();
            logger.info({ taskId, handler: this.options.handlerName, status: result.status }, "task completed");
          } catch (err) {
            const error = err instanceof Error ? err.message : String(err);
            logger.error({ err, handler: this.options.handlerName }, "task failed");
            msg.nak();
          }
        }
        this.activeIter = null;
      } catch (err) {
        this.activeIter = null;
        if (this.stopped) break;
        logger.error({ err, handler: this.options.handlerName }, "consume error, retrying");
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
