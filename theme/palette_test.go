package theme

import (
	"reflect"
	"testing"

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
	grape, err := Lookup("grape")
	if err != nil || !reflect.DeepEqual(grape, All()["grape"]) {
		t.Fatalf("Lookup(grape) = %v, %v", grape, err)
	}
	if upper, err := Lookup("GRAPE"); err != nil || !reflect.DeepEqual(upper, grape) {
		t.Errorf("Lookup is case-sensitive: %v", err)
	}
	if def, err := Lookup(""); err != nil || !reflect.DeepEqual(def, Default) {
		t.Errorf("an empty name is not the default: %v", err)
	}
	if _, err := Lookup("chartreuse"); err == nil || err.Error() != `unknown theme "chartreuse"` {
		t.Errorf("a misspelling gave %v", err)
	}
}

// The presets are variables, so the registry must read them, not a copy.
func TestLookupSeesAPresetChangedAtRuntime(t *testing.T) {
	old := Banana
	t.Cleanup(func() { Banana = old })
	Banana.Colors.Accent = "#ABCDEF"

	got, err := Lookup("banana")
	require.NoError(t, err)
	require.Equal(t, "#ABCDEF", got.Colors.Accent)
	require.Equal(t, "#ABCDEF", All()["banana"].Colors.Accent)
}
