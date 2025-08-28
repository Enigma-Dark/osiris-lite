# Examples

This directory contains example files to help you get started with Osiris Lite.

## Configuration Examples

### `config.yaml`
Example configuration file for Osiris Lite. Copy this to `~/.osiris.yaml` and modify for your environment:

```bash
cp examples/config.yaml ~/.osiris.yaml
# Edit ~/.osiris.yaml with your settings
```

## Docker Examples

### `Dockerfile`
Example Dockerfile for Echidna fuzzing with Foundry integration. This creates a Docker image with:
- Echidna latest for property-based testing
- Foundry for Solidity development tools
- Make for build automation

To use:
```bash
# Copy to your project
cp examples/Dockerfile /path/to/your/project/

# Or reference it in your config
dockerfile: "examples/Dockerfile"
```

## Quick Start

1. Copy the config example:
   ```bash
   cp examples/config.yaml ~/.osiris.yaml
   ```

2. Edit the config with your settings:
   ```yaml
   remote: "your-server"
   remote-path: "/path/to/your/project"
   results-path: "./fuzzing-results"
   dockerfile: "examples/Dockerfile"
   ```

3. Run Osiris Lite:
   ```bash
   osiris-lite run "fuzz_command"
   ```
    Replace `fuzz_command` with the actual fuzz command you want to execute.