// Errors are JSON {error} with a status; every route runs inside route().

export class HttpError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
  }
}

export function jsonError(status: number, error: string, headers?: HeadersInit): Response {
  return Response.json({ error }, { status, headers });
}

// route wraps a handler so thrown HttpErrors become JSON responses and
// anything else is a 500 that names no request content.
export function route<A extends unknown[]>(
  handler: (req: Request, ...rest: A) => Promise<Response>,
): (req: Request, ...rest: A) => Promise<Response> {
  return async (req, ...rest) => {
    try {
      return await handler(req, ...rest);
    } catch (err) {
      if (err instanceof HttpError) return jsonError(err.status, err.message);
      console.error("unhandled error:", err instanceof Error ? err.name : typeof err);
      return jsonError(500, "internal error");
    }
  };
}

// readJson reads at most maxBytes of body and parses it as a JSON object.
export async function readJson(req: Request, maxBytes: number): Promise<Record<string, unknown>> {
  const declared = Number(req.headers.get("content-length") ?? "");
  if (Number.isFinite(declared) && declared > maxBytes) {
    throw new HttpError(413, `body is larger than ${maxBytes} bytes`);
  }
  if (!req.body) throw new HttpError(400, "body must be a JSON object");
  const reader = req.body.getReader();
  const chunks: Uint8Array[] = [];
  let size = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.byteLength;
    if (size > maxBytes) {
      await reader.cancel();
      throw new HttpError(413, `body is larger than ${maxBytes} bytes`);
    }
    chunks.push(value);
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(Buffer.concat(chunks).toString("utf8"));
  } catch {
    throw new HttpError(400, "body is not valid JSON");
  }
  if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
    throw new HttpError(400, "body must be a JSON object");
  }
  return parsed as Record<string, unknown>;
}

// onlyKeys refuses fields a route doesn't know.
export function onlyKeys(body: Record<string, unknown>, allowed: string[]): void {
  for (const k of Object.keys(body)) {
    if (!allowed.includes(k)) throw new HttpError(400, `unknown field "${k}"`);
  }
}
