# Tour the Journey to the West screen. Run with: make recordings
Output docs/recordings/journey.gif
Set Shell zsh
Set FontSize 16
Set Width 1200
Set Height 800

Hide
Type 'D=$(mktemp -d) && go build -o $D/mandarin ./cmd/mandarin && cp -R wiki $D/ && clear && $D/mandarin -wiki $D/wiki'
Enter
Sleep 4s
Show

Sleep 2s
Type "j"
Sleep 2s
Tab
Sleep 1500ms
Tab
Sleep 1500ms
Tab
Sleep 1500ms
Escape
Sleep 1s
Hide
Type "q"
Show
