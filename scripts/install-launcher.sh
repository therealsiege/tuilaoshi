#!/bin/zsh
# Puts TuiLaoShi.app on the Desktop: double-click it to open the tutor in a
# Terminal window. Safe to rerun; it replaces the app.
set -e
REPO="${0:A:h:h}"
APP="$HOME/Desktop/TuiLaoShi.app"
WORK=$(mktemp -d); trap 'rm -rf "$WORK"' EXIT

rm -rf "$APP"
osacompile -o "$APP" -e "tell application \"Terminal\"
  activate
  do script \"exec '$REPO/scripts/launch.sh'\"
end tell"

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
echo "Installed $APP"
