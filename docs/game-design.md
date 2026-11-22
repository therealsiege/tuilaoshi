# Game design: RPG loops that make you study

TuiLaoShi already has the bones of an RPG: a rank that rises with mastered
characters, 81 stages, a gauntlet of soldiers before each demon, companions
who join and assist, a purse of 铜钱 and 元宝, achievements, and story scenes.
This note looks at what makes well-loved RPGs satisfying, maps each loop onto
learning Chinese, and picks what to build next.

**The rule for every mechanic: rewards follow real learning (a due review
passed, a word recalled from memory, a character mastered, a text read),
never repetition of what is already easy.** A game loop that can be farmed
teaches farming.

## Prior art, loop by loop

| Loop | Who does it well | What makes it work | In TuiLaoShi | Learning effect |
|---|---|---|---|---|
| Leveling and stat curves | Dragon Quest, Final Fantasy | Visible numbers rise on a curve; each level brings something new | Level = mastered characters (the existing rank), not XP; each level-up names what it unlocks (a region, a shop shelf, a companion skill) | Mastery, which needs spaced recall, is the only way up |
| Levels that gate content | WaniKani | Radicals → kanji → vocab per level; you can't rush | Borrow: stages open on characters learned or mastered (already). Avoid: WaniKani's hard lockouts and its punishing review pile after a break | Keeps new material at the reactivation frontier |
| Overworld and travel | Pokémon routes and towns | A map you move across; towns are rests between routes | The 81 stages as a walkable road: towns (宝象国, 车迟国, 女儿国…) hold the shrine, an inn, notice board and NPCs; routes between hold the gauntlets | Each place carries a vocabulary theme, so words get a location (memory palace) |
| Loot and gear that change play | Diablo, Monster Hunter | Drops that alter how you fight; crafting from what you hunted | Shrine shop (#26): gear changes battle rules (another reveal, more HP, a longer burn timer), never the review schedule. Monster Hunter-style: soldiers drop materials (妖丹, 虎皮) only on hard-word wins, crafted into gear | Harder recall pays better loot |
| Companion bonds | Persona social links, Fire Emblem supports | Time spent with a character deepens the bond and unlocks scenes and skills | 悟空, 八戒, 沙僧, 白龙马 each have bond ranks raised by talking to them in Chinese (chat mode, Claude-graded) and by battles they assist; each rank unlocks a short story scene and upgrades their assist | Conversation practice with a reason to come back |
| Side quests and NPCs | Witcher 3, Dragon Quest | Small stories with a request and a reward | Notice-board requests in towns: read a letter, write a reply, translate a sign, answer a villager; Claude grades, rewards coins and loot | Reading and writing in context, not drills |
| Collections | Pokédex | Seeing gaps makes you want to fill them | Enemy cards (#27) plus a character collection: every mastered character is a card, grouped by radical family; complete a family for a 元宝 and a title | Radical families are how characters are remembered |
| Daily rhythm | Animal Crossing, Duolingo | A few small things to do each day; the world changes with time | Three daily bounties from the notice board (clear today's due reviews, write two sentences, read a passage); the daily bonus already exists. Borrow Duolingo's streak; avoid its anxiety: a streak shield bought with 铜钱, no lost progress | Daily spaced practice is what makes SRS work |
| Adaptive difficulty | Resident Evil 4, Left 4 Dead director | The game quietly tunes itself to keep you in flow | Recent accuracy shifts the question mix: under 70% leans on recognition and hints, over 90% leans on writing and listening and adds an elite to the gauntlet | Keeps recall effortful but winnable (desirable difficulty) |
| Secrets | Zelda, Metroid | Hidden things reward curiosity | Hidden encounters for a mastered radical family; 成语 stones on the map that open with the idiom typed; a "nightmare" rematch of beaten bosses on your weakest words | Rewards depth, not just breadth |
| A meaningful economy | Stardew Valley, Monster Hunter | Coins are scarce enough to choose with | Copper buys gear and bounty rerolls, ingots buy titles, companion gifts and rare gear; nothing buys skipping a review | Spending is a reason to earn, earning means recalling |

## What to avoid

- **Grinding.** Rewards for re-answering easy words. Today a practice battle
  pays a copper coin for every right answer, so it can be farmed. Fix: copper
  only for a card's first right answer of the day, and more for due or shaky
  cards.
- **Pay-to-skip learning.** No item reschedules a review or opens a stage.
- **Loss aversion.** No lost coins or progress for missed days; streaks can be
  shielded.
- **Busywork menus.** Every new screen must lead to a question, a text or a
  conversation within one keypress.

## Shortlist, by impact on learning

1. **Fair pay** (small): copper per card per day, extra for due and shaky
   cards; harder wins drop loot. Stops grinding before the economy grows.
2. **Shrine shop and gear** (#26): something worth earning for. Gear changes
   battle rules; titles and looks on a character sheet.
3. **Levels with level-up moments**: level from mastered characters, a
   level-up screen that says what opened, stats from level.
4. **Collections** (#27, extended): enemy cards plus character cards by
   radical family; region and family set rewards.
5. **Daily bounties**: three small tasks a day from the notice board, mixing
   review, writing and reading.
6. **Adaptive difficulty**: question mix and gauntlet size from recent
   accuracy.
7. **Companion bonds**: chat with a companion in Chinese to raise the bond,
   which upgrades their assist and opens a scene.
8. **Overworld map and towns**: the road as a map with towns, inns, notice
   boards and side-quest NPCs.
9. **Side quests**: town requests to read, write and translate, Claude-graded.
10. **Secrets**: 成语 stones, hidden radical-family encounters, nightmare
    rematches.

Items 1-6 build on what exists; 7-10 are larger and need content.
