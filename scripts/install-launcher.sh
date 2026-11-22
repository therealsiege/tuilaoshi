#!/bin/zsh
# Puts a self-contained TuiLaoShi.app on the Desktop, plus a `tuilaoshi`
# command on this Mac. The app carries the tutor (built for Apple silicon and
# Intel), a copy of the wiki, and the API keys from ~/.config/tuilaoshi/env,
# so it runs on any Mac that shares this Desktop, with nothing else installed.
# On this Mac it runs from the repo on the real wiki instead. Safe to rerun.
set -e
REPO="${0:A:h:h}"
HOST="$(/usr/sbin/scutil --get LocalHostName)"
APP="$HOME/Desktop/TuiLaoShi.app"
RES="$APP/Contents/Resources"
WORK=$(mktemp -d); trap 'rm -rf "$WORK"' EXIT

rm -rf "$APP"
osacompile -o "$APP" -e 'tell application "Terminal"
  activate
  do script "exec " & quoted form of (POSIX path of (path to me) & "Contents/Resources/start.sh")
end tell'

# The tutor, for both chip types; pure Go, so nothing to install to run it.
(cd "$REPO" &&
  CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$WORK/arm64" ./cmd/mandarin &&
  CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o "$WORK/amd64" ./cmd/mandarin)
lipo -create -output "$RES/tuilaoshi" "$WORK/arm64" "$WORK/amd64"

# A snapshot of the wiki (private episodes included) to seed other Macs.
rsync -a --delete --exclude '.DS_Store' "$REPO/wiki/" "$RES/wiki/"

# API keys, private to the user: the app lives in their iCloud Desktop, never the repo.
if [[ -f "$HOME/.config/tuilaoshi/env" ]]; then
  cp "$HOME/.config/tuilaoshi/env" "$RES/env"
  chmod 600 "$RES/env"
fi

cat > "$RES/start.sh" <<START
#!/bin/zsh
# Starts TuiLaoShi. On $HOST it runs from the repo on the real wiki; on any
# other Mac it runs the copy in this app, keeping that Mac's progress in
# ~/Library/Application Support/TuiLaoShi.
RES="\${0:A:h}"
if [[ "\$(/usr/sbin/scutil --get LocalHostName 2>/dev/null)" == '$HOST' && -x '$REPO/scripts/launch.sh' ]]; then
  exec '$REPO/scripts/launch.sh'
fi
DATA="\$HOME/Library/Application Support/TuiLaoShi"
mkdir -p "\$DATA/wiki"
# Add new pages and episodes from this app without overwriting progress.
/usr/bin/rsync -a --ignore-existing "\$RES/wiki/" "\$DATA/wiki/"
[[ -f "\$RES/env" ]] && source "\$RES/env"
export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
printf '\e[8;45;120t'   # ask the terminal for 120x45
exec "\$RES/tuilaoshi" -wiki "\$DATA/wiki"
START
chmod +x "$RES/start.sh"

# Icon: render assets/icon.svg with Quick Look, then build an .icns.
if qlmanage -t -s 1024 -o "$WORK" "$REPO/assets/icon.svg" >/dev/null 2>&1; then
  SET="$WORK/icon.iconset"; mkdir "$SET"
  for s in 16 32 128 256 512; do
    sips -z $s $s "$WORK/icon.svg.png" --out "$SET/icon_${s}x${s}.png" >/dev/null
    sips -z $((s*2)) $((s*2)) "$WORK/icon.svg.png" --out "$SET/icon_${s}x${s}@2x.png" >/dev/null
  done
  iconutil -c icns "$SET" -o "$RES/applet.icns"
  touch "$APP"
fi

# A command for terminal sessions on this Mac.
BIN=/opt/homebrew/bin
if [[ -w "$BIN" ]]; then
  ln -sf "$REPO/scripts/launch.sh" "$BIN/tuilaoshi"
  echo "Installed the tuilaoshi command in $BIN"
fi
echo "Installed $APP ($(du -sh "$APP" | cut -f1); runs from the repo on $HOST, self-contained elsewhere)"
