# ghboard demo — record with: vhs demo.tape
# Install VHS: https://github.com/charmbracelet/vhs

Output demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Dracula"

# Start the app (requires GITHUB_TOKEN to be set)
Type "GITHUB_TOKEN=$GITHUB_TOKEN ghboard"
Enter
Sleep 3s

# Heatmap tab — navigate around
Sleep 1s
Right
Right
Right
Sleep 500ms
Down
Down
Sleep 500ms
Left
Left
Sleep 500ms

# Switch to Stars tab
Type "2"
Sleep 2s

# Search for something
Type "/"
Sleep 300ms
Type "go"
Sleep 1s
Escape
Sleep 500ms

# Navigate list
Down
Down
Down
Sleep 500ms

# Switch to Notifications tab
Type "3"
Sleep 2s

# Navigate and mark one read
Down
Sleep 300ms
Type "r"
Sleep 500ms

# Show help
Type "?"
Sleep 2s

# Quit
Type "q"
