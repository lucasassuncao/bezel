# overlay: the ready-made modals: alert, confirm, prompt, help and pager.
#
# Record from the repository root: vhs examples/overlay/overlay.tape

Output examples/overlay/overlay.gif

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
Type "go build -o overlay.exe ./examples/overlay"
Enter
Wait
Type "clear"
Enter
Show

Type ".\overlay.exe"
Sleep 400ms
Enter
Wait+Screen /Confirm/
Sleep 1.2s

# -- Alert: any key closes it ----------------------------------------------------------

Enter
Wait+Screen /Disk almost full/
Sleep 2s
Escape
Sleep 800ms

# -- Confirm: y answers at once ----------------------------------------------------------

Down
Enter
Wait+Screen /Delete notes.txt/
Sleep 2s
Type "y"
Wait+Screen /deleted notes.txt/
Sleep 1.5s

# -- Prompt: refused while empty, accepted with a name ---------------------------------

Down
Enter
Wait+Screen /a name is required/
Sleep 1.5s
Type "todo.md"
Sleep 1s
Enter
Wait+Screen /created todo.md/
Sleep 1.5s

# -- Help ---------------------------------------------------------------------------------

Down
Enter
Wait+Screen /Navigation/
Sleep 2s
Escape
Sleep 800ms

# -- Pager: a long output, scrolled and closed --------------------------------------------

Down
Enter
Wait+Screen /go test/
Sleep 1.5s
Down@200ms 6
Sleep 800ms
PageDown
Sleep 1s
Type "G"
Wait+Screen /1.284s/
Sleep 1.5s
Escape
Sleep 800ms

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item overlay.exe"
Enter
Wait
