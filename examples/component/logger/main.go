// Logger: a run that sorts downloads, logged as a tree under each message;
// --json prints the same calls as JSON lines.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lucasassuncao/bezel/component/logger"
	"github.com/lucasassuncao/bezel/theme"
)

// move is one file the run finds and where it goes.
type move struct{ category, name, dest string }

var moves = []move{
	{"installers", "GitHubDesktopSetup-x64.exe", `C:\Users\lucas\Downloads\Installers\exe`},
	{"installers", "tor-browser-windows-x86_64-portable-15.0.26.exe", `C:\Users\lucas\Downloads\Installers\exe`},
	{"installers", "go1.27.3.windows-amd64.msi", `C:\Users\lucas\Downloads\Installers\msi`},
	{"everything-else", "Gittyup-2.0.0-x86_64.AppImage", `C:\Users\lucas\Downloads\ToOrganize`},
}

// newLogger is the tree on a terminal, or JSON lines for --json.
func newLogger(w io.Writer, asJSON bool) *slog.Logger {
	if asJSON {
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: logger.LevelTrace, ReplaceAttr: logger.ReplaceAttr}))
	}
	return slog.New(logger.New(w, theme.Resolve(theme.Mint, true)).WithLevel(logger.LevelTrace))
}

func run(log *slog.Logger, pause time.Duration) {
	log.Info("[images] No .jpg files found")
	time.Sleep(pause)
	log.Log(context.Background(), logger.LevelTrace, "scanning", "dir", `C:\Users\lucas\Downloads`)
	time.Sleep(pause)
	for _, m := range moves {
		log.Warn(fmt.Sprintf("[%s] 1 file to move", m.category), "name", m.name)
		time.Sleep(pause)
		log.Info(fmt.Sprintf("[%s] Moved 1 file", m.category),
			"source", `C:\Users\lucas\Downloads\`+m.name,
			"destination", m.dest+`\`+m.name)
		time.Sleep(pause)
	}
	log.Error("[fonts] cannot read the folder", "err", "open C:\\Users\\lucas\\Fonts: access is denied")
	time.Sleep(pause)
	log.Info("run complete",
		"moved", len(moves),
		slog.Group("skipped", "files", 0, "categories", 1),
		"batch_id", "batch_146ba710dbd6bd7b")
}

func main() {
	asJSON := flag.Bool("json", false, "print JSON lines")
	flag.Parse()
	run(newLogger(os.Stdout, *asJSON), 300*time.Millisecond)
}
