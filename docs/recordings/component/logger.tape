# logger: a run logged as a tree under each message, wrapped at the width.
#
# Record from the repository root: vhs docs/recordings/component/logger.tape

Output docs/recordings/component/logger.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        900
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
Type "go build -o logger.exe ./examples/component/logger"
Enter
Wait
# clear runs first, so the command line is gone before the camera rolls, and
# the sleep after holds the last frame without a prompt under it.
Type "clear; .\logger.exe; Start-Sleep 60"
Enter
Wait+Screen /No .jpg/
Show

Wait+Screen /batch_id/
Sleep 3s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Ctrl+C
Sleep 500ms
Type "Remove-Item logger.exe"
Enter
Wait
