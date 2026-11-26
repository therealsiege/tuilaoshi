#!/bin/zsh
# make sweep: opens every screen of the real game at 80x24 and 120x45 on a
# throwaway copy of the wiki (offline, no speech) and fails when a screen's
# title isn't on top or a line is cut at the window's edge.
set -u
REPO="${0:A:h:h}"
D=$(mktemp -d)
S="tlsweep$$"
trap 'tmux kill-session -t $S 2>/dev/null; rm -rf "$D"' EXIT
cd "$REPO" || exit 1
go build -o "$D/t" ./cmd/mandarin || exit 1
rsync -a --exclude 'progress.db*' --exclude backups --exclude 'private/vocabulary' wiki/ "$D/wiki/"

# A journey at 宝象国's gate (tribulation 22), with food for the walk.
mkdir -p "$D/export/saves"
{
  echo "name: sweep"; echo "created: \"2026-11-20T09:00:00Z\""; echo "played: \"2026-11-20T09:00:00Z\""
  echo "cleared:"; for i in $(seq 1 21); do echo "    $i: \"2026-11-20\""; done
  echo "coins:"; echo "    tongqian: 500"
  echo "region: baihu"; echo "x: 23"; echo "y: 4"
  echo "trail:"; echo "    day: 1"; echo "    food: 5000"; echo "    water: 30"; echo "    pace: steady"; echo "    rations: filling"
} > "$D/export/saves/sweep.yaml"
printf 'speech: false\neffects: false\n' > "$D/export/settings.yaml"

# name|keys after continuing|title on top|1 if lines may fill the width
screens=(
  "map||region|0"
  "camp|c|营地|0"
  "flashcards|c d|Drill|0"
  "review|c r|Review Queue|0"
  "reading|c t|Reading|0"
  "storyteller|c t s|Reading|0"
  "episodes|c e|剧场|0"
  "forge|c z|字炉|0"
  "chat|c c|Chat|1"
  "spar|c f|BATTLE|0"
  "demon|c m|营地|0"
  "diary|c w|日记|0"
  "letter|c n|营地|0"
  "journey|c j|西游记|0"
  "shrine|c g|土地庙|0"
  "collection|c o|图鉴|0"
  "party|c p|师徒|0"
  "progress|c s|学习统计|0"
  "web|c s w|盘丝洞|0"
  "levelcheck|c a|Assess|0"
  "browse|c b|Wiki Browser|1"
  "festivals|c y|节日|0"
  "radicals|c x|Radicals|0"
  "voices|c v|Voice|0"
  "settings|c ,|设置|0"
  "town|Left|宝象国|0"
  "riddle|Left l|灯谜|0"
  "gate|Left p|通关文牒|0"
  "haggle|Left h|宝象国|0"
  "rumour|Left f|宝象国|0"
)
failed=0
for size in 80x24 120x45; do
  W=${size%x*}; H=${size#*x}
  echo "$size"
  for entry in $screens; do
    IFS='|' read -r name keys title fullok <<< "$entry"
    "$D/t" -wiki "$D/wiki" -import "$D/export" >/dev/null
    tmux kill-session -t $S 2>/dev/null
    tmux new-session -d -s $S -x $W -y $H "cd '$REPO' && env -u ANTHROPIC_API_KEY '$D/t' -wiki '$D/wiki'"
    sleep 1.5; tmux send-keys -t $S Enter; sleep 1.5
    for k in ${=keys}; do tmux send-keys -t $S "$k"; sleep 1.2; done
    sleep 0.5
    tmux capture-pane -pt $S | python3 "$REPO/scripts/sweep_check.py" $W "$name" "$title" "$fullok" || failed=1
  done
done
exit $failed
