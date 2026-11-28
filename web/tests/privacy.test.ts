import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import { parseInline, parseMarkdown } from "@/lib/markdown";

const served = readFileSync(new URL("../content/privacy-policy.md", import.meta.url), "utf8");

describe("the privacy policy page", () => {
  it("serves the same text as the App Store copy", () => {
    const store = readFileSync(new URL("../../ios/AppStore/privacy-policy.md", import.meta.url), "utf8");
    expect(served).toBe(store);
  });

  it("reads the whole policy into headings, paragraphs and lists", () => {
    const blocks = parseMarkdown(served);
    expect(blocks[0]).toEqual({ kind: "h1", text: [{ kind: "text", text: "TuiLaoShi privacy policy" }] });
    const headings = blocks.filter((b) => b.kind === "h2").map((b) => (b.kind === "h2" ? b.text[0].text : ""));
    expect(headings).toContain("What the server keeps when you sign in");
    expect(headings).toContain("Contact");
    // No Markdown syntax is left over for the page to show raw.
    expect(JSON.stringify(blocks)).not.toMatch(/\*\*|^#|"- /);
  });
});

describe("parseMarkdown", () => {
  it("keeps paragraphs apart and joins wrapped lines", () => {
    expect(parseMarkdown("one\ntwo\n\nthree")).toEqual([
      { kind: "p", text: [{ kind: "text", text: "one two" }] },
      { kind: "p", text: [{ kind: "text", text: "three" }] },
    ]);
  });

  it("collects bullets, including a wrapped one", () => {
    expect(parseMarkdown("- a\n  more\n- b")).toEqual([
      { kind: "ul", items: [[{ kind: "text", text: "a more" }], [{ kind: "text", text: "b" }]] },
    ]);
  });

  it("finds bold and links, leaving the sentence's full stop outside the link", () => {
    expect(parseInline("**Yes.** See https://example.com/x.")).toEqual([
      { kind: "bold", text: "Yes." },
      { kind: "text", text: " See " },
      { kind: "link", text: "https://example.com/x" },
      { kind: "text", text: "." },
    ]);
  });
});
