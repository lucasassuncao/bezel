# shell: two tabs, two panes, a status that expires and a busy spinner.
#
# Record from the repository root: vhs examples/shell/shell.tape

Output examples/shell/shell.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      14
Set Width         1200
Set Height        700
Set Theme         "Catppuccin Mocha"
Set Framerate     30
Set PlaybackSpeed 1.0
Set TypingSpeed   45ms
Set Padding       20

# -- Build, off camera ----------------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o shell.exe ./examples/shell"
Enter
Wait
Type "clear"
Enter
Show

Type ".\shell.exe"
Sleep 400ms
Enter
Wait+Screen /main.go/
Sleep 1.5s

Down@300ms 2
Sleep 800ms

# -- A status that clears itself ----------------------------------------------------

Type "s"
Wait+Screen /saved/
Sleep 2.5s

# -- Busy, then idle -------------------------------------------------------------------

Type "b"
Wait+Screen /indexing/
Sleep 2.5s

# -- Tabs and panes --------------------------------------------------------------------

Tab
Wait+Screen /ideas/
Sleep 1.5s
Shift+Tab
Sleep 1.2s

# -- Every key, one keystroke away ---------------------------------------------------

Type "?"
Sleep 2.5s
Escape
Sleep 1s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item shell.exe"
Enter
Wait
