import type { StackTrace } from "./client";

type NextFunction = (err?: unknown) => void;

interface IncomingMessage {
  method?: string;
  url?: string;
  originalUrl?: string;
}

interface ServerResponse {
  statusCode: number;
  on(event: string, listener: () => void): void;
}

interface FastifyRequest {
  method: string;
  url: string;
}

interface FastifyReply {
  statusCode: number;
  elapsedTime: number;
}

type FastifyDone = (err?: Error) => void;

interface FastifyInstance {
  addHook(
    name: "onResponse",
    handler: (request: FastifyRequest, reply: FastifyReply, done: FastifyDone) => void
  ): void;
}

function levelFromStatus(statusCode: number): "info" | "warn" | "error" {
  if (statusCode >= 500) return "error";
  if (statusCode >= 400) return "warn";
  return "info";
}

export function stacktraceMiddleware(client: StackTrace) {
  return (req: IncomingMessage, res: ServerResponse, next: NextFunction): void => {
    const start = Date.now();

    res.on("finish", () => {
      const duration = Date.now() - start;
      const method = req.method ?? "UNKNOWN";
      const path = (req as { originalUrl?: string }).originalUrl ?? req.url ?? "/";
      const status = res.statusCode;

      client[levelFromStatus(status)](`${method} ${path} ${status}`, {
        metadata: {
          method,
          path,
          statusCode: status,
          durationMs: duration,
        },
      });
    });

    next();
  };
}

export function stacktraceFastifyPlugin(client: StackTrace) {
  return (fastify: FastifyInstance, _opts: Record<string, unknown>, done: (err?: Error) => void): void => {
    fastify.addHook("onResponse", (request, reply, hookDone) => {
      const method = request.method;
      const path = request.url;
      const status = reply.statusCode;
      const duration = Math.round(reply.elapsedTime);

      client[levelFromStatus(status)](`${method} ${path} ${status}`, {
        metadata: {
          method,
          path,
          statusCode: status,
          durationMs: duration,
        },
      });

      hookDone();
    });

    done();
  };
}
