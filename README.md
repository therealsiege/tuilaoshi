<p align="center">
  <h1 align="center">推老师 TuīLǎoShī</h1>
  <p align="center"><strong>Your Mandarin tutor lives in the terminal.</strong></p>
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white" alt="Go 1.22+"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="MIT License"></a>
  <a href="https://github.com/therealsiege/tuilaoshi/stargazers"><img src="https://img.shields.io/github/stars/therealsiege/tuilaoshi?style=flat" alt="GitHub Stars"></a>
</p>

---

No apps. No accounts. No distractions. Just you, your terminal, and 5,300+ Chinese characters waiting to be learned.

TuīLǎoShī is a terminal-native Mandarin tutor that combines spaced-repetition drilling, AI-powered conversation, text-to-speech, and a Journey to the West-themed progression system — all in a single Go binary you can run anywhere.

<!-- ## Screenshots -->

<!-- > _Coming soon._ -->

## What You Get

### An AI tutor that actually teaches

Chat with Claude in Chinese or English. Ask questions about grammar, get corrections on your sentences, or just have a conversation. Every response is spoken aloud with natural Chinese TTS. It's like having a patient language partner available 24/7 — minus the small talk.

### Five ways to drill

- **Recognition** — See the character, type the meaning
- **Production** — See the English, produce the Chinese
- **Listening** — Hear the word spoken, type what you heard
- **Sentence** — Fill in the missing word from context
- **Random Mix** — All modes shuffled to keep you sharp

Each answer is evaluated by Claude with structured feedback — not just "right" or "wrong," but *why* and *how to remember it*.

### Spaced repetition that sticks

SM-2-inspired scheduling tracks every card's interval, ease factor, and next review date. Rate each answer Again / Hard / Good / Easy, and the algorithm adapts. Your SRS state lives right in the content files — no opaque database, no sync issues, no lock-in.

### 5,300+ characters out of the box

Ships with pre-built wiki pages covering the full HSK1–6 curriculum plus specialized domain packs:

| Pack | Coverage |
|------|----------|
| **HSK 1–6** | Core curriculum from beginner to advanced |
| **Business** | Corporate, finance, marketing, accounting |
| **Technology** | Computing, AI, cybersecurity, robotics |
| **Healthcare** | Anatomy, TCM, pharmacy, mental health |
| **Sports** | Martial arts, athletics, fitness, competition |

Every character page includes pinyin, meaning, radical, stroke count, compound words, and example sentences.

### Journey to the West

Gamification that actually fits. Your study progress maps onto the 81 tribulations of 西游记:

**Cultivation Ranks** — Rise from 凡人 (Mortal) through 练气 (Qi Refining), 筑基 (Foundation Building), 金丹 (Golden Core), all the way to 斗战胜佛 (Victorious Fighting Buddha).

**Companions** — Unlock the classic party as you hit milestones: 唐僧 Tang Sanzang guides your journey from day one. 孙悟空 Sun Wukong joins at 50 characters. 猪八戒 Zhu Bajie, 沙悟净 Sha Wujing, and 白龙马 Bai Longma follow as you grow.

**Boss Encounters** — Face demon kings from the novel: 混世魔王 Havoc Monster, 白骨精 White Bone Demon, 红孩儿 Red Boy, all the way up to 如来佛祖 Buddha himself at 4,000 characters.

**Sacred Artifacts** — Earn 金箍棒 Ruyi Jingu Bang, 芭蕉扇 Banana Leaf Fan, 火眼金睛 Fiery Eyes, and more through mastery and study streaks.

### Pinyin input built in

Three input modes toggled with `ctrl+space`:

- **Normal** — Standard English typing
- **Pinyin** — Type `ni3 hao3`, get `nǐ hǎo` automatically
- **Chinese** — Full pinyin composition with dictionary candidate selection

### Text-to-speech

Chinese and English voices via Fish Audio. Cards are spoken aloud during drills. Pick your preferred voice from multiple options per language.

### Everything is a file

Your entire curriculum is plain markdown with YAML frontmatter. Browse it in the TUI, edit it in vim, version it with git, back it up however you want. No database. No migrations. No vendor lock-in.

```
wiki/
├── characters/    5,300+ character pages
├── vocabulary/    Topic-grouped word tables
├── grammar/       Pattern pages with examples
├── readings/      Graded passages with annotations
└── reference/     Quick-reference sheets
```

## Get Started

### Install

```bash
git clone https://github.com/therealsiege/tuilaoshi.git
cd tuilaoshi
go build -o tuilaoshi ./cmd/mandarin
./tuilaoshi
```

Or with `go install`:

```bash
go install github.com/therealsiege/tuilaoshi/cmd/mandarin@latest
```

### Configure

```bash
# Required — powers drill evaluation and AI chat
export ANTHROPIC_API_KEY=sk-ant-...

# Optional — enables text-to-speech
export FISH_AUDIO_API_KEY=...
```

### Navigate

| Key | What it does |
|-----|--------------|
| `d` | Start a drill |
| `c` | Chat with AI tutor |
| `r` | Review due cards |
| `b` | Browse the wiki |
| `j` | Journey to the West |
| `s` | Study statistics |
| `v` | Change TTS voice |
| `/` | Search |
| `q` | Quit |

## Why a Terminal?

Language learning apps are full of notifications, upsells, and gamification dark patterns designed to maximize engagement metrics rather than actual learning.

TuīLǎoShī has none of that.

The terminal is where you already work. There's no context switch, no app store, no account to create. The interaction is fundamentally text in, text out — which turns out to be exactly what character drilling needs. You type an answer, you get feedback. No tap targets, no drag-and-drop, no animations eating your battery.

Your content is yours. Plain markdown files you can read, edit, grep, and git push. The SRS state lives in the same files as the content. Everything is local, transparent, and under your control.

## Under the Hood

Single-binary Go application. No external runtime dependencies.

| Layer | Tech |
|-------|------|
| TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) + [Glamour](https://github.com/charmbracelet/glamour) |
| AI | [Anthropic SDK for Go](https://github.com/anthropics/anthropic-sdk-go) (Claude) |
| SRS | SM-2 algorithm with four difficulty grades |
| TTS | [Fish Audio](https://fish.audio) streaming API |
| Content | Markdown + YAML frontmatter, parsed at startup |

```
cmd/mandarin/         Entry point
cmd/generate/         Character page generator
internal/
  app/                Root model, view routing
  audio/              Fish Audio TTS client
  claude/             Anthropic API client
  journey/            Journey to the West engine
  model/              Shared types
  ui/
    browser/          Wiki page browser
    chat/             AI conversation view
    dashboard/        Home screen
    drill/            Drill mode picker and flow
    input/            Pinyin/Chinese input method
    journey/          Journey to the West view
    review/           SRS review queue
    stats/            Statistics and charts
  wiki/               Markdown parser, SRS scheduler, store
```

## Contributing

Contributions welcome. Some ideas:

- **Content** — character, vocabulary, grammar, or reading pages at any level
- **Drill modes** — tone drills, radical decomposition, handwriting recognition
- **Platform** — testing and fixes for Linux and Windows
- **Tests** — unit and integration coverage

Open an issue before starting significant work.

## License

[MIT](LICENSE)
