package main

import (
	"os"

	"github.com/pupmme/pupmsub/app"
	"github.com/pupmme/pupmsub/cmd"
)

func main() {
	if len(os.Args) < 2 {
		app.Start()
		return
	}
	cmd.Execute()
}
