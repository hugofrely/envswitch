package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/internal/history"
)

var (
	historyLimit int
	historyAll   bool
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View environment switch history",
	Long: `View the history of environment switches.

The history shows recent switches with timestamps, success status,
duration, and any errors that occurred.

Examples:
  # Show last 10 switches (default)
  envswitch history

  # Show last 20 switches
  envswitch history --limit 20

  # Show all history
  envswitch history --all

  # Show detailed view of history
  envswitch history show

  # Clear history
  envswitch history clear`,
	RunE: runHistory,
}

var historyShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show detailed history view",
	Long:  `Show a detailed view of the switch history with all information.`,
	RunE:  runHistoryShow,
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear switch history",
	Long:  `Clear all switch history entries.`,
	RunE:  runHistoryClear,
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.AddCommand(historyShowCmd)
	historyCmd.AddCommand(historyClearCmd)

	// Add flags to main command
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 10, "Number of entries to show")
	historyCmd.Flags().BoolVar(&historyAll, "all", false, "Show all history entries")

	// Add flags to show subcommand
	historyShowCmd.Flags().IntVarP(&historyLimit, "limit", "n", 10, "Number of entries to show")
	historyShowCmd.Flags().BoolVar(&historyAll, "all", false, "Show all history entries")
}

func runHistory(cmd *cobra.Command, args []string) error {
	hist, err := history.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	if len(hist.Entries) == 0 {
		fmt.Println("No switch history found.")
		fmt.Println()
		fmt.Println("Switch between environments to build your history:")
		fmt.Println("  envswitch switch <environment>")
		return nil
	}

	// Determine how many entries to show
	limit := historyLimit
	if historyAll {
		limit = len(hist.Entries)
	}

	entries := hist.GetLast(limit)

	// Display header
	fmt.Printf("Switch History (showing %d of %d):\n", len(entries), len(hist.Entries))
	fmt.Println()

	// Display entries in reverse order (most recent first)
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		displayHistoryEntry(&entry, false)
	}

	if !historyAll && len(hist.Entries) > historyLimit {
		fmt.Printf("\nShowing last %d entries. Use --all to see all %d entries.\n", historyLimit, len(hist.Entries))
	}

	return nil
}

func runHistoryShow(cmd *cobra.Command, args []string) error {
	hist, err := history.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	if len(hist.Entries) == 0 {
		fmt.Println("No switch history found.")
		return nil
	}

	// Determine how many entries to show
	limit := historyLimit
	if historyAll {
		limit = len(hist.Entries)
	}

	entries := hist.GetLast(limit)

	fmt.Printf("Detailed Switch History (showing %d of %d):\n", len(entries), len(hist.Entries))
	fmt.Println()

	// Display entries in reverse order (most recent first)
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		displayHistoryEntry(&entry, true)
		if i > 0 {
			fmt.Println()
		}
	}

	return nil
}

func runHistoryClear(cmd *cobra.Command, args []string) error {
	hist := &history.History{
		Entries: []history.SwitchEntry{},
	}

	if err := hist.Save(); err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	fmt.Println("✅ History cleared successfully")
	return nil
}

func displayHistoryEntry(entry *history.SwitchEntry, detailed bool) {
	// Format timestamp
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")

	// Status indicator
	status := "✅"
	if !entry.Success {
		status = "❌"
	}

	// Duration
	duration := formatDuration(entry.DurationMs)

	if detailed {
		// Detailed view
		fmt.Printf("─────────────────────────────────────────────────────\n")
		fmt.Printf("Time:     %s\n", timestamp)
		fmt.Printf("Switch:   %s → %s\n", entry.From, entry.To)
		fmt.Printf("Status:   %s %s\n", status, getStatusText(entry.Success))
		fmt.Printf("Duration: %s\n", duration)

		if entry.ToolsCount > 0 {
			fmt.Printf("Tools:    %d tool(s) restored\n", entry.ToolsCount)
		}

		if entry.BackupPath != "" {
			fmt.Printf("Backup:   %s\n", entry.BackupPath)
		}

		if entry.ErrorMsg != "" {
			fmt.Printf("Error:    %s\n", entry.ErrorMsg)
		}
	} else {
		// Compact view
		fromTo := fmt.Sprintf("%s → %s", entry.From, entry.To)
		fmt.Printf("%s %s  %-30s  %s", status, timestamp, fromTo, duration)

		if entry.ErrorMsg != "" {
			fmt.Printf(" (error: %s)", truncateString(entry.ErrorMsg, 40))
		}
		fmt.Println()
	}
}

func getStatusText(success bool) string {
	if success {
		return "Success"
	}
	return "Failed"
}

func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	seconds := float64(ms) / 1000.0
	if seconds < 60 {
		return fmt.Sprintf("%.2fs", seconds)
	}
	minutes := int(seconds / 60)
	remainingSeconds := int(seconds) % 60
	return fmt.Sprintf("%dm%ds", minutes, remainingSeconds)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
