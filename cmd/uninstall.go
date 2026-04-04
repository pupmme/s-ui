package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall sub (stop service, remove files)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping sub service...")
		exec.Command("systemctl", "stop", "sub").Run()
		exec.Command("systemctl", "disable", "sub").Run()
		exec.Command("rm", "/etc/systemd/system/sub.service").Run()

		files := []string{
			"/usr/local/bin/sub",
			"/etc/sub",
			"/var/www/sub",
		}
		for _, f := range files {
			if _, err := os.Stat(f); err == nil {
				if err := os.RemoveAll(f); err != nil {
					fmt.Println("Failed to remove", f, ":", err)
				} else {
					fmt.Println("Removed:", f)
				}
			}
		}
		fmt.Println("sub uninstalled.")
	},
}
