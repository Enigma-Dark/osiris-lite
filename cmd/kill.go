package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func killCommand(cmd *cobra.Command, args []string) error {
	target := ""
	if len(args) > 0 {
		target = args[0]
	}

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

	if target == "all" {
		fmt.Println("Killing all jobs...")
		return client.KillAll(image)
	}

	if target != "" {
		fmt.Printf("Killing container: %s\n", target)
		return client.KillContainer(target)
	}

	fmt.Println("\nUse 'kill all' or 'kill <container_id>'")
	return nil
}
