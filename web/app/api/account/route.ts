import { getStore } from "@/lib/deps";
import { route } from "@/lib/http";
import { authenticate } from "@/lib/session";

export const runtime = "nodejs";

// DELETE (Bearer) -> 204: deletes the account and everything the server
// keeps for it, from every device, and refuses every token issued so far.
// Progress kept on the devices themselves is not touched.
export const DELETE = route(async (req) => {
  const { user } = await authenticate(req);
  await getStore().deleteAccount(user);
  return new Response(null, { status: 204 });
});
