#!/bin/zsh
# Puts TuiLaoShi.app on the Desktop and a `tuilaoshi` command on the PATH.
# The app works from any Mac that shares this Desktop (iCloud): on this Mac
# it runs the tutor here; elsewhere it connects here over mosh (or ssh) and
# runs it, so the wiki, progress and keys stay in one place. Safe to rerun.
set -e
REPO="${0:A:h:h}"
HOST="$(scutil --get LocalHostName)"
# Reach this Mac by its Tailscale name when it has one (works away from home),
# else by its local network name.
REMOTE="$HOST.local"
TS=$(command -v tailscale || true)
if [[ -n "$TS" ]]; then
  REMOTE=$("$TS" status --json 2>/dev/null | python3 -c 'import json,sys; print(json.load(sys.stdin)["Self"]["DNSName"].rstrip("."))' 2>/dev/null || echo "$REMOTE")
fi
APP="$HOME/Desktop/TuiLaoShi.app"
WORK=$(mktemp -d); trap 'rm -rf "$WORK"' EXIT

rm -rf "$APP"
osacompile -o "$APP" -e 'tell application "Terminal"
  activate
  do script "exec " & quoted form of (POSIX path of (path to me) & "Contents/Resources/start.sh")
end tell'

cat > "$APP/Contents/Resources/start.sh" <<START
#!/bin/zsh
# Starts TuiLaoShi on $HOST: directly when this is $HOST, otherwise over mosh
# (or ssh) from another Mac.
REPO='$REPO'
if [[ "\$(/usr/sbin/scutil --get LocalHostName 2>/dev/null)" == '$HOST' ]]; then
  exec "\$REPO/scripts/launch.sh"
fi
export PATH="/opt/homebrew/bin:/usr/local/bin:\$PATH"
echo "Connecting to $HOST ($REMOTE)..."
if command -v mosh >/dev/null; then
  exec mosh '$REMOTE' -- "\$REPO/scripts/launch.sh"
fi
exec ssh -t '$REMOTE' "\$REPO/scripts/launch.sh"
START
chmod +x "$APP/Contents/Resources/start.sh"

# Icon: render assets/icon.svg with Quick Look, then build an .icns.
if qlmanage -t -s 1024 -o "$WORK" "$REPO/assets/icon.svg" >/dev/null 2>&1; then
  SET="$WORK/icon.iconset"; mkdir "$SET"
  for s in 16 32 128 256 512; do
    sips -z $s $s "$WORK/icon.svg.png" --out "$SET/icon_${s}x${s}.png" >/dev/null
    sips -z $((s*2)) $((s*2)) "$WORK/icon.svg.png" --out "$SET/icon_${s}x${s}@2x.png" >/dev/null
  done
  iconutil -c icns "$SET" -o "$APP/Contents/Resources/applet.icns"
  touch "$APP"
fi

# A command for terminal sessions on this Mac, including over mosh/ssh.
BIN=/opt/homebrew/bin
if [[ -w "$BIN" ]]; then
  ln -sf "$REPO/scripts/launch.sh" "$BIN/tuilaoshi"
  echo "Installed the tuilaoshi command in $BIN"
fi
echo "Installed $APP (runs on $HOST; other Macs connect to $REMOTE)"
