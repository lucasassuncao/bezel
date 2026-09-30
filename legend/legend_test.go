package legend

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// A nil can must not hand back the caller's own slice.
func TestFilterAlwaysReturnsACopy(t *testing.T) {
	in := []Entry{New("a", "alpha")}
	out := Filter(in, nil)
	out[0] = New("b", "beta")
	require.Equal(t, "alpha", in[0].Help().Desc)
}

func entries(n int) []Entry {
	out := make([]Entry, n)
	for i := range out {
		out[i] = New(string(rune('a'+i)), "action number "+string(rune('a'+i)))
	}
	return out
}

func TestPackDropsWholeEntriesFromTheEndAndCountsThem(t *testing.T) {
	lines := Pack(entries(8), 40, 2, Style{HelpKey: "?"})
	require.Len(t, lines, 2)
	for _, l := range lines {
		require.LessOrEqual(t, ansi.StringWidth(l), 40)
	}
	last := ansi.Strip(lines[1])
	require.Contains(t, last, "in [?]")
	require.NotContains(t, strings.Join(lines, ""), "action number h")
}

func TestPackKeepsTheFirstEntryHoweverNarrow(t *testing.T) {
	lines := Pack(entries(3), 5, 1, Style{})
	require.Len(t, lines, 1)
	require.Contains(t, ansi.Strip(lines[0]), "[a]")
	require.LessOrEqual(t, ansi.StringWidth(lines[0]), 5)
}

func TestPackFitsEverythingWhenThereIsRoom(t *testing.T) {
	lines := Pack(entries(3), 200, 2, Style{})
	require.Len(t, lines, 1)
	require.NotContains(t, ansi.Strip(lines[0]), "in [?]")
}

func TestPackOfNothingIsNothing(t *testing.T) {
	require.Nil(t, Pack(nil, 80, 2, Style{}))
}

func TestFilterDropsWhatCannotRun(t *testing.T) {
	es := []Entry{New("r", "read"), New("w", "write", "write"), New("d", "delete", "write")}
	kept := Filter(es, func(c Capability) bool { return c != "write" })
	require.Len(t, kept, 1)
	require.Equal(t, "r", kept[0].Help().Key)
	require.Len(t, Filter(es, nil), 3)
}

func TestHintLineJoinsWithDots(t *testing.T) {
	got := ansi.Strip(HintLine(entries(2), Style{}))
	require.Equal(t, "[a] action number a · [b] action number b", got)
}

// SetEnabled(false) on an Entry must hide it, as FromBindings already does.
func TestDisabledEntriesStayOutOfEveryView(t *testing.T) {
	off := New("x", "off")
	off.SetEnabled(false)
	es := []Entry{off, New("y", "on")}

	for name, can := range map[string]func(Capability) bool{"nil can": nil, "some can": func(Capability) bool { return true }} {
		kept := Filter(es, can)
		require.Len(t, kept, 1, name)
		require.Equal(t, "y", kept[0].Help().Key, name)
	}
	require.NotContains(t, ansi.Strip(Pack(es, 80, 2, Style{})[0]), "off")
	require.NotContains(t, ansi.Strip(HintLine(es, Style{})), "off")
}

func TestFromBindingsSkipsDisabled(t *testing.T) {
	off := key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "off"))
	off.SetEnabled(false)
	on := key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "on"))
	es := FromBindings([]key.Binding{off, on})
	require.Len(t, es, 1)
	require.Equal(t, "y", es[0].Help().Key)
}

func TestPackNamesTheHelpKeyOnlyWhenThereIsOne(t *testing.T) {
	more := Pack(entries(8), 40, 1, Style{})
	require.Contains(t, ansi.Strip(more[0]), "more")
	require.NotContains(t, ansi.Strip(more[0]), "[?]")
	help := Pack(entries(8), 40, 1, Style{HelpKey: "?"})
	require.Contains(t, ansi.Strip(help[0]), "in [?]")
}

func TestPackOpensWithHelpWhenEverythingFits(t *testing.T) {
	lines := Pack(entries(3), 200, 2, Style{HelpKey: "?"})
	require.Len(t, lines, 1)
	require.True(t, strings.HasPrefix(ansi.Strip(lines[0]), "[?] help  [a] action number a"))
	require.NotContains(t, ansi.Strip(lines[0]), "in [?]")
}

// An app that lists help decides where it goes; it is not pinned twice.
func TestPackShowsHelpOnceWhereTheAppListsIt(t *testing.T) {
	es := append(entries(2), New("?", "help"))
	line := ansi.Strip(Pack(es, 200, 2, Style{HelpKey: "?"})[0])
	require.Equal(t, 1, strings.Count(line, "[?]"))
	require.True(t, strings.HasSuffix(line, "[?] help"))
}

func TestPackKeepsHelpHoweverNarrow(t *testing.T) {
	lines := Pack(entries(3), 12, 1, Style{HelpKey: "?"})
	require.Len(t, lines, 1)
	require.LessOrEqual(t, ansi.StringWidth(lines[0]), 12)
	require.True(t, strings.HasPrefix(ansi.Strip(lines[0]), "[?] help"))
}

func TestPackOfOnlyHelpIsHelp(t *testing.T) {
	lines := Pack(nil, 80, 2, Style{HelpKey: "?"})
	require.Equal(t, []string{"[?] help"}, []string{ansi.Strip(lines[0])})
}

func grouped(keys string, group int) Entry {
	e := New(keys, "do "+keys)
	e.Group = group
	return e
}

func strip(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = ansi.Strip(l)
	}
	return out
}

// Each group opens its own row when the rows allow it, however much room is
// left on the one before; declaration order between groups does not matter.
func TestPackStartsEachGroupOnItsOwnRow(t *testing.T) {
	in := []Entry{grouped("s", 2), grouped("a", 0), grouped("e", 1), grouped("b", 0), grouped("u", 1)}
	lines := strip(Pack(in, 200, 3, Style{HelpKey: "?"}))
	require.Equal(t, []string{
		"[?] help  [a] do a  [b] do b",
		"[e] do e  [u] do u",
		"[s] do s",
	}, lines)
}

// A group too long for one row wraps inside its own rows.
func TestAGroupWrapsOnlyIntoItsOwnRows(t *testing.T) {
	in := []Entry{grouped("a", 0), grouped("b", 0), grouped("c", 1)}
	lines := strip(Pack(in, 10, 3, Style{}))
	require.Equal(t, []string{"[a] do a", "[b] do b", "[c] do c"}, lines)
}

// More groups than rows: the groups keep their order but share rows, and what
// still does not fit is counted as before.
func TestGroupsFallBackToFillingWhenTheRowsRunOut(t *testing.T) {
	in := []Entry{grouped("a", 0), grouped("b", 1), grouped("c", 2)}
	require.Equal(t, []string{"[a] do a  [b] do b  [c] do c"}, strip(Pack(in, 80, 1, Style{})))

	lines := strip(Pack(entries(8), 40, 2, Style{HelpKey: "?"}))
	require.Contains(t, lines[1], "in [?]", "ungrouped entries pack as they always did")
}

// A group with nothing in it takes no row.
func TestAnEmptyGroupTakesNoRow(t *testing.T) {
	in := []Entry{grouped("a", 0), grouped("c", 2)}
	require.Equal(t, []string{"[a] do a", "[c] do c"}, strip(Pack(in, 80, 3, Style{})))
}
