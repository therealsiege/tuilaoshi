# Fight a demon from the dashboard. Run with: make recordings
Output docs/recordings/battle.gif
Set Shell zsh
Set FontSize 16
Set Width 1200
Set Height 800
Set TypingSpeed 60ms

# Play on a copy of the wiki so the recording doesn't change real review schedules.
Hide
Type 'D=$(mktemp -d) && go build -o $D/mandarin ./cmd/mandarin && cp -R wiki $D/ && clear && $D/mandarin -wiki $D/wiki'
Enter
Sleep 4s
Show

Sleep 2s
Type "f"
Sleep 2s
Enter
Sleep 2s
Type "1"
Sleep 2s
Enter
Sleep 1s
Type "2"
Sleep 2s
Enter
Sleep 1s
Type "3"
Sleep 3s
Escape
Sleep 1s
Hide
Type "q"
Show
