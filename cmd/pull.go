package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func pullCommand(cmd *cobra.Command, args []string) error {
	// Use resultsPath flag as default, but allow override with argument
	if len(args) > 0 {
		resultsPath = args[0]
	}

	fmt.Printf("Pulling results to: %s\n", resultsPath)

	// Use SSH client for remote execution
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

	return client.PullResults(remotePath, resultsPath)
}
