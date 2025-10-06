package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hugofrely/envswitch/pkg/environment"
)

// ExecuteHooks executes a list of hooks
func ExecuteHooks(hooks []environment.Hook, envName string) error {
	for i, hook := range hooks {
		if err := executeHook(hook, envName, i+1, len(hooks)); err != nil {
			return err
		}
	}
	return nil
}

// executeHook executes a single hook
func executeHook(hook environment.Hook, envName string, index, total int) error {
	description := hook.Description
	if description == "" {
		if hook.Command != "" {
			description = hook.Command
		} else {
			description = "custom script"
		}
	}

	fmt.Printf("  Running hook %d/%d: %s\n", index, total, description)

	var cmd *exec.Cmd
	if hook.Command != "" {
		// Execute as shell command
		// #nosec G204 - Command execution from trusted user configuration is intentional
		cmd = exec.Command("sh", "-c", hook.Command)
	} else if hook.Script != "" {
		// Execute as inline script
		// #nosec G204 - Script execution from trusted user configuration is intentional
		cmd = exec.Command("sh", "-c", hook.Script)
	} else {
		return fmt.Errorf("hook has neither command nor script")
	}

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ENVSWITCH_ENV=%s", envName),
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("    ✗ Hook failed: %v\n", err)
		if len(output) > 0 {
			fmt.Printf("    Output: %s\n", strings.TrimSpace(string(output)))
		}
		return fmt.Errorf("hook failed: %w", err)
	}

	if hook.Verify {
		fmt.Printf("    ✓ Verified\n")
	} else {
		fmt.Printf("    ✓ Completed\n")
	}

	return nil
}
