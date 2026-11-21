# Feature map

Every screen, how to reach it from the dashboard, and what tests it. A test
(`internal/ui/features_test.go`) fails if a dashboard key is missing here.

| Key | Screen | What it does | Code | Tests |
|-----|--------|--------------|------|-------|
| — | Dashboard | Wiki counts, items due, today's progress, cultivation rank | `internal/ui/dashboard` | — |
| `d` | Drill 练习 | Flashcards graded by Claude (exact typed Chinese is graded instantly, offline); modes: recognition, production, listening, sentence, random | `internal/ui/drill` | screen snapshots (mode picker, prompt) |
| `f` | Battle 战斗 | Turn-based fight against a demon from the chapters reached; `b` challenges an unlocked boss; listening questions when audio works (`r` replays); write questions typed with the pinyin input; order questions (rebuild an example sentence from shuffled words, 1-9 / backspace); results save to word schedules | `internal/ui/rpg`, `internal/rpg` | snapshots (intro, fighting), [state chart](battle-states.md) transitions, save/reschedule tests, one-right-answer property test (random decks + real wiki) |
| `b` | Browse 览 | Page list with rendered markdown | `internal/ui/browser` | — |
| `/` | Search 找 | Browse, filtered by search | `internal/ui/browser` | — |
| `c` | Chat 话 | Conversation with the Claude tutor; Chinese replies are spoken | `internal/ui/chat` | — |
| `r` | Review 复 | Due items only, graded with spaced repetition | `internal/ui/review` | screen snapshots (list, prompt) |
| `s` | Stats 统 | Daily/weekly activity, streak, review levels, characters mastered | `internal/ui/stats` | review-level bands, mastered count matches journey |
| `j` | Journey 游 | Journey to the West progress: rank, companions, bosses, artifacts, quests, achievements, lore, bestiary | `internal/ui/journey`, `internal/journey` | mastery rule |
| `t` | Reading 读 | Read texts from `wiki/readings` word by word with pinyin and meanings; enter marks a word you didn't know so it comes back tomorrow | `internal/ui/reader` | segmentation, glosses, mark-for-review, screen snapshots (list, text) |
| `a` | Assess 测 | Adaptive diagnostic per skill (characters, vocabulary, grammar sentences, reading vocabulary): starts at HSK3, 3 questions per level, writes the levels and priority areas to `wiki/overview.md` | `internal/ui/assessment`, `internal/assess` | staircase, overview rewrite, full run on a test wiki, intro snapshot |
| `x` | Radicals 部 | Radical families explorer | `internal/ui/radicals` | — |
| `v` | Voice 声 | Choose Chinese and English text-to-speech voices | `internal/app` | — |
| `q` | Quit 退 | Leave the app (from the dashboard) | `internal/app` | — |

Shared pieces:

- Saving progress: `wiki.Store.RecordReview` is the only path that writes review schedules; session files go through `internal/wiki/sessions.go`. Tests in `internal/wiki`.
- Speech: `ui.SpeakCmd` (Fish Audio with `FISH_AUDIO_API_KEY`, otherwise macOS `say -v Tingting`); `ctrl+r` replays in drill and review; failures show as "Audio failed".
- Pinyin input: `internal/ui/input`; syllable validation tests.

Recordings (`make recordings`, needs `brew install vhs`; tapes in `docs/tapes`, played on a copy of the wiki):
[battle](recordings/battle.gif), [journey](recordings/journey.gif).

Journey content (tribulations, demons, bosses, companions, lore, dialogue) is data in `internal/journey/content/*.yaml`, embedded in the binary; `internal/journey/content_test.go` validates it as a whole, so adding a demon or lore passage is a YAML-only change.
