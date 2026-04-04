package cmd

import (
	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/service"
	"github.com/pupmme/sub/web"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Web UI management",
}

var webStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show web status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		logger.Info("Web port: ", cfg.Web.Port)
		logger.Info("Cert: ", cfg.Web.Cert)
		logger.Info("Key: ", cfg.Web.Key)
	},
}

var webAdminCmd = &cobra.Command{
	Use:   "admin [username] [password]",
	Short: "Set admin credentials",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		config.Load()
		cfg := config.Get()
		cfg.Web.Username = args[0]
		hash, err := bcrypt.GenerateFromPassword([]byte(args[1]), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("hash password failed: ", err)
			return
		}
		cfg.Web.Password = string(hash)
		config.Set(cfg)
		if err := config.Save(); err != nil {
			logger.Error("save config failed: ", err)
			return
		}
		logger.Info("admin credentials updated")
	},
}

func init() {
	webCmd.AddCommand(webStatusCmd)
	webCmd.AddCommand(webAdminCmd)
}
