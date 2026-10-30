// Package doccheck type-checks the Go snippets in the guides against bezel's
// API, so a guide that names a function bezel no longer has fails the build.
package doccheck

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// imports is every package a snippet may name, under the name it uses.
var imports = map[string]string{
	"bezel":     "github.com/lucasassuncao/bezel",
	"animation": "github.com/lucasassuncao/bezel/animation",
	"text":      "github.com/lucasassuncao/bezel/animation/text",
	"scene":     "github.com/lucasassuncao/bezel/animation/scene",
	"effect":    "github.com/lucasassuncao/bezel/effect",
	"draw":      "github.com/lucasassuncao/bezel/draw",
	"key":       "github.com/lucasassuncao/bezel/key",
	"layout":    "github.com/lucasassuncao/bezel/layout",
	"legend":    "github.com/lucasassuncao/bezel/legend",
	"theme":     "github.com/lucasassuncao/bezel/theme",
	"overlay":   "github.com/lucasassuncao/bezel/overlay",
	"shell":     "github.com/lucasassuncao/bezel/shell",
	"bezeltest": "github.com/lucasassuncao/bezel/bezeltest",
	"tea":       "charm.land/bubbletea/v2",
	"lipgloss":  "charm.land/lipgloss/v2",
	"fmt":       "fmt",
	"strings":   "strings",
	"time":      "time",
	"os":        "os",
	"slog":      "log/slog",
}

// components are the packages under component/, each imported by its name.
var components = []string{
	"accordion", "browser", "chart", "cursor", "filepicker", "form", "icon", "inline", "input", "list",
	"logger", "paginator", "progress", "spinner", "stopwatch", "table", "textbox", "themebrowser",
	"timer", "toast", "tree", "viewport",
}

// ours are the names a problem must mention to be about bezel's API.
var ours = func() []string {
	out := slices.Clone(components)
	for name, path := range imports {
		if strings.Contains(path, "/bezel") {
			out = append(out, name)
		}
	}
	return out
}()

// Problem is one thing a snippet says that bezel's API does not have.
type Problem struct {
	Line int // the snippet's first line in its file
	Msg  string
}

var fence = regexp.MustCompile("(?s)```go\n(.*?)```")

// Checker type-checks snippets against compiled export data, which go list
// builds once: parsing bezel and its dependencies for every snippet takes
// minutes.
type Checker struct {
	fset *token.FileSet
	imp  types.Importer
}

// NewChecker lists the export data of every package a snippet may import.
func NewChecker() (*Checker, error) {
	args := []string{"list", "-export", "-deps", "-f", "{{.ImportPath}}={{.Export}}"}
	for _, path := range imports {
		args = append(args, path)
	}
	for _, name := range components {
		args = append(args, "github.com/lucasassuncao/bezel/component/"+name)
	}
	out, err := exec.Command("go", args...).Output() // #nosec G204 -- fixed arguments
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}
	exports := map[string]string{}
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if path, file, ok := strings.Cut(strings.TrimSpace(l), "="); ok && file != "" {
			exports[path] = file
		}
	}
	fset := token.NewFileSet()
	lookup := func(path string) (io.ReadCloser, error) { return os.Open(exports[path]) }
	return &Checker{fset: fset, imp: importer.ForCompiler(fset, "gc", lookup)}, nil
}

// Check returns every problem the Go snippets of the markdown in src have
// with bezel's API; what they say about names of their own is left alone.
func (c *Checker) Check(src string) []Problem {
	var out []Problem
	for _, m := range fence.FindAllStringSubmatchIndex(src, -1) {
		line := strings.Count(src[:m[2]], "\n") + 1
		for _, msg := range c.check(src[m[2]:m[3]]) {
			out = append(out, Problem{Line: line, Msg: msg})
		}
	}
	return out
}

// ReadFile is Check over the markdown file at path.
func (c *Checker) ReadFile(path string) ([]Problem, error) {
	b, err := os.ReadFile(path) // #nosec G304 -- the repo's own guides
	if err != nil {
		return nil, err
	}
	return c.Check(string(b)), nil
}

// check type-checks one snippet in the first shape that parses; one that
// parses in none is prose in a go fence, as a signature with "...".
func (c *Checker) check(body string) []string {
	for _, file := range shapes(body) {
		f, err := parser.ParseFile(c.fset, "snippet.go", file, parser.SkipObjectResolution)
		if err == nil {
			return c.typecheck(f)
		}
	}
	return nil
}

// shapes are the ways body could sit in a file: as declarations, as
// statements in a function, or as statements followed by declarations.
func shapes(body string) []string {
	var head strings.Builder
	head.WriteString("package snippet\n\nimport (\n")
	for name, path := range imports {
		fmt.Fprintf(&head, "\t%s %q\n", name, path)
	}
	for _, name := range components {
		fmt.Fprintf(&head, "\t%s %q\n", name, "github.com/lucasassuncao/bezel/component/"+name)
	}
	head.WriteString(")\n\n")
	h := head.String()
	out := []string{h + body, h + "func _() {\n" + body + "\n}\n"}
	if at := strings.Index(body, "\nfunc "); at >= 0 {
		out = append(out, h+"func _() {\n"+body[:at]+"\n}\n"+body[at:])
	}
	return out
}

// typecheck reports the errors that name one of bezel's packages: the
// snippet's own names, undeclared here, are not the guide's fault.
func (c *Checker) typecheck(f *ast.File) []string {
	var msgs []string
	conf := types.Config{
		Importer: c.imp,
		Error: func(err error) {
			if msg := err.(types.Error).Msg; about(msg) {
				msgs = append(msgs, msg)
			}
		},
	}
	_, _ = conf.Check("snippet", c.fset, []*ast.File{f}, nil)
	return msgs
}

// about reports a message about bezel's API rather than the snippet's names.
func about(msg string) bool {
	if strings.HasSuffix(msg, "not used") { // an import, a variable, a bare shell.Config{...} shown alone
		return false
	}
	for _, name := range ours {
		if strings.Contains(msg, name+".") {
			return true
		}
	}
	return false
}
