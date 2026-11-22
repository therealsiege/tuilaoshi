#!/bin/zsh
# Starts TuiLaoShi on this repo's wiki: rebuilds it if the code changed, then
# runs it full-window. The desktop launcher (make launcher) runs this.
REPO="${0:A:h:h}"
export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/local/go/bin:$PATH"

# Optional API keys (ANTHROPIC_API_KEY for chat and graded drills,
# FISH_AUDIO_API_KEY for premium voices), kept out of the repo.
[[ -f "$HOME/.config/tuilaoshi/env" ]] && source "$HOME/.config/tuilaoshi/env"

printf '\e[8;45;120t'   # ask the terminal for 120x45
cd "$REPO" || exit 1
if ! go build -o "$REPO/tuilaoshi" ./cmd/mandarin; then
  if [[ -x "$REPO/tuilaoshi" ]]; then
    echo "Build failed; starting the last working version."; sleep 2
  else
    echo "Build failed and there is no earlier version to run."; read -k1 "?Press any key to close."; exit 1
  fi
fi
exec "$REPO/tuilaoshi" -wiki "$REPO/wiki"
