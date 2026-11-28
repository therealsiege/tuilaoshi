// A small reader for the privacy policy's Markdown (content/privacy-policy.md):
// headings, paragraphs, bullet lists, **bold** and bare https:// links. That is
// all the policy uses, so the page needs no Markdown library.

export type Inline = { kind: "text" | "bold" | "link"; text: string };

export type Block =
  | { kind: "h1" | "h2" | "p"; text: Inline[] }
  | { kind: "ul"; items: Inline[][] };

const INLINE = /\*\*(.+?)\*\*|(https:\/\/[^\s)]+[^\s).,;:])/g;

export function parseInline(s: string): Inline[] {
  const out: Inline[] = [];
  let last = 0;
  for (const m of s.matchAll(INLINE)) {
    if (m.index > last) out.push({ kind: "text", text: s.slice(last, m.index) });
    out.push(m[1] !== undefined ? { kind: "bold", text: m[1] } : { kind: "link", text: m[2] });
    last = m.index + m[0].length;
  }
  if (last < s.length) out.push({ kind: "text", text: s.slice(last) });
  return out;
}

export function parseMarkdown(md: string): Block[] {
  const blocks: Block[] = [];
  let para: string[] = [];
  let list: string[] | null = null;
  const flush = () => {
    if (para.length) blocks.push({ kind: "p", text: parseInline(para.join(" ")) });
    if (list) blocks.push({ kind: "ul", items: list.map(parseInline) });
    para = [];
    list = null;
  };
  for (const raw of md.split("\n")) {
    const line = raw.trim();
    if (!line) {
      flush();
    } else if (line.startsWith("## ") || line.startsWith("# ")) {
      flush();
      const h1 = line.startsWith("# ");
      blocks.push({ kind: h1 ? "h1" : "h2", text: parseInline(line.slice(h1 ? 2 : 3)) });
    } else if (line.startsWith("- ")) {
      if (para.length) flush();
      (list ??= []).push(line.slice(2));
    } else if (list) {
      list[list.length - 1] += " " + line; // a wrapped bullet
    } else {
      para.push(line);
    }
  }
  flush();
  return blocks;
}
