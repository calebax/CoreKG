import { describe, it, expect, vi } from "vitest";

vi.mock("@corekg/logger", () => ({
  createLogger: () => ({
    info: vi.fn(),
    error: vi.fn(),
    warn: vi.fn(),
  }),
}));

vi.mock("@nats-io/jetstream", () => ({
  AckPolicy: { Explicit: "explicit" },
}));

import { TaskConsumer } from "./consumer.js";

function createHangingIterator() {
  let resolveNext: ((v: IterableIterator<never>) => void) | null = null;
  const iter: any = {
    [Symbol.asyncIterator]: () => iter,
    next: () => new Promise((resolve) => { resolveNext = resolve as any; }),
    close: vi.fn().mockImplementation(() => {
      if (resolveNext) resolveNext({ value: undefined, done: true } as any);
      return Promise.resolve();
    }),
    closed: () => Promise.resolve(undefined),
    status: () => ({ [Symbol.asyncIterator]: async function* () {} }),
  };
  return iter;
}

describe("TaskConsumer", () => {
  it("should add consumer config when not found", async () => {
    const handler = vi.fn();
    const mockConsumer = { consume: vi.fn().mockResolvedValue(createHangingIterator()) };
    const mockJs = { consumers: { get: vi.fn().mockResolvedValue(mockConsumer) } };
    const mockJsm = {
      consumers: {
        info: vi.fn().mockRejectedValue(new Error("not found")),
        add: vi.fn(),
      },
    };

    const consumer = new TaskConsumer(mockJs as any, mockJsm as any, {
      stream: "TEST",
      subject: "core.task.test",
      durableName: "test-durable",
      ackWaitMs: 60000,
    }, handler);

    const p = consumer.start();
    await new Promise((r) => setTimeout(r, 50));
    await consumer.stop();
    await p;

    expect(mockJsm.consumers.add).toHaveBeenCalledWith("TEST", expect.objectContaining({
      durable_name: "test-durable",
      filter_subject: "core.task.test",
      ack_wait: 60000 * 1_000_000,
    }));
  });

  it("should skip add when consumer exists", async () => {
    const handler = vi.fn();
    const mockConsumer = { consume: vi.fn().mockResolvedValue(createHangingIterator()) };
    const mockJs = { consumers: { get: vi.fn().mockResolvedValue(mockConsumer) } };
    const mockJsm = {
      consumers: {
        info: vi.fn().mockResolvedValue({}),
        add: vi.fn(),
      },
    };

    const consumer = new TaskConsumer(mockJs as any, mockJsm as any, {
      stream: "TEST",
      subject: "core.task.test",
      durableName: "test-durable",
    }, handler);

    const p = consumer.start();
    await new Promise((r) => setTimeout(r, 50));
    await consumer.stop();
    await p;

    expect(mockJsm.consumers.add).not.toHaveBeenCalled();
  });

  it("should ack message on success", async () => {
    const handler = vi.fn().mockResolvedValue({ status: "success" });
    const mockMsg = {
      json: () => ({ task_id: "t1", payload: { key: "val" } }),
      ack: vi.fn(),
      nak: vi.fn(),
    };
    let callCount = 0;
    const mockConsumer = {
      consume: vi.fn().mockImplementation(async () => {
        callCount++;
        if (callCount === 1) {
          return {
            [Symbol.asyncIterator]: async function* () { yield mockMsg; },
            close: vi.fn().mockResolvedValue(undefined),
            closed: () => Promise.resolve(undefined),
            status: () => ({ [Symbol.asyncIterator]: async function* () {} }),
          };
        }
        return createHangingIterator();
      }),
    };
    const mockJs = { consumers: { get: vi.fn().mockResolvedValue(mockConsumer) } };
    const mockJsm = {
      consumers: { info: vi.fn().mockResolvedValue({}), add: vi.fn() },
    };

    const consumer = new TaskConsumer(mockJs as any, mockJsm as any, {
      stream: "TEST",
      subject: "core.task.test",
      durableName: "test-durable",
    }, handler);

    const p = consumer.start();
    await new Promise((r) => setTimeout(r, 50));
    await consumer.stop();
    await p;

    expect(handler).toHaveBeenCalledWith("t1", { key: "val" });
    expect(mockMsg.ack).toHaveBeenCalled();
  });
});
