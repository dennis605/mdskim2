package main

import (
	"fmt"
	"os"

	"github.com/dennis605/mdskim2/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: debug-view <workspace>")
		os.Exit(1)
	}
	m := app.New(os.Args[1])
	fmt.Println(m.View())
}
