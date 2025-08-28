package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func logsCommand(cmd *cobra.Command, args []string) error {
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

	var containerID string
	if len(args) > 0 {
		containerID = args[0]
	}

	return client.ConnectToLogs(image, containerID)
}
