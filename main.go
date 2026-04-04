package main

import (
	"os"

	"github.com/pupmme/sub/app"
	"github.com/pupmme/sub/cmd"
)

func main() {
	if len(os.Args) < 2 {
		app.Start()
		return
	}
	cmd.Execute()
}
