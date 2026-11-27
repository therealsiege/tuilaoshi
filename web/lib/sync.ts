import { HttpError } from "./http";
import type { ExportFiles } from "./store";

// A sync payload is one device's export folder (internal/wiki/backup.go,
// Store.Export) as {files: {relative path: file contents}}:
//   schedules.yaml, updated.yaml, reviews.csv, and every doc under its
//   name (journey.yaml, saves/<id>.yaml, ...). Nothing under private/ is
//   ever accepted: the wiki's private material (the frequency list, the
//   learner's own episodes) must never reach the server or another device.
// The Go side writes each device's files to root/<device>/<path> and runs
// Store.Sync(root, name), which reads only .yaml files and reviews.csv.

const SEGMENT = "[A-Za-z0-9_][A-Za-z0-9._-]{0,99}";
const YAML_PATH = new RegExp(`^(?:${SEGMENT}/){0,4}${SEGMENT}\\.yaml$`);
const MAX_FILES = 5000;
const MAX_PATH = 255;

export function isExportPath(p: string): boolean {
  return p.length <= MAX_PATH && !p.startsWith("private/") && (p === "reviews.csv" || YAML_PATH.test(p));
}

export function checkFiles(v: unknown): ExportFiles {
  if (typeof v !== "object" || v === null || Array.isArray(v)) {
    throw new HttpError(400, "files must be an object of relative path to file contents");
  }
  const entries = Object.entries(v);
  if (entries.length > MAX_FILES) throw new HttpError(400, `an export may hold at most ${MAX_FILES} files`);
  const files: ExportFiles = {};
  for (const [path, body] of entries) {
    if (!isExportPath(path)) {
      throw new HttpError(400, `bad path "${path.slice(0, 80)}": only relative .yaml paths and reviews.csv, no "..", "." or "/" prefixes`);
    }
    if (typeof body !== "string") throw new HttpError(400, `contents of "${path}" must be a string`);
    if (body.includes("\u0000")) throw new HttpError(400, `contents of "${path}" hold a NUL character`);
    files[path] = body;
  }
  return files;
}
