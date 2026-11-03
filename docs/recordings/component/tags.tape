# tags: chips that filter a list; enter adds a topic, tab completes it, backspace on an empty field takes the last off.
#
# Record from the repository root: vhs docs/recordings/component/tags.tape

Output docs/recordings/component/tags.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        460
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
Type "go build -o tags.exe ./examples/component/tags"
Enter
Wait
Type "clear"
Enter
Type ".\tags.exe"
Enter
Wait+Screen /8 of 8/
Show
Sleep 1.5s

# -- Two topics, one completed with tab, then the last one off -------------------------

Type@150ms "tu"
Sleep 800ms
Tab
Sleep 600ms
Enter
Sleep 1.5s
Type@150ms "st"
Sleep 800ms
Tab
Sleep 600ms
Enter
Sleep 2s
Backspace
Sleep 1.5s
Type@150ms "framework"
Enter
Sleep 2s

Hide
Escape
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item tags.exe"
Enter
Wait
