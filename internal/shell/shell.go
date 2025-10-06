package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hugofrely/envswitch/internal/config"
)

// GenerateInitScript generates the shell initialization script for the specified shell
func GenerateInitScript(shellType string, cfg *config.Config) (string, error) {
	if !cfg.EnablePromptIntegration {
		return "# Prompt integration is disabled in config\n", nil
	}

	switch shellType {
	case "bash":
		return generateBashScript(cfg)
	case "zsh":
		return generateZshScript(cfg)
	case "fish":
		return generateFishScript(cfg)
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellType)
	}
}

// InstallShellIntegration automatically installs shell integration
func InstallShellIntegration(shellType string, cfg *config.Config) (string, error) {
	// Generate the script
	script, err := GenerateInitScript(shellType, cfg)
	if err != nil {
		return "", err
	}

	// Determine config file path
	configFile, err := getShellConfigFile(shellType)
	if err != nil {
		return "", err
	}

	// Check if already installed
	if isAlreadyInstalled(configFile) {
		return configFile, fmt.Errorf("shell integration already installed in %s", configFile)
	}

	// Append to config file
	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Add marker and script
	marker := "\n# envswitch shell integration - DO NOT EDIT THIS SECTION\n"
	endMarker := "# end envswitch shell integration\n"

	if _, err := file.WriteString(marker); err != nil {
		return "", fmt.Errorf("failed to write marker: %w", err)
	}

	if _, err := file.WriteString(script); err != nil {
		return "", fmt.Errorf("failed to write script: %w", err)
	}

	if _, err := file.WriteString(endMarker); err != nil {
		return "", fmt.Errorf("failed to write end marker: %w", err)
	}

	return configFile, nil
}

// getShellConfigFile returns the path to the shell configuration file
func getShellConfigFile(shellType string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	switch shellType {
	case "bash":
		// Prefer .bashrc, fallback to .bash_profile
		bashrc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc, nil
		}
		return filepath.Join(home, ".bash_profile"), nil
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "fish":
		configDir := filepath.Join(home, ".config", "fish")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create fish config directory: %w", err)
		}
		return filepath.Join(configDir, "config.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellType)
	}
}

// isAlreadyInstalled checks if envswitch integration is already in the config file
func isAlreadyInstalled(configFile string) bool {
	file, err := os.Open(configFile)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "envswitch shell integration") {
			return true
		}
	}

	return false
}

// generateBashScript generates the bash initialization script
func generateBashScript(cfg *config.Config) (string, error) {
	tmpl := `# envswitch prompt integration for bash
__envswitch_prompt() {
    local env_name=$(cat ~/.envswitch/current.lock 2>/dev/null)
    if [ -n "$env_name" ]; then
        {{if .Color}}printf "\033[{{.Color}}m"{{end}}
        printf "{{.Format}}" "$env_name"
        {{if .Color}}printf "\033[0m"{{end}}
    fi
}

# Add envswitch to PS1
if [[ "$PS1" != *__envswitch_prompt* ]]; then
    export PS1="$(__envswitch_prompt)$PS1"
fi

# Auto-load environment variables on switch
__envswitch_load_vars() {
    local env_name=$(cat ~/.envswitch/current.lock 2>/dev/null)
    if [ -n "$env_name" ]; then
        local env_file="$HOME/.envswitch/environments/$env_name/snapshots/env-vars.env"
        if [ -f "$env_file" ]; then
            while IFS='=' read -r key value; do
                # Skip comments and empty lines
                [[ "$key" =~ ^#.*$ ]] && continue
                [[ -z "$key" ]] && continue
                # Export the variable
                export "$key=$value"
            done < "$env_file"
        fi
    fi
}
`

	data := struct {
		Format string
		Color  string
	}{
		Format: parsePromptFormat(cfg.PromptFormat),
		Color:  parsePromptColor(cfg.PromptColor),
	}

	t, err := template.New("bash").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateZshScript generates the zsh initialization script
func generateZshScript(cfg *config.Config) (string, error) {
	// Build the script manually to avoid template parsing issues with zsh color syntax
	var script strings.Builder

	script.WriteString("# envswitch prompt integration for zsh\n")
	script.WriteString("setopt PROMPT_SUBST\n\n")
	script.WriteString("__envswitch_prompt() {\n")
	script.WriteString("    local env_name=$(cat ~/.envswitch/current.lock 2>/dev/null)\n")
	script.WriteString("    if [[ -n \"$env_name\" ]]; then\n")

	color := parseZshColor(cfg.PromptColor)
	if color != "" {
		script.WriteString(fmt.Sprintf("        printf \"%%F{%s}\"\n", color))
	}

	format := parsePromptFormat(cfg.PromptFormat)
	script.WriteString(fmt.Sprintf("        printf %q \"$env_name\"\n", format))

	if color != "" {
		script.WriteString("        printf \"%f\"\n")
	}

	script.WriteString("    fi\n")
	script.WriteString("}\n\n")
	script.WriteString("# Add envswitch to PROMPT\n")
	script.WriteString("if [[ \"$PROMPT\" != *__envswitch_prompt* ]]; then\n")
	script.WriteString("    export PROMPT=\"$(__envswitch_prompt)$PROMPT\"\n")
	script.WriteString("fi\n\n")
	script.WriteString("# Auto-load environment variables on switch\n")
	script.WriteString("__envswitch_load_vars() {\n")
	script.WriteString("    local env_name=$(cat ~/.envswitch/current.lock 2>/dev/null)\n")
	script.WriteString("    if [[ -n \"$env_name\" ]]; then\n")
	script.WriteString("        local env_file=\"$HOME/.envswitch/environments/$env_name/snapshots/env-vars.env\"\n")
	script.WriteString("        if [[ -f \"$env_file\" ]]; then\n")
	script.WriteString("            while IFS='=' read -r key value; do\n")
	script.WriteString("                # Skip comments and empty lines\n")
	script.WriteString("                [[ \"$key\" =~ ^#.*$ ]] && continue\n")
	script.WriteString("                [[ -z \"$key\" ]] && continue\n")
	script.WriteString("                # Export the variable\n")
	script.WriteString("                export \"$key=$value\"\n")
	script.WriteString("            done < \"$env_file\"\n")
	script.WriteString("        fi\n")
	script.WriteString("    fi\n")
	script.WriteString("}\n")

	return script.String(), nil
}

// generateFishScript generates the fish initialization script
func generateFishScript(cfg *config.Config) (string, error) {
	tmpl := `# envswitch prompt integration for fish
function __envswitch_prompt
    set -l env_name (cat ~/.envswitch/current.lock 2>/dev/null)
    if test -n "$env_name"
        {{if .Color}}set_color {{.Color}}{{end}}
        printf "{{.Format}}" "$env_name"
        {{if .Color}}set_color normal{{end}}
    end
end

# Add envswitch to fish_prompt
function fish_prompt
    echo -n (__envswitch_prompt)
    # Your original prompt here
end

# Auto-load environment variables on switch
function __envswitch_load_vars
    set -l env_name (cat ~/.envswitch/current.lock 2>/dev/null)
    if test -n "$env_name"
        set -l env_file "$HOME/.envswitch/environments/$env_name/snapshots/env-vars.env"
        if test -f "$env_file"
            while read -l line
                # Skip comments and empty lines
                if string match -qr '^#' "$line"; or test -z "$line"
                    continue
                end
                # Parse KEY=VALUE
                set -l parts (string split -m 1 '=' $line)
                if test (count $parts) -eq 2
                    set -gx $parts[1] $parts[2]
                end
            end < "$env_file"
        end
    end
end
`

	data := struct {
		Format string
		Color  string
	}{
		Format: parsePromptFormat(cfg.PromptFormat),
		Color:  parseFishColor(cfg.PromptColor),
	}

	t, err := template.New("fish").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// parsePromptFormat converts the config prompt format to shell-compatible format
func parsePromptFormat(format string) string {
	if format == "" {
		return "(%s) "
	}
	// Replace {env} with %s for printf
	return strings.ReplaceAll(format, "{env}", "%s")
}

// parsePromptColor converts color names to ANSI escape codes for bash
func parsePromptColor(color string) string {
	colors := map[string]string{
		"black":   "30",
		"red":     "31",
		"green":   "32",
		"yellow":  "33",
		"blue":    "34",
		"magenta": "35",
		"cyan":    "36",
		"white":   "37",
		"default": "",
	}

	if code, ok := colors[color]; ok {
		return code
	}
	return ""
}

// parseZshColor converts color names to zsh color codes
func parseZshColor(color string) string {
	// zsh uses color names directly
	if color == "" || color == "default" {
		return ""
	}
	return color
}

// parseFishColor converts color names to fish color names
func parseFishColor(color string) string {
	// fish uses color names directly
	if color == "" || color == "default" {
		return "normal"
	}
	return color
}
