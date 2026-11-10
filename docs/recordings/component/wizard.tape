# wizard: a dev container set up over three pages; a check holds the first, the tools filter fuzzily.
#
# Record from the repository root: vhs docs/recordings/component/wizard.tape

Output docs/recordings/component/wizard.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        640
Set Theme         {"name":"Catppuccin Mocha on Mint","black":"#45475a","red":"#f38ba8","green":"#a6e3a1","yellow":"#f9e2af","blue":"#89b4fa","magenta":"#f5c2e7","cyan":"#94e2d5","white":"#bac2de","brightBlack":"#585b70","brightRed":"#f38ba8","brightGreen":"#a6e3a1","brightYellow":"#f9e2af","brightBlue":"#89b4fa","brightMagenta":"#f5c2e7","brightCyan":"#94e2d5","brightWhite":"#a6adc8","background":"#0B1B15","foreground":"#cdd6f4","selection":"#585b70","cursor":"#f5e0dc"}
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
Type "go build -o wizard.exe ./examples/component/wizard"
Enter
Wait
Type "clear"
Enter
Type ".\wizard.exe"
Enter
Wait+Screen /new dev container/
Show
Sleep 1.5s

# -- Project: next without a name is held by the page's check ------------------------------

PageDown
Wait+Screen /needs a name/
Sleep 1.5s
Type@120ms "api"
Sleep 600ms
Tab
Sleep 500ms
Enter
Sleep 600ms
Down@400ms 2
Sleep 600ms
Enter
Sleep 800ms
PageDown
Wait+Screen /Tools/
Sleep 1s

# -- Tools: the list opens, typing filters it, space checks ------------------------------------

Enter
Sleep 800ms
Space
Sleep 600ms
Type@200ms "doc"
Sleep 1s
Space
Sleep 800ms
Backspace@200ms 3
Type@200ms "go"
Sleep 800ms
Space
Sleep 1s
Enter
Sleep 1s
PageDown
Wait+Screen /Finish/
Sleep 1s

# -- Review: back and forth keeps what was chosen, then Finish ----------------------------------

PageUp
Sleep 1.2s
PageDown
Sleep 800ms
Right@500ms 1
Tab
Sleep 500ms
Space
Sleep 800ms
PageDown
Wait+Screen /starts now/
Sleep 3s

Hide
Ctrl+C
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item wizard.exe"
Enter
Wait
