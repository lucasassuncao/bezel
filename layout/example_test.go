package layout_test

import (
	"fmt"

	"github.com/lucasassuncao/bezel/layout"
)

func ExampleResolve() {
	screen := layout.Rows(
		layout.Fixed("header", layout.Lines(1)),
		layout.Columns(
			layout.Fixed("sidebar", layout.Ratio(1, 4), layout.Min(20)),
			layout.Fill("main"),
		),
	)
	rects, _ := layout.Resolve(screen, layout.Rect{W: 100, H: 30}, nil)
	for _, name := range layout.Order(screen) {
		fmt.Printf("%-7s %+v\n", name, rects[name])
	}
	// Output:
	// header  {X:0 Y:0 W:100 H:1}
	// sidebar {X:0 Y:1 W:25 H:29}
	// main    {X:25 Y:1 W:75 H:29}
}

// A leaf the present func rejects gives its space to its siblings.
func ExampleResolve_present() {
	screen := layout.Columns(layout.Fixed("sidebar", layout.Lines(30)), layout.Fill("main"))
	noSidebar := func(name string) bool { return name != "sidebar" }
	rects, _ := layout.Resolve(screen, layout.Rect{W: 100, H: 30}, noSidebar)
	fmt.Printf("%+v\n", rects["main"])
	// Output:
	// {X:0 Y:0 W:100 H:30}
}

func ExampleBox_Collapse() {
	screen := layout.Columns(layout.Fill("list"), layout.Fill("detail")).Collapse(80, "detail")
	rects, collapsed := layout.Resolve(screen, layout.Rect{W: 60, H: 20}, nil)
	_, hasList := rects["list"]
	fmt.Println(collapsed, hasList, fmt.Sprintf("%+v", rects["detail"]))
	// Output:
	// true false {X:0 Y:0 W:60 H:20}
}
