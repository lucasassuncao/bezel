// Ask: one form field asked below a CLI's output, with nothing taking the
// screen. The field comes from the command line: ask select, ask tags, and
// so on; each prints a line before the question and one with the answer.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/component/form"
	"github.com/lucasassuncao/bezel/component/inline"
	"github.com/lucasassuncao/bezel/theme"
)

// scene is one question: what the CLI printed before it, the field, and what
// it does with the answer.
type scene struct {
	before string
	field  form.Field
	after  func(form.Field) string
}

// scenes are the fields, by the name given on the command line.
func scenes() map[string]scene {
	release := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	return map[string]scene{
		"input": {"creating a project", form.NewInput("Name").WithPlaceholder("my-api").WithWidth(24),
			func(f form.Field) string { return "created ./" + f.(form.Input).Value() }},
		"select": {"3 regions available", form.NewSelect("Region", []string{"europe", "us-east", "asia"}),
			func(f form.Field) string { return "deploying to " + f.(form.Select).Value() }},
		"multiselect": {"8 tools found", form.NewMultiSelect("Tools", []string{"git", "docker", "node", "golang", "python", "rust", "deno", "dotnet"}).WithHeight(5),
			func(f form.Field) string {
				return "installing " + strings.Join(f.(form.MultiSelect).Values(), " ")
			}},
		"radio": {"setting up the container", form.NewRadio("Shell", []string{"bash", "zsh", "fish"}),
			func(f form.Field) string { return "shell set to " + f.(form.Radio).Value() }},
		"checkbox": {"the container is built", form.NewCheckbox("Start it now"),
			func(f form.Field) string {
				return map[bool]string{true: "started", false: "left stopped"}[f.(form.Checkbox).Checked()]
			}},
		"toggle": {"first run", form.NewToggle("Telemetry"),
			func(f form.Field) string {
				return map[bool]string{true: "telemetry on", false: "telemetry off"}[f.(form.Toggle).On()]
			}},
		"number": {"port 8080 is taken", form.NewNumber("Port", 8081).WithRange(1024, 65535),
			func(f form.Field) string { return fmt.Sprintf("listening on :%v", f.(form.Number).Value()) }},
		"slider": {"8 cores", form.NewSlider("Workers", 4).WithRange(1, 8),
			func(f form.Field) string { return fmt.Sprintf("starting %v workers", f.(form.Slider).Value()) }},
		"date": {"v1.4.0 is ready", form.NewDate("Release", release),
			func(f form.Field) string { return "scheduled for " + f.(form.Date).Value() }},
		"tags": {"opening an issue", form.NewTags("Labels").WithSuggestions([]string{"bug", "docs", "feature", "ui"}),
			func(f form.Field) string { return "labelled " + strings.Join(f.(form.Tags).Tags(), ", ") }},
	}
}

// names are the scenes' names in order, for the usage line.
func names() []string {
	var out []string
	for n := range scenes() {
		out = append(out, n)
	}
	slices.Sort(out)
	return out
}

// run is main with its terminal passed in: w gets the CLI's own lines, opts
// reach the question's program, so a test can answer it.
func run(w io.Writer, name string, th theme.Resolved, opts ...tea.ProgramOption) error {
	s, ok := scenes()[name]
	if !ok {
		return fmt.Errorf("usage: ask %s", strings.Join(names(), "|"))
	}
	fmt.Fprintln(w, s.before)
	got, err := inline.Ask(s.field, th, opts...)
	if errors.Is(err, inline.ErrAborted) {
		fmt.Fprintln(w, "nothing changed")
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(w, s.after(got))
	return nil
}

func main() {
	th := theme.Resolve(theme.Mint(), true)
	name := ""
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	if err := run(os.Stdout, name, th); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
