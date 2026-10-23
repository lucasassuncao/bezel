# inline: a CLI that logs, asks one question below the logs, and draws a bar.
#
# Record from the repository root: vhs examples/inline/inline.tape

Output examples/inline/inline.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        400
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
Type "go build -o inline.exe ./examples/inline"
Enter
Wait
# clear runs first, so the command line is gone before the camera rolls, and
# the sleep after holds the last frame without a prompt under it.
Type "clear; .\inline.exe; Start-Sleep 60"
Enter
Wait+Screen /fetching index/
Show

Wait+Screen /Download them now/
Sleep 1.5s
Type "y"
Wait+Screen /done/
Sleep 2.5s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Ctrl+C
Sleep 500ms
Type "Remove-Item inline.exe"
Enter
Wait
