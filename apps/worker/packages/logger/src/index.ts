import pino from "pino";

export function createLogger(name: string, level = "info"): pino.Logger {
  return pino({
    name,
    level,
    formatters: {
      level(label) {
        return { level: label };
      },
    },
    timestamp: pino.stdTimeFunctions.isoTime,
  });
}

export function withContext(logger: pino.Logger, fields: Record<string, unknown>): pino.Logger {
  return logger.child(fields);
}
