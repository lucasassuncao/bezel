# animation, light: the light tab: lights and presets on a card, a diff and its undo
#
# Record from the repository root: vhs docs/recordings/animation/light.tape

Output docs/recordings/animation/light.gif

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

# -- Build and start, off camera -------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o anim-light.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\anim-light.exe"
Enter
Wait+Screen /every animation/
Tab@600ms 5
Wait+Screen /ghost/
Sleep 1s
Show

Sleep 1.5s
Type "f"
Sleep 2s
Type "u"
Sleep 2s
Type "e"
Sleep 2.4s
Type "e"
Sleep 1s
Type "e"
Sleep 1.6s
Type "e"
Sleep 1s
Type "e"
Sleep 1.6s
Type "e"
Sleep 2.6s
Type "e"
Sleep 2.6s
Type "e"
Sleep 1s
Type "e"
Sleep 2.6s
Type "e"
Sleep 1s
Type "e"
Sleep 1s
Type "e"
Sleep 1.2s
Type "e"
Sleep 1s
Type "e"
Sleep 1s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item anim-light.exe"
Enter
Wait
