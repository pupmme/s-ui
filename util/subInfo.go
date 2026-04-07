package util

import (
	"fmt"

	"github.com/pupmme/pupmsub/db"
)

func GetHeaders(client *db.Client, updateInterval int) []string {
	var headers []string
	headers = append(headers, fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", client.Up, client.Down, client.Volume, client.Expiry))
	headers = append(headers, fmt.Sprintf("%d", updateInterval))
	headers = append(headers, client.Name)
	return headers
}
