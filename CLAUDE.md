# Mandarin Reactivation Wiki — Schema

## Learner Profile

- **Background**: 4 years university Mandarin (reading, writing, speaking, listening)
- **Characters**: Simplified (简体字)
- **Current state**: Rusty — can read characters and recognize grammar patterns, but retrieval is slow and vocabulary has shrunk
- **Goal**: Full literacy — reading, writing, and conversation back to university level and beyond

## Directory Structure

```
mandarin-wiki/
├── CLAUDE.md          # this schema file
├── raw/               # immutable source materials (articles, texts, exercises)
│   └── assets/        # images, audio references, screenshots
├── wiki/
│   ├── index.md       # content catalog organized by category
│   ├── log.md         # chronological session record
│   ├── overview.md    # current state assessment and learning roadmap
│   ├── grammar/       # grammar pattern pages
│   ├── vocabulary/    # vocabulary pages by topic and level
│   ├── characters/    # character analysis (radicals, components, stroke order)
│   ├── readings/      # texts with annotations and comprehension notes
│   ├── sessions/      # individual tutoring session records
│   └── reference/     # quick-reference pages (measure words, particles, etc.)
└── .gitignore
```

## Page Conventions

- All wiki pages are markdown
- Chinese text always appears with pinyin on first use: 你好 (nǐ hǎo)
- Character pages include: character, pinyin, meaning, radical, stroke count, common compounds, example sentences
- Grammar pages include: pattern, structure, examples (Chinese + pinyin + English), common errors, related patterns
- Vocabulary pages group words by topic with: word, pinyin, meaning, example sentence, usage notes
- Session pages record: date, topics covered, new items introduced, items reviewed, weak spots identified, homework

## Frontmatter

Wiki pages use YAML frontmatter for queryability:

```yaml
---
type: grammar | vocabulary | character | reading | session | reference
level: HSK1 | HSK2 | HSK3 | HSK4 | HSK5 | HSK6 | advanced
status: draft | active | review-needed
last_reviewed: 2026-11-07
tags: [topic tags]
---
```

## Operations

### Session (primary workflow)
A tutoring session. Flow:
1. Review weak spots from previous sessions
2. Introduce or reactivate target material (grammar, vocabulary, characters)
3. Practice through exercises, translation, or conversation
4. Assess comprehension
5. Log session to `wiki/sessions/` and update `log.md`
6. Update relevant grammar/vocabulary/character pages
7. Note items for spaced review

### Ingest
Process new source material (article, text, exercise set). Flow:
1. Read the source in `raw/`
2. Extract vocabulary, grammar patterns, and cultural notes
3. Create/update wiki pages for new items
4. Write a summary page in `wiki/readings/`
5. Update `index.md` and `log.md`

### Query
Answer questions about Mandarin grammar, usage, or culture. Flow:
1. Check `index.md` for relevant wiki pages
2. Read relevant pages
3. Synthesize answer with examples
4. If the answer reveals a gap, create a new wiki page
5. File valuable explorations back into the wiki

### Drill
Targeted practice on specific items. Flow:
1. Select items based on review schedule or weak spots
2. Present exercises (character recognition, fill-in-the-blank, translation, sentence construction)
3. Evaluate responses
4. Update review status on relevant pages
5. Log drill results

### Assess
Diagnostic evaluation. Flow:
1. Test across categories: characters, vocabulary, grammar, reading comprehension
2. Calibrate current level per category
3. Update `overview.md` with assessment results
4. Identify priority areas for upcoming sessions

### Lint
Wiki health check. Look for:
- Grammar pages missing example sentences
- Vocabulary items without usage context
- Characters referenced but lacking their own page
- Orphan pages with no cross-references
- Items not reviewed in 30+ days
- Contradictions or outdated notes

## Spaced Review Tracking

Items enter the review cycle when introduced. Track in frontmatter:
- `introduced`: date first covered
- `last_reviewed`: date of most recent review
- `review_interval`: current interval in days (starts at 1, doubles on success, resets on failure)
- `next_review`: computed from last_reviewed + review_interval

Priority review items surface at the start of each session.

## Session Conventions

- Mix reactivation (old material surfacing) with new material
- Favor recognition before production: reading before writing, listening before speaking
- Use context-rich examples, not isolated words
- Connect new items to known items via shared radicals, grammar patterns, or semantic fields
- Progressively increase Chinese-only interaction as proficiency returns
- Include at least one "use it in a sentence" exercise per new item

## Tone Marking

Use numbered pinyin in filenames and data (e.g., ni3 hao3) and diacritical pinyin in display text (e.g., nǐ hǎo).

## Level Calibration

Given 4 years of university study, likely peak was around HSK4-5. Current rusty state is probably functional at HSK3 with passive recognition up to HSK5. Sessions should probe this boundary and work the reactivation frontier.
