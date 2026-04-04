package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var webStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show web status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Web UI: http://localhost:2053")
		fmt.Println("Config: /etc/sub/config.json")
		fmt.Println("Data: /etc/sub/singbox.json")
		fmt.Println("Binary: /usr/local/bin/sub")
	},
}

var webAdminCmd = &cobra.Command{
	Use:   "admin [username]",
	Short: "Show admin user",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("admin")
		} else {
			fmt.Println("username:", args[0])
		}
	},
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Web UI management",
}
