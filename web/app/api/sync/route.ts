import { getStore } from "@/lib/deps";
import { route } from "@/lib/http";
import { authenticate } from "@/lib/session";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

// GET (Bearer) -> {devices: [{device, updated, files}]}: every other device's export.
export const GET = route(async (req) => {
  const { user, device } = await authenticate(req);
  const all = await getStore().listExports(user);
  return Response.json({ devices: all.filter((e) => e.device !== device) });
});
