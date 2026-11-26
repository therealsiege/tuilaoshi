#!/bin/zsh
# make reminder AT=19:00: schedules TuiLaoShi's daily study reminder (a
# macOS notification from `tuilaoshi -remind`) with launchd, at AT each day.
# make reminder-off removes it. Uses the tuilaoshi command make launcher
# installs, on this repo's wiki.
set -eu
REPO="${0:A:h:h}"
LABEL=com.tuilaoshi.reminder
PLIST="$HOME/Library/LaunchAgents/$LABEL.plist"
if [[ "${1:-}" == "off" ]]; then
  launchctl bootout "gui/$(id -u)/$LABEL" 2>/dev/null || true
  rm -f "$PLIST"
  echo "The daily reminder is off."
  exit 0
fi
AT="${1:-19:00}"
if [[ ! "$AT" =~ '^[0-9]{1,2}:[0-9]{2}$' ]]; then
  echo "AT must be a time like 19:00" >&2; exit 1
fi
HOUR=$((10#${AT%%:*})); MINUTE=$((10#${AT##*:}))
BIN=$(command -v tuilaoshi || true)
[[ -n "$BIN" ]] || { echo "No tuilaoshi command: run make launcher first." >&2; exit 1; }
mkdir -p "$HOME/Library/LaunchAgents"
cat > "$PLIST" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>$LABEL</string>
  <key>ProgramArguments</key>
  <array><string>$BIN</string><string>-wiki</string><string>$REPO/wiki</string><string>-remind</string></array>
  <key>StartCalendarInterval</key>
  <dict><key>Hour</key><integer>$HOUR</integer><key>Minute</key><integer>$MINUTE</integer></dict>
</dict>
</plist>
PLIST
launchctl bootout "gui/$(id -u)/$LABEL" 2>/dev/null || true
launchctl bootstrap "gui/$(id -u)" "$PLIST"
echo "The daily reminder is set for $AT."
