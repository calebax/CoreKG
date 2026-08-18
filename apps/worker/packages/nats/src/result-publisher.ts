import type { JetStreamClient, JetStreamManager } from "@nats-io/jetstream";
import { RetentionPolicy } from "@nats-io/jetstream";
import { createLogger } from "@corekg/logger";
import { STREAM_NAMES, SUBJECT_WILDCARDS } from "@corekg/rpc";

const logger = createLogger("result-publisher");

export interface TaskResultMessage {
  task_id: number;
  worker_id: string;
  task_type: string;
  status: "success" | "fail";
  result?: string;
  error_message?: string;
}

export interface PublishResultInput {
  status: "success" | "fail";
  result?: unknown;
  error?: string;
}

export class ResultPublisher {
  constructor(
    private js: JetStreamClient,
    private jsm: JetStreamManager,
    private workerId: string,
  ) {}

  async ensureStream(): Promise<void> {
    try {
      await this.jsm.streams.info(STREAM_NAMES.result);
    } catch {
      await this.jsm.streams.add({
        name: STREAM_NAMES.result,
        subjects: [SUBJECT_WILDCARDS.resultAll],
        retention: RetentionPolicy.Limits,
      });
    }
  }

  async publish(
    resultSubject: string,
    taskId: number,
    taskType: string,
    result: PublishResultInput,
  ): Promise<void> {
    const msg: TaskResultMessage = {
      task_id: taskId,
      worker_id: this.workerId,
      task_type: taskType,
      status: result.status,
      result: result.result !== undefined ? JSON.stringify(result.result) : undefined,
      error_message: result.error,
    };
    const data = new TextEncoder().encode(JSON.stringify(msg));
    await this.js.publish(resultSubject, data);
    logger.info({ taskId, taskType, status: result.status, subject: resultSubject }, "result published");
  }
}
