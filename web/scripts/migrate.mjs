// Runs every db/*.sql file in name order against DATABASE_URL. Each file
// must be safe to run again (CREATE ... IF NOT EXISTS and the like).
import { readdir, readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import pg from "pg";

const url = process.env.DATABASE_URL;
if (!url) {
  console.error("DATABASE_URL is not set");
  process.exit(1);
}
const dir = fileURLToPath(new URL("../db/", import.meta.url));
const files = (await readdir(dir)).filter((f) => f.endsWith(".sql")).sort();
const client = new pg.Client({ connectionString: url });
await client.connect();
try {
  for (const f of files) {
    await client.query(await readFile(dir + f, "utf8"));
    console.log("applied", f);
  }
} finally {
  await client.end();
}
