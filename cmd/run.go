package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func runCommand(cmd *cobra.Command, args []string) error {
	command := strings.Join(args, " ")
	fmt.Printf("Running: %s\n", command)

	// Sync files using rsync (keeping this as external process)
	fmt.Println("Syncing files...")
	syncCmd := exec.Command("rsync", "-avz", "--delete",
		"--exclude=.git", "--exclude=out", "--exclude=cache", "--exclude=osiris-lite",
		"-e", "ssh -F "+os.Getenv("HOME")+"/.ssh/config",
		"./", remote+":"+remotePath)
	syncCmd.Stdout = os.Stdout
	syncCmd.Stderr = os.Stderr
	if err := syncCmd.Run(); err != nil {
		return err
	}

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

	return client.RunRemoteCommand(remotePath, image, container, command)
}
