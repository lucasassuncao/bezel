# paginator: 42 items, 8 to a page; the arrows turn the page, t switches to dots.
#
# Record from the repository root: vhs examples/paginator/paginator.tape

Output examples/paginator/paginator.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        640
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
Type "go build -o paginator.exe ./examples/paginator"
Enter
Wait
Type "clear"
Enter
Type ".\paginator.exe"
Enter
Wait+Screen /1\/6/
Show
Sleep 1.5s

# -- Forward as numbers, then back as dots ------------------------------------------

Right@900ms 3
Sleep 1s
Type "t"
Sleep 1.5s
Left@900ms 2
Sleep 1s
Type "t"
Sleep 1.5s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item paginator.exe"
Enter
Wait
