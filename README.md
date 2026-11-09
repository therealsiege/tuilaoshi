# mandarin-wiki

A terminal UI for Mandarin Chinese study, built for reactivating rusty university-level Mandarin. Combines a local markdown wiki of characters, vocabulary, grammar, and readings with spaced-repetition drills, Claude-powered answer evaluation, and text-to-speech.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Features

- **Dashboard** -- daily stats, streak tracking, review progress, wiki content overview
- **Drill modes** -- recognition (Chinese to English), production (English to Chinese), listening (TTS to Chinese), sentence completion, random mix
- **Spaced repetition** -- SM-2 inspired scheduler with difficulty grading (Again / Hard / Good / Easy)
- **Review queue** -- surfaces due items based on SRS intervals
- **Wiki browser** -- browse and read all character, vocabulary, grammar, and reading pages
- **Chat** -- conversational Mandarin tutor powered by Claude (styled message bubbles)
- **Text-to-speech** -- Chinese and English voices via Fish Audio, with a voice picker
- **Session logging** -- drill results saved to `wiki/sessions/` with per-session stats

## Wiki content

| Category | Pages | Coverage |
|----------|-------|----------|
| Characters | 43 | HSK1-2 individual character analysis |
| Vocabulary | 21 | Topic-grouped tables (family, food, travel, etc.) |
| Grammar | 21 | Pattern pages with examples and common errors |
| Readings | 5 | Graded passages HSK2-3 with vocab and comprehension |

All content is plain markdown with YAML frontmatter for SRS tracking. Add your own pages and they'll appear in the browser and drill system automatically.

## Requirements

- Go 1.21+
- An [Anthropic API key](https://console.anthropic.com/) for drill evaluation and chat
- (Optional) A [Fish Audio API key](https://fish.audio/) for TTS

## Setup

```bash
git clone git@github.com:therealsiege/mandarin-wiki.git
cd mandarin-wiki
cp .env.example .env   # then add your API keys
```

Create a `.env` file in the repo root:

```
ANTHROPIC_API_KEY=sk-ant-...
FISH_AUDIO_API_KEY=...          # optional, for TTS
```

## Build and run

```bash
make build    # builds to cmd/mandarin/mandarin
make run      # builds and runs
```

Or directly:

```bash
go build -o mandarin ./cmd/mandarin
./mandarin
```

The binary looks for `wiki/` relative to the executable or the current working directory. Override with `MANDARIN_WIKI_DIR=/path/to/wiki`.

## Keybindings

### Dashboard

| Key | Action |
|-----|--------|
| `d` | Start a drill |
| `b` | Browse wiki |
| `c` | Chat with tutor |
| `r` | Review due items |
| `v` | Change voice |
| `q` | Quit |
| `?` | Help overlay |

### Drill / Review

| Key | Action |
|-----|--------|
| `enter` | Submit answer / confirm |
| `1-4` | Rate difficulty (Again/Hard/Good/Easy) |
| `esc` | Back |

### Browser

| Key | Action |
|-----|--------|
| `up/down` | Navigate |
| `enter` | Open page |
| `/` | Filter |
| `esc` | Back |

## Project structure

```
cmd/mandarin/       Entry point
internal/
  app/              Root Bubble Tea model, view routing
  audio/            Fish Audio TTS client
  claude/           Anthropic SDK client for evaluation and chat
  model/            Shared types (Card, Page, Session, DrillMode)
  ui/               Styles, keys, view components
    browser/        Wiki page browser
    chat/           Tutor chat with message bubbles
    dashboard/      Stats dashboard
    drill/          Mode picker and drill flow
    review/         SRS review queue
  wiki/             Markdown parser, SRS scheduler, store
wiki/               Markdown content (characters, vocabulary, grammar, readings)
```

## Adding content

Wiki pages are markdown files with YAML frontmatter:

```yaml
---
type: character    # character | vocabulary | grammar | reading
level: HSK1
status: active
last_reviewed: 2026-11-09
tags: [verb, HSK1]
introduced: 2026-11-09
review_interval: 1
next_review: 2026-11-10
---
```

- **Characters**: one file per character in `wiki/characters/`, named by pinyin+tone number (e.g. `hao3.md`)
- **Vocabulary**: topic tables in `wiki/vocabulary/` with Word | Pinyin | Meaning | Example columns
- **Grammar**: pattern pages in `wiki/grammar/` with structure, examples, common errors
- **Readings**: graded passages in `wiki/readings/` with Chinese text, pinyin, translation, vocab table, and comprehension questions

New pages are automatically discovered by the drill and browser systems.

## License

Private repository.
