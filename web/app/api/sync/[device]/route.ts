import { checkDevice, config } from "@/lib/config";
import { getStore } from "@/lib/deps";
import { HttpError, onlyKeys, readJson, route } from "@/lib/http";
import { authenticate } from "@/lib/session";
import { checkFiles } from "@/lib/sync";

export const runtime = "nodejs";

// PUT (Bearer) {files: {path: contents}} -> {device, updated, files: count}
export const PUT = route(async (req: Request, ctx: { params: Promise<{ device: string }> }) => {
  const session = await authenticate(req);
  const device = checkDevice((await ctx.params).device);
  if (device !== session.device) throw new HttpError(403, "a device may only write its own export");
  const body = await readJson(req, config.syncMaxBytes());
  onlyKeys(body, ["files"]);
  const files = checkFiles(body.files);
  const updated = new Date();
  await getStore().putExport(session.user, device, files, updated);
  return Response.json({ device, updated: updated.toISOString(), files: Object.keys(files).length });
});
