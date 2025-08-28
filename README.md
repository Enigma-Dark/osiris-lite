<p>
  <img src="assets/logo.png" width="150" />
</p>

# Osiris Lite

A clean, plug and play CLI tool for managing remote fuzzing jobs.

## Tool Compatibility

This tool has been **tested and validated to work with [Echidna](https://github.com/trailofbits/echidna)**. While the architecture and functionality is designed to be general-purpose by executing any "command" and may be compatible with other fuzzing tools like [Medusa](https://github.com/medusa-fuzzer/medusa), proper testing and validation for additional tools will be conducted in future releases. Examples and documentation for each fully supported tool will be provided as support is added.

## Sequential Job Execution

Osiris Lite is designed to execute fuzzing jobs **sequentially** for the same project. It does not incorporate corpus or artifact merging between runs. Each execution for a given project reuses the same artifacts from previous runs, making it ideal for isolated testing sessions or multiple sessions for different projects.

## Features

- **Flexible Configuration**: YAML/JSON/TOML config files + environment variables + command-line flags
- **SSH Integration**: Full SSH config support (ProxyJump, key-based auth, etc.)
- **Docker Orchestration**: Remote Docker container management
- **File Synchronization**: Efficient rsync-based file transfer
- **Live Monitoring**: Real-time job status and system monitoring

## Architecture

### Implementation Stack

- **Transport**: `SSH` - Uses native SSH with full config support (ProxyJump, etc.)
- **FileSync**: `Rsync` - Uses rsync over SSH for efficient file transfer
- **DockerOps**: `Docker` - Docker commands over SSH transport
- **SystemOps**: `System` - System monitoring over SSH transport

## Server Setup

**Required**: SSH server, Docker daemon, rsync, basic Unix tools.

**Quick install** (Ubuntu/Debian): `sudo apt install docker.io rsync openssh-server`

**Verify**: `docker --version && rsync --version && ssh -V`

**Docker permissions**: `sudo usermod -aG docker $USER && newgrp docker`

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/Enigma-Dark/osiris-lite.git
cd osiris-lite

# Install dependencies
make deps

# Build the tool
make build

# Install globally (optional)
make install
```

### Via Go Install

```bash
# Install latest version
go install github.com/Enigma-Dark/osiris-lite@latest

# Or install a specific version
go install github.com/Enigma-Dark/osiris-lite@v1.0.1
```

## Configuration

Osiris Lite supports **multiple configuration methods** with clear precedence:

### 1. **Config Files** (Recommended)

Create `$HOME/.osiris.yaml` or use `--config` flag:

```yaml
# ~/.osiris.yaml
remote: "my-server"
remote-path: "/home/user/project"
results-path: "./corpus"
dockerfile: "test/enigma-dark-invariants/remote/DOCKERFILE"
image: "osiris-fuzzer" # Optional, defaults to "osiris-fuzzer"
container: "osiris-runner" # Optional, defaults to "osiris-runner"
password: "" # Optional, prefer SSH keys
```

### 2. **Environment Variables**

```bash
export OSIRIS_REMOTE="my-server"
export OSIRIS_REMOTE_PATH="/home/user/project"
export OSIRIS_REMOTE_PASSWORD="your-ssh-password"  # Optional, prefer SSH keys
export OSIRIS_IMAGE="my-fuzzer"                    # Optional, defaults to "osiris-fuzzer"
export OSIRIS_CONTAINER="my-runner"                # Optional, defaults to "osiris-runner"
```

**Security Note**: Environment variables are recommended for sensitive configuration like server paths, passwords, and internal network details. This prevents accidentally exposing sensitive information in public repositories or config files that might be shared or committed to version control.

### 3. **Command-line Flags**

```bash
osiris-lite --remote my-server --remote-path /home/user/project status
```

### Configuration Precedence (highest to lowest):

1. **Command-line flags** (override everything)
2. **Environment variables** (override config file)
3. **Config file values** (override defaults)
4. **Default values**

### Default Values

- `--dockerfile`: `test/enigma-dark-invariants/remote/DOCKERFILE`
- `--image`: `osiris-fuzzer`
- `--container`: `osiris-runner`
- `--config`: `$HOME/.osiris.yaml`
- `--results-path`: No default (must be specified)
- `--remote`: No default (must be specified)
- `--remote-path`: No default (must be specified)

### SSH Configuration

Uses your existing SSH config (`~/.ssh/config`). Supports all SSH features:

- ProxyJump for complex routing
- Key-based authentication
- Compression, ciphers, etc.

Example SSH config:

```ssh
# ~/.ssh/config
Host my-server
    HostName your-server.example.com
    User your-username
    IdentityFile ~/.ssh/id_rsa
    ProxyJump bastion-host  # Optional
```

## Usage

### Global Flags

- `--config` - Config file path (default: `$HOME/.osiris.yaml`)
- `--version` - Show version information
- `--remote` - SSH host alias
- `--remote-path` - Remote working directory
- `--results-path` - Local directory for results
- `--dockerfile` - Path to Dockerfile relative to remote-path (default: `test/enigma-dark-invariants/remote/DOCKERFILE`)
- `--image` - Docker image name (default: `osiris-fuzzer`)
- `--container` - Container name (default: `osiris-runner`)
- `--password` - SSH password (prefer SSH keys)

### Commands

**Run fuzzing tests:**

```bash
# Using config file
osiris-lite run "make echidna"

# Using flags
osiris-lite run "make echidna" --remote machine --remote-path /home/user/project

# Using custom config
osiris-lite --config ./my-config.yaml run "echidna test/Contract.sol"
```

**Check job status:**

```bash
osiris-lite status
```

**Kill running jobs:**

```bash
osiris-lite kill all               # Kill everything
osiris-lite kill container_id      # Kill specific container
```

**Pull results:**

```bash
osiris-lite pull                           # Pull to configured results-path
osiris-lite pull ./local/results/         # Pull to custom path
```

**Note**: If `--results-path` is not specified and no argument is provided, the command will fail. You must either:

- Set `results-path` in your config file
- Use the `--results-path` flag
- Provide a path as an argument: `osiris-lite pull ./my-results`

**View container logs:**

```bash
osiris-lite logs                           # Connect to first running container
osiris-lite logs container_id              # Connect to specific container
```

**Note**: The `logs` command connects to running containers and streams their output in real-time. Press `Ctrl+C` to disconnect from the logs stream.

## Development

### Project Structure

```
osiris-lite/
├── cmd/                              # CLI implementation
│   ├── root.go                       # Main command with Viper config
│   ├── ssh.go                        # SSH client implementation
│   ├── run.go                        # Run command
│   ├── status.go                     # Status command
│   ├── kill.go                       # Kill command
│   └── pull.go                       # Pull command
├── build/                            # Build output
│   └── osiris-lite                   # Compiled binary
├── .git/                             # Git repository
├── .gitignore                        # Go-specific gitignore
├── Makefile                          # Build automation
├── go.mod                            # Go module definition
├── go.sum                            # Go module checksums
└── README.md                         # Documentation
```

## Dependencies

- **Cobra**: CLI framework with subcommands
- **Viper**: Configuration management (YAML/JSON/TOML + env vars)
- **SSH Config**: SSH configuration parsing
- **Go Crypto**: Native SSH client implementation

### Examples

Check the [`examples/`](examples/) directory for configuration and Docker examples to get started quickly.

### Quick Start Example

```bash
# 1. Create config file
cat > ~/.osiris.yaml << EOF
remote: "my-server"
remote-path: "/home/user/project"
results-path: "./corpus"
dockerfile: "Dockerfile"
image: "my-fuzzer"
container: "my-runner"
EOF

# 2. Run a fuzzing command
osiris-lite run "make echidna"

# 3. Check status
osiris-lite status

# 4. Check logs
osiris-lite logs <container_id> # Connect to specific container

# 5. Kill running jobs
osiris-lite kill container_id # Kill specific container
osiris-lite kill all # Kill all containers

# 6. Pull results to your local machine
osiris-lite pull
```

## Why This Architecture?

**Clean Separation of Concerns:**

- CLI layer handles user interaction and configuration
- SSH layer handles secure remote communication
- Docker layer handles container orchestration
- File sync layer handles efficient data transfer

**Maintainable:**

- Configuration-driven behavior
- Simple, focused components

**Reliable:**

- Native SSH handles complex networking (ProxyJump)
- Robust configuration with multiple sources
- Fail-fast validation
- Battle-tested underlying tools (SSH, rsync, Docker)

This is **orchestration done right** - clean Go code with flexible configuration managing battle-tested system tools.

### Configuration Security

- **Environment Variables**: Use environment variables for sensitive configuration (server paths, passwords, internal network details)
- **Public Repositories**: Environment variables prevent accidentally exposing sensitive information in public repos
- **Team Sharing**: Avoid committing config files with server details to version control
- **CI/CD**: Use environment variables in CI/CD pipelines for secure configuration

**Security Warning**: This tool currently uses `ssh.InsecureIgnoreHostKey()` which bypasses host key verification. Recommended to use with trusted hosts.

## License

GNU Affero General Public License v3.0 - see LICENSE file for details.

This license allows you to:

- ✅ Use, modify, and distribute the software
- ✅ Build upon and extend the functionality
- ❌ Use in proprietary/commercial software without open-sourcing
- 📝 Share modifications under the same license terms
