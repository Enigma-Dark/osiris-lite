package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func statusCommand(cmd *cobra.Command, args []string) error {
	fmt.Println("Checking status...")

	var client *SSHClient
	var err error

	if password != "" {
		client, err = NewSSHClientWithPassword(remote, password)
	} else {
		client, err = NewSSHClient(remote)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to remote: %w", err)
	}
	defer client.Close()

	return client.GetStatus(image)
}
