package palette

import (
	"strings"
	"testing"
)

// Whatever is typed, the palette resolves to an available item the name
// leads to, or to an error: never to a refused item, never to a panic.
func FuzzResolveOnlyRunsWhatTheNameLeadsTo(f *testing.F) {
	for _, s := range []string{"", ":", "goto /a b", "  :c  ", "copy", "delete", "co", "::x", "\t:goto\t"} {
		f.Add(s)
	}
	m := open()
	f.Fuzz(func(t *testing.T, line string) {
		name, arg := splitLine(line)
		if strings.Contains(name, " ") || arg != strings.TrimSpace(arg) {
			t.Fatalf("splitLine(%q) = %q, %q", line, name, arg)
		}
		it, gotArg, err := m.resolve(line)
		if err != nil {
			return
		}
		if !it.Available || !strings.HasPrefix(it.Name, name) || gotArg != arg {
			t.Fatalf("resolve(%q) = %q (available %v), arg %q", line, it.Name, it.Available, gotArg)
		}
	})
}
