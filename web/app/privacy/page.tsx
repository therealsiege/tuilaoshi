import { readFileSync } from "node:fs";
import { join } from "node:path";
import type { Metadata } from "next";
import { parseMarkdown, type Inline } from "@/lib/markdown";

// The privacy policy the App Store listing links to. Its text is
// content/privacy-policy.md, a copy of ios/AppStore/privacy-policy.md
// (tests/privacy.test.ts keeps the two the same), read at build time.

export const dynamic = "force-static";
export const metadata: Metadata = { title: "TuiLaoShi privacy policy" };

function Text({ parts }: { parts: Inline[] }) {
  return parts.map((p, i) =>
    p.kind === "bold" ? <strong key={i}>{p.text}</strong>
    : p.kind === "link" ? <a key={i} href={p.text}>{p.text}</a>
    : p.text,
  );
}

export default function Privacy() {
  const blocks = parseMarkdown(readFileSync(join(process.cwd(), "content", "privacy-policy.md"), "utf8"));
  return (
    <main style={{ maxWidth: "42rem", margin: "0 auto", padding: "1.5rem 1rem", lineHeight: 1.5, fontFamily: "system-ui, sans-serif" }}>
      {blocks.map((b, i) => {
        if (b.kind === "ul") return <ul key={i}>{b.items.map((item, j) => <li key={j}><Text parts={item} /></li>)}</ul>;
        const Tag = b.kind;
        return <Tag key={i}><Text parts={b.text} /></Tag>;
      })}
    </main>
  );
}
