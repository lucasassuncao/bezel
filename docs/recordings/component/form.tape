# form: one of each control, filled in and submitted; the select describes its options while open.
#
# Record from the repository root: vhs docs/recordings/component/form.tape

Output docs/recordings/component/form.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        640
Set Theme         {"name":"Campbell","black":"#0C0C0C","red":"#C50F1F","green":"#13A10E","yellow":"#C19C00","blue":"#0037DA","magenta":"#881798","cyan":"#3A96DD","white":"#CCCCCC","brightBlack":"#767676","brightRed":"#E74856","brightGreen":"#16C60C","brightYellow":"#F9F1A5","brightBlue":"#3B78FF","brightMagenta":"#B4009E","brightCyan":"#61D6D6","brightWhite":"#F2F2F2","background":"#0C0C0C","foreground":"#CCCCCC","selection":"#767676","cursor":"#FFFFFF"}
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
Type "go build -o form.exe ./examples/component/form"
Enter
Wait
Type "clear"
Enter
Type ".\form.exe"
Enter
Wait+Screen /new dev container/
Show
Sleep 1.5s

# -- Name: typed into the input --------------------------------------------------------------

Type@120ms "api"
Sleep 1s

# -- Image: the select opens, the arrows move, enter picks -----------------------------------

Tab
Sleep 600ms
Enter
Wait+Screen /alpine/
Sleep 800ms
Down@400ms 2
Sleep 1.8s
Enter
Sleep 1s

# -- Start: a date, picked from the calendar it opens --------------------------------------

Tab
Sleep 600ms
Enter
Wait+Screen /November 2026/
Sleep 800ms
Down@400ms 1
Sleep 400ms
Right@400ms 2
Sleep 1.5s
Enter
Sleep 1s

# -- Shell, docker and the port: a radio, a toggle and a checkbox -----------------------------

Tab
Sleep 500ms
Right@500ms 1
Sleep 800ms

# -- CPUs and memory: a number steps, a slider slides -----------------------------------------

Tab
Sleep 500ms
Up@400ms 2
Sleep 800ms
Tab
Sleep 500ms
Right@150ms 8
Sleep 1s

# -- Labels: enter turns what is typed into a chip -----------------------------------------

Tab
Sleep 500ms
Type@150ms "web"
Enter
Sleep 600ms
Type@150ms "api"
Enter
Sleep 1s
Tab
Sleep 500ms
Space
Sleep 800ms
Tab
Sleep 500ms
Space
Sleep 1s

# -- Create: the status row reports the values -----------------------------------------------

Tab
Sleep 700ms
Enter
Wait+Screen /docker on/
Sleep 3s

Hide
Ctrl+C
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item form.exe"
Enter
Wait
