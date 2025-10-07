package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hugofrely/envswitch/internal/updater"
	"github.com/hugofrely/envswitch/internal/version"
)

var (
	cfgFile string
	verbose bool
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "envswitch",
	Short: "EnvSwitch - Manage your development environments",
	Long: `EnvSwitch is a powerful CLI tool that captures, saves and restores
the complete state of your development environments.

Think of it as snapshots for your CLI tools: when you switch from one
environment to another, EnvSwitch automatically saves the current state
(authentications, configurations, contexts) and restores the exact state
of the target environment.`,
	PersistentPreRun: checkForUpdates,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set version information
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version.Version, version.GitCommit, version.BuildDate)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.envswitch/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug mode")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home + "/.envswitch")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

// checkForUpdates is called before any command runs to check for new versions
func checkForUpdates(cmd *cobra.Command, args []string) {
	// Skip update check for certain commands
	if cmd.Name() == "update" || cmd.Name() == "version" || cmd.Name() == "completion" || cmd.Name() == "help" {
		return
	}

	// Skip if not in a terminal (e.g., piped output)
	if !isTerminal() {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return // Silently skip if we can't get home dir
	}

	configDir := home + "/.envswitch"
	if !updater.ShouldCheckForUpdate(configDir) {
		return
	}

	info, err := updater.CheckForUpdate()
	if err != nil {
		// Silently ignore update check failures
		if debug {
			fmt.Fprintf(os.Stderr, "Update check failed: %v\n", err)
		}
		return
	}

	if info.Available {
		fmt.Fprintf(os.Stderr, "\n💡 New version available: %s → %s\n", info.CurrentVersion, info.LatestVersion)
		fmt.Fprintf(os.Stderr, "   Run 'envswitch update' for update instructions\n\n")
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// test
