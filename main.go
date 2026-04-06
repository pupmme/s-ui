package main

import (
	"os"

	"github.com/pupmme/pupmmesub/app"
	"github.com/pupmme/pupmmesub/cmd"
)

func main() {
	if len(os.Args) < 2 {
		app.Start()
		return
	}
	cmd.Execute()
}
