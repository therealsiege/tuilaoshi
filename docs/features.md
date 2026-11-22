# Feature map

Every screen, how to reach it from the dashboard, and what tests it. A test
(`internal/ui/features_test.go`) fails if a dashboard key is missing here.

| Key | Screen | What it does | Code | Tests |
|-----|--------|--------------|------|-------|
| — | Dashboard | Today's numbers, the purse (铜钱 copper cash: one per word answered right, once a day per word, a daily bonus at 20 answers plus the streak, battle wins; 元宝 silver ingots from bosses; saved in `wiki/journey.yaml`), one suggested next step, the menu grouped into Study / Journey / More (`ui.DashboardMenu`, also the help screen), and where the journey stands | `internal/ui/dashboard` | screen snapshots (wide, narrow), next-step choice, menu matches key bindings |
| `d` | Drill 练习 | Flashcards graded by Claude (exact typed Chinese is graded instantly, offline); modes: recognition, production, listening, sentence, random | `internal/ui/drill` | screen snapshots (mode picker, prompt) |
| `f` | Battle 战斗 | Turn-based fight: a stage's story scene the first time it opens (read word by word, tab to fight), then a gauntlet: one or two of the stage's foot soldiers (小妖: minion, elite, captain; `content/soldiers.yaml`), HP carrying between fights with a short rest, then the open tribulation's demon or boss with that stage's themed words (beating it clears the stage), or practice on the coming stage's words, mostly against soldiers of the road so far, while the next stage still needs mastered characters; wins pay 铜钱 by rank (a practice win only when it recalled a word not yet paid today) and a boss pays a 元宝; listening questions when audio works (`r` replays); write questions typed with the pinyin input; order questions (rebuild an example sentence from shuffled words, 1-9 / backspace); demon abilities (blind, disguise, burn, venom) and companion assists (w reveal, shield, first strike, half combo); results save to word schedules | `internal/ui/rpg`, `internal/rpg` | snapshots (intro, fighting), [state chart](battle-states.md) transitions, gauntlet and reward tests, save/reschedule tests, one-right-answer property test (random decks + real wiki) |
| `g` | Shrine 土地庙 | The Earth God's shop and a character sheet: spend 铜钱 and 元宝 on gear for four slots (weapon: damage per hit; robe: max HP; hat: longer burn timer, half combo; charm: shield, clear sight against blinding, an extra reveal) and titles; items stock the shelf as stages are cleared (`content/gear.yaml`); enter buys and puts on, or takes off; gear applies in every battle and never touches review schedules | `internal/ui/shrine`, `internal/rpg/abilities.go`, `internal/wiki/gear.go` | screen snapshot, buy / wear / take off / can't afford / not yet stocked, gear effects against a blinding burning demon, purse arithmetic |
| `b` | Browse 览 | Every wiki page, sorted readings → grammar → vocabulary → characters by HSK level, with the selected page rendered beside the list | `internal/ui/browser` | fits the terminal, cursor stays visible, sort order, search, truncation, screen snapshot |
| `/` | Search 找 | Browse with the search box open: titles, pinyin (tones optional), meanings and tags; every word must match; whole-word matches first | `internal/ui/browser` | pinyin without tones, j/k type in search, ranking, esc clears |
| `c` | Chat 话 | Conversation with the Claude tutor; Chinese replies are spoken | `internal/ui/chat` | — |
| `r` | Review 复 | Due items only, graded with spaced repetition | `internal/ui/review` | screen snapshots (list, prompt) |
| `s` | Stats 统 | Daily/weekly activity, streak, review levels, characters mastered | `internal/ui/stats` | review-level bands, mastered count matches journey |
| `j` | Journey 游 | Journey to the West progress: rank, companions, bosses, artifacts, quests, achievements, lore, bestiary | `internal/ui/journey`, `internal/journey` | mastery rule |
| `t` | Reading 读 | Read word by word with pinyin and meanings: texts from `wiki/readings`, journey dialogue episodes (original and adapted from the novel) as their stage is reached, and your own episodes in the git-ignored `wiki/private/episodes` (a subtitle file, or a folder with Chinese subtitles, English subtitles matched by timing, and the original audio, where p plays the current line); e shows a line's English, enter marks a word you didn't know so it comes back tomorrow | `internal/ui/reader` | segmentation, glosses, mark-for-review, episodes, screen snapshots (list, text) |
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

Journey content (tribulations, demons, bosses, companions, lore, dialogue, vocabulary themes, story scenes, dialogue episodes, demon and companion abilities) is data in `internal/journey/content/*.yaml`, embedded in the binary; `internal/journey/content_test.go` validates it as a whole, so adding a demon or lore passage is a YAML-only change.

Launching: `make launcher` puts a self-contained TuiLaoShi.app on the Desktop (the tutor built for Apple silicon and Intel, a wiki snapshot, and the API keys from `~/.config/tuilaoshi/env`) and a `tuilaoshi` command in /opt/homebrew/bin. On this Mac the app runs from the repo on the real wiki; on another Mac sharing the Desktop it runs on its own, keeping that Mac's progress in `~/Library/Application Support/TuiLaoShi`.
