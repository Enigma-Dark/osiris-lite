package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	remote         string
	remotePath     string
	resultsPath    string
	dockerfilePath string
	password       string
	image          string
	container      = "osiris-runner"

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:     "osiris-lite",
		Short:   "Remote fuzzing workflow management",
		Version: "1.0.1", // Change this to the version of the CLI on release
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	// Global config file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.osiris.yaml)")

	// Version flag
	rootCmd.PersistentFlags().Bool("version", false, "Show version information")

	// Existing flags
	rootCmd.PersistentFlags().StringVarP(&remote, "remote", "r", "", "Remote server (SSH config alias)")
	rootCmd.PersistentFlags().StringVar(&remotePath, "remote-path", "", "Remote working directory")
	rootCmd.PersistentFlags().StringVar(&resultsPath, "results-path", "", "Local directory for pulling results")
	rootCmd.PersistentFlags().StringVarP(&dockerfilePath, "dockerfile", "d", "test/enigma-dark-invariants/remote/DOCKERFILE", "Path to Dockerfile relative to remote-path")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password for SSH authentication (optional)")
	rootCmd.PersistentFlags().StringVar(&image, "image", "osiris-fuzzer", "Docker image name")
	rootCmd.PersistentFlags().StringVar(&container, "container", "osiris-runner", "Container name")

	// Bind flags to viper
	viper.BindPFlag("remote", rootCmd.PersistentFlags().Lookup("remote"))
	viper.BindPFlag("remote-path", rootCmd.PersistentFlags().Lookup("remote-path"))
	viper.BindPFlag("results-path", rootCmd.PersistentFlags().Lookup("results-path"))
	viper.BindPFlag("dockerfile", rootCmd.PersistentFlags().Lookup("dockerfile"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("image", rootCmd.PersistentFlags().Lookup("image"))
	viper.BindPFlag("container", rootCmd.PersistentFlags().Lookup("container"))

	// Add subcommands
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "run [command]",
			Short: "Run command in Docker on remote",
			Args:  cobra.MinimumNArgs(1),
			RunE:  runCommand,
		},
		&cobra.Command{
			Use:   "status",
			Short: "Check active jobs",
			RunE:  statusCommand,
		},
		&cobra.Command{
			Use:   "kill [container_id|all]",
			Short: "Kill jobs",
			RunE:  killCommand,
		},
		&cobra.Command{
			Use:   "pull [optional_path]",
			Short: "Pull results",
			RunE:  pullCommand,
		},
		&cobra.Command{
			Use:   "logs [container_id]",
			Short: "Connect to container logs",
			RunE:  logsCommand,
		},
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Load .env file if it exists
	_ = godotenv.Load()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".osiris" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".osiris")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Bind specific environment variables
	viper.BindEnv("remote", "OSIRIS_REMOTE")
	viper.BindEnv("password", "OSIRIS_REMOTE_PASSWORD")
	viper.BindEnv("remote-path", "OSIRIS_REMOTE_PATH")
	viper.BindEnv("image", "OSIRIS_IMAGE")
	viper.BindEnv("container", "OSIRIS_CONTAINER")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// Update variables from viper config
	if viper.IsSet("remote") {
		remote = viper.GetString("remote")
	}
	if viper.IsSet("remote-path") {
		remotePath = viper.GetString("remote-path")
	}
	if viper.IsSet("results-path") {
		resultsPath = viper.GetString("results-path")
	}
	if viper.IsSet("dockerfile") {
		dockerfilePath = viper.GetString("dockerfile")
	}
	if viper.IsSet("password") {
		password = viper.GetString("password")
	}
	if viper.IsSet("image") {
		image = viper.GetString("image")
	}
	if viper.IsSet("container") {
		container = viper.GetString("container")
	}
}

func Execute() error {
	return rootCmd.Execute()
}
