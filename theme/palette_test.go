package theme

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestThemeRegistryHasNoDuplicateNames guards against copy-paste: a name
// listed twice would silently overwrite a theme in All() and print twice in
// Categories(), with no compiler error.
func TestThemeRegistryHasNoDuplicateNames(t *testing.T) {
	seen := make(map[string]string) // theme name -> category that claimed it

	for _, cat := range themeRegistry {
		for _, nt := range cat.themes {
			if prev, dup := seen[nt.name]; dup {
				t.Errorf("themeRegistry: %q appears in both %q and %q", nt.name, prev, cat.category)
			}
			seen[nt.name] = cat.category
		}
	}
}

// A name typed on a command line: case does not matter, empty is the default,
// and a misspelling is refused by name.
func TestLookupFindsAThemeByName(t *testing.T) {
	mint, err := Lookup("mint")
	if err != nil || !reflect.DeepEqual(mint, Mint()) {
		t.Fatalf("Lookup(mint) = %v, %v", mint, err)
	}
	if upper, err := Lookup("MINT"); err != nil || !reflect.DeepEqual(upper, mint) {
		t.Errorf("Lookup is case-sensitive: %v", err)
	}
	if def, err := Lookup(""); err != nil || !reflect.DeepEqual(def, Default()) {
		t.Errorf("an empty name is not the default: %v", err)
	}
	if _, err := Lookup("chartreuse"); err == nil || err.Error() != `unknown theme "chartreuse"` {
		t.Errorf("a misspelling gave %v", err)
	}
}

// A preset is built fresh each time: changing one copy reaches no other.
func TestAPresetCannotBeChangedForEveryone(t *testing.T) {
	m := Mint()
	m.Colors.Accent = "#ABCDEF"
	got, err := Lookup("mint")
	require.NoError(t, err)
	require.NotEqual(t, "#ABCDEF", got.Colors.Accent)
	require.NotEqual(t, "#ABCDEF", Mint().Colors.Accent)
	require.NotEqual(t, "#ABCDEF", All()["mint"].Colors.Accent)
}

// A redirected end must not reach lipgloss's query: on Windows it reads the
// console instead, which can wait forever for an answer that never comes.
func TestDarkTerminalSkipsTheQueryWhenRedirected(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()

	done := make(chan bool, 1)
	go func() { done <- darkTerminal(r, w) }()
	select {
	case dark := <-done:
		require.True(t, dark, "no answer reads as dark, as lipgloss does")
	case <-time.After(time.Second):
		t.Fatal("darkTerminal asked a pipe")
	}
}
