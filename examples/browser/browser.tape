# browser: recipe names left, the recipe right; tab moves the keys across.
#
# Record from the repository root: vhs examples/browser/browser.tape

Output examples/browser/browser.gif

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
Type "go build -o browser.exe ./examples/browser"
Enter
Wait
Type "clear"
Enter
Show

Type ".\browser.exe"
Sleep 400ms
Enter
Wait+Screen /PANCAKES/
Sleep 2s

Down
Sleep 1.2s
Down
Sleep 1.2s
Up@600ms 2
Sleep 1s

Tab
Sleep 800ms
Down@300ms 6
Sleep 1.5s
Tab
Sleep 1s

Enter
Wait+Screen /cooking/
Sleep 2s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item browser.exe"
Enter
Wait
