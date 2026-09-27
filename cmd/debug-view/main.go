package main

import (
	"fmt"
	"os"

	"github.com/dennis605/mdskim2/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: debug-view <workspace> [file]")
		os.Exit(1)
	}
	m := app.New(os.Args[1])
	if len(os.Args) >= 3 {
		m.LoadFileForTest(os.Args[2])
	}
	if len(os.Args) >= 4 {
		// tab: 0=Preview, 1=TOC, 2=Backlinks
		tab := 0
		fmt.Sscanf(os.Args[3], "%d", &tab)
		m.SetRightTab(tab)
	}
	fmt.Println(m.View())
}
