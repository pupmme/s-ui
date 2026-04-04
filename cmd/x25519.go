package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"
)

var x25519Cmd = &cobra.Command{
	Use:   "x25519",
	Short: "Generate x25519 key pair",
	Run: func(cmd *cobra.Command, args []string) {
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			fmt.Println("Generate failed:", err)
			return
		}
		pub := make([]byte, 32)
		copy(pub, priv[32:])
		fmt.Println("Private key:", base64.StdEncoding.EncodeToString(priv))
		fmt.Println("Public key:", base64.StdEncoding.EncodeToString(pub))
	},
}
