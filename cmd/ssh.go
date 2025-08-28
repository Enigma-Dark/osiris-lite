package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	client *ssh.Client
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func loadSSHKey(keyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(expandPath(keyPath))
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

func NewSSHClient(hostAlias string) (*SSHClient, error) {
	return NewSSHClientWithPassword(hostAlias, "")
}

func NewSSHClientWithPassword(hostAlias, pwd string) (*SSHClient, error) {
	// Extract settings for the host alias using ssh_config.Get
	host := ssh_config.Get(hostAlias, "HostName")
	if host == "" {
		return nil, fmt.Errorf("no hostname found for host '%s' in SSH config", hostAlias)
	}

	user := ssh_config.Get(hostAlias, "User")
	if user == "" {
		return nil, fmt.Errorf("no user found for host '%s' in SSH config", hostAlias)
	}

	port := ssh_config.Get(hostAlias, "Port")
	if port == "" {
		port = "22"
	}

	proxyJump := ssh_config.Get(hostAlias, "ProxyJump")
	identityFile := ssh_config.Get(hostAlias, "IdentityFile")

	// Prepare authentication methods
	var authMethods []ssh.AuthMethod

	// Try SSH key first if available
	if identityFile != "" && identityFile != "~/.ssh/identity" {
		if auth, err := loadSSHKey(identityFile); err == nil {
			authMethods = append(authMethods, auth)
		}
	}

	// Add password authentication if provided
	if pwd != "" {
		authMethods = append(authMethods, ssh.Password(pwd))
	}

	// Add keyboard-interactive for password authentication
	authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		// For now, return empty to skip interactive prompts
		// TODO: Add proper password support
		return make([]string, len(questions)), nil
	}))

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available for host '%s'", hostAlias)
	}

	// Configure SSH client
	sshConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	var client *ssh.Client
	var err error

	// Handle ProxyJump
	if proxyJump != "" {
		client, err = connectWithProxy(hostAlias, proxyJump, sshConfig, pwd)
		if err != nil {
			return nil, fmt.Errorf("failed to connect via proxy: %w", err)
		}
	} else {
		// Direct connection
		addr := net.JoinHostPort(host, port)
		client, err = ssh.Dial("tcp", addr, sshConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect directly: %w", err)
		}
	}

	return &SSHClient{client: client}, nil
}

func connectWithProxy(targetAlias, proxyAlias string, targetConfig *ssh.ClientConfig, targetPassword string) (*ssh.Client, error) {
	// Get proxy settings
	proxyHost := ssh_config.Get(proxyAlias, "HostName")
	proxyUser := ssh_config.Get(proxyAlias, "User")
	proxyPort := ssh_config.Get(proxyAlias, "Port")
	if proxyPort == "" {
		proxyPort = "22"
	}
	proxyIdentityFile := ssh_config.Get(proxyAlias, "IdentityFile")

	// Load proxy SSH key
	proxyAuth, err := loadSSHKey(proxyIdentityFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load proxy SSH key: %w", err)
	}

	// Configure proxy SSH client
	proxyConfig := &ssh.ClientConfig{
		User:            proxyUser,
		Auth:            []ssh.AuthMethod{proxyAuth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to proxy
	proxyAddr := net.JoinHostPort(proxyHost, proxyPort)
	proxyClient, err := ssh.Dial("tcp", proxyAddr, proxyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}

	// Get target settings
	targetHost := ssh_config.Get(targetAlias, "HostName")
	targetPort := ssh_config.Get(targetAlias, "Port")
	if targetPort == "" {
		targetPort = "22"
	}

	// Dial target through proxy
	targetAddr := net.JoinHostPort(targetHost, targetPort)
	proxyConn, err := proxyClient.Dial("tcp", targetAddr)
	if err != nil {
		proxyClient.Close()
		return nil, fmt.Errorf("failed to dial target through proxy: %w", err)
	}

	// Create SSH connection to target
	targetConn, targetChans, targetReqs, err := ssh.NewClientConn(proxyConn, targetAddr, targetConfig)
	if err != nil {
		proxyConn.Close()
		proxyClient.Close()
		return nil, fmt.Errorf("failed to create target SSH connection: %w", err)
	}

	targetClient := ssh.NewClient(targetConn, targetChans, targetReqs)
	return targetClient, nil
}

func (s *SSHClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *SSHClient) RunCommand(command string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}
	return string(output), nil
}

func (s *SSHClient) RunCommandWithLiveOutput(command string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Set up pipes to stream output in real-time
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Run(command); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func (s *SSHClient) GetStatus(image string) error {
	fmt.Println("📋 Checking status on remote server...")

	// Check Docker containers
	fmt.Println("┌─ Docker Containers")
	containersCmd := fmt.Sprintf(`docker ps --filter "ancestor=%s" --format "{{.ID}} {{.Names}} {{.Status}} ({{.RunningFor}})" 2>/dev/null || true`, image)
	containers, err := s.RunCommand(containersCmd)
	if err != nil {
		return fmt.Errorf("failed to check containers: %w", err)
	}

	if strings.TrimSpace(containers) == "" {
		fmt.Println("│  No active containers")
	} else {
		for _, line := range strings.Split(strings.TrimSpace(containers), "\n") {
			if line != "" {
				fmt.Printf("│  %s\n", line)
			}
		}
	}
	fmt.Println("│")

	// Check fuzzer processes
	fmt.Println("├─ Fuzzer Processes")
	processesCmd := `pgrep -a -i fuzzer || true`
	processes, err := s.RunCommand(processesCmd)
	if err != nil {
		return fmt.Errorf("failed to check processes: %w", err)
	}

	if strings.TrimSpace(processes) == "" {
		fmt.Println("│  No active processes")
	} else {
		for _, line := range strings.Split(strings.TrimSpace(processes), "\n") {
			if line != "" {
				// Just show PID and command name
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					pid := parts[0]
					command := parts[1]
					fmt.Printf("│  %s: %s\n", pid, command)
				} else {
					fmt.Printf("│  %s\n", line)
				}
			}
		}
	}
	fmt.Println("│")

	// Check System Resources
	fmt.Println("└─ System Resources")

	// CPU
	cpuCmd := `top -bn1 | grep "Cpu(s)" | awk '{print $2 + $4 "%"}' || echo "N/A"`
	cpu, err := s.RunCommand(cpuCmd)
	if err == nil {
		fmt.Printf("   CPU: %s", strings.TrimSpace(cpu))
	}

	// Memory
	memCmd := `free -h | grep Mem | awk '{print $3 " used / " $2 " total"}' || echo "N/A"`
	mem, err := s.RunCommand(memCmd)
	if err == nil {
		fmt.Printf("   Memory: %s\n", strings.TrimSpace(mem))
	}

	// Disk
	diskCmd := `df -h / | tail -1 | awk '{print $3 " used / " $2 " total (" $5 ")"}' || echo "N/A"`
	disk, err := s.RunCommand(diskCmd)
	if err == nil {
		fmt.Printf("   Disk: %s\n", strings.TrimSpace(disk))
	}

	return nil
}

func (s *SSHClient) KillAll(image string) error {
	fmt.Println("Killing all jobs on remote server...")

	// Stop and remove Docker containers
	fmt.Println("Stopping Docker containers...")
	stopCmd := fmt.Sprintf(`docker ps --filter "ancestor=%s" -q | xargs -r docker stop --timeout -1 || true`, image)
	_, err := s.RunCommand(stopCmd)
	if err != nil {
		return fmt.Errorf("failed to stop containers: %w", err)
	}

	rmCmd := fmt.Sprintf(`docker ps -a --filter "ancestor=%s" -q | xargs -r docker rm || true`, image)
	_, err = s.RunCommand(rmCmd)
	if err != nil {
		return fmt.Errorf("failed to remove containers: %w", err)
	}

	fmt.Println("All jobs killed.")
	return nil
}

func (s *SSHClient) KillContainer(containerID string) error {
	// Stop container gracefully, then remove it (same flow as KillAll)
	fmt.Printf("Stopping container %s gracefully...\n", containerID)
	stopCmd := fmt.Sprintf("docker stop --timeout -1 %s", containerID)
	output, err := s.RunCommand(stopCmd)
	if err != nil {
		// Check if container doesn't exist
		if strings.Contains(output, "No such container") {
			fmt.Printf("Container %s does not exist\n", containerID)
			return nil
		}
		return fmt.Errorf("failed to stop container: %w", err)
	}

	fmt.Printf("Killed container: %s\n", containerID)
	return nil
}

func (s *SSHClient) RunRemoteCommand(remotePath, image, container, command string) error {
	fmt.Println("Connected to remote server...")

	// Build Docker image
	fmt.Println("Building Docker image...")
	buildCmd := fmt.Sprintf(`cd %s && docker build -t "%s" -f %s .`, remotePath, image, dockerfilePath)
	output, err := s.RunCommand(buildCmd)
	if err != nil {
		fmt.Printf("Docker build output:\n%s\n", output)
		return fmt.Errorf("failed to build Docker image: %w", err)
	}
	fmt.Printf("Docker build completed successfully\n")

	// Run command in Docker with live output
	fmt.Println("Running command...")
	suffix := time.Now().UnixNano() % 10000 // Just last 4 digits
	containerName := fmt.Sprintf("%s-%d", container, suffix)
	dockerCmd := fmt.Sprintf(`cd %s && docker run --rm -v "%s:/app" -w /app --name "%s" "%s" bash -c "%s"`,
		remotePath, remotePath, containerName, image, command)

	// Use live output streaming for the Docker run command
	err = s.RunCommandWithLiveOutput(dockerCmd)
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	fmt.Printf("\nCommand completed successfully\n")
	return nil
}

func (s *SSHClient) PullResults(remoteRootPath, resultsPath string) error {
	// Create the corpus path on remote
	remoteResultsPath := filepath.Join(remoteRootPath, resultsPath)

	fmt.Println("Pulling results from remote server...")

	// Use rsync to pull the files
	cmd := exec.Command("rsync", "-avz",
		"-e", "ssh -F "+os.Getenv("HOME")+"/.ssh/config",
		remote+":"+remoteResultsPath+"/", resultsPath+"/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (s *SSHClient) SyncFiles(localPath, remotePath string) error {
	// For now, fall back to external rsync but with better password handling
	// We'll use SSH config which should work with our established connection

	// Create the remote directory first via SSH
	_, err := s.RunCommand(fmt.Sprintf("mkdir -p %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Use rsync but let it use our established SSH config
	cmd := exec.Command("rsync", "-avz", "--delete",
		"--exclude=.git", "--exclude=out", "--exclude=cache", "--exclude=osiris-lite",
		"-e", "ssh -F "+os.Getenv("HOME")+"/.ssh/config",
		localPath, remote+":"+remotePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (s *SSHClient) ConnectToLogs(image, containerID string) error {
	// If no container ID provided, find a running container
	if containerID == "" {
		fmt.Println("🔍 Finding running container...")
		containersCmd := fmt.Sprintf(`docker ps --filter "ancestor=%s" --format "{{.ID}}" | head -1`, image)
		output, err := s.RunCommand(containersCmd)
		if err != nil {
			return fmt.Errorf("failed to find containers: %w", err)
		}

		containerID = strings.TrimSpace(output)
		if containerID == "" {
			return fmt.Errorf("no running containers found for image %s", image)
		}
		fmt.Printf("📺 Connecting to container: %s\n", containerID)
	}

	// Connect to container logs with live streaming
	fmt.Println("Connecting to logs... (Press Ctrl+C to disconnect)")
	logsCmd := fmt.Sprintf("docker logs -f %s", containerID)

	return s.RunCommandWithLiveOutput(logsCmd)
}
