package main

import (
	"fmt"
	"runtime"

	"github.com/0xYeah/fltk2go"
	"github.com/0xYeah/fltk2go/foundation"
	"github.com/0xYeah/fltk2go/uikit/label"
	"github.com/0xYeah/fltk2go/uikit/window"
)

func main() {
	runtime.LockOSThread()
	fmt.Println("shmClaw - fltk2go GUI starting...")

	win := window.NewUIWindow(600, 400, "shmClaw - fltk2go")
	root := win.RootView()

	title := label.NewUILabel(&foundation.Rect{X: 20, Y: 20, Width: 560, Height: 40}, "shmClaw FLTK GUI")
	title.SetFontSize(24)

	root.AddSubview(title)
	win.Show()

	fltk2go.Run()
}
