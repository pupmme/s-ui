package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "View sub logs (alias: journalctl -u sub -f)",
	Run: func(cmd *cobra.Command, args []string) {
		cmd2 := exec.Command("journalctl", "-u", "sub", "-f", "--no-pager")
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		cmd2.Run()
	},
}

var logLines int
var logCmd2 = &cobra.Command{
	Use:   "log",
	Short: "Show last N lines of log",
	Run: func(cmd *cobra.Command, args []string) {
		cmd2 := exec.Command("journalctl", "-u", "sub", "-n", fmt.Sprintf("%d", logLines), "--no-pager")
		var out bytes.Buffer
		cmd2.Stdout = &out
		cmd2.Run()
		fmt.Print(strings.TrimSpace(out.String()))
	},
}
