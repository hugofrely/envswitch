package environment

import (
	"fmt"

	"github.com/hugofrely/envswitch/pkg/plugin"
)

// SyncPluginsToEnvironments ajoute les plugins installés à tous les environnements
// avec enabled: true par défaut
func SyncPluginsToEnvironments() error {
	// Charger tous les plugins
	plugins, err := plugin.ListInstalledPlugins()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	if len(plugins) == 0 {
		return nil // Pas de plugins
	}

	// Charger tous les environnements
	environments, err := ListEnvironments()
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	// Pour chaque environnement
	for _, env := range environments {

		modified := false

		// Pour chaque plugin
		for _, p := range plugins {
			toolName := p.Metadata.ToolName

			// Vérifier si le tool existe déjà dans l'environnement
			if _, exists := env.Tools[toolName]; !exists {
				// Ajouter le tool avec enabled: true par défaut
				env.Tools[toolName] = ToolConfig{
					Enabled:      true,
					SnapshotPath: fmt.Sprintf("snapshots/%s", toolName),
				}
				modified = true
			}
		}

		// Sauvegarder si modifié
		if modified {
			if err := env.Save(); err != nil {
				return fmt.Errorf("failed to save environment %s: %w", env.Name, err)
			}
		}
	}

	return nil
}

// EnsurePluginInEnvironment s'assure qu'un plugin est présent dans un environnement
func EnsurePluginInEnvironment(env *Environment, toolName string) bool {
	if _, exists := env.Tools[toolName]; !exists {
		env.Tools[toolName] = ToolConfig{
			Enabled:      true,
			SnapshotPath: fmt.Sprintf("snapshots/%s", toolName),
		}
		return true
	}
	return false
}

// SyncPluginsOnLoad charge un environnement et synchronize les plugins
func SyncPluginsOnLoad(envName string) (*Environment, error) {
	env, err := LoadEnvironment(envName)
	if err != nil {
		return nil, err
	}

	// Charger les plugins
	plugins, err := plugin.ListInstalledPlugins()
	if err != nil {
		// Pas critique, on continue
		return env, nil
	}

	modified := false
	for _, p := range plugins {
		if EnsurePluginInEnvironment(env, p.Metadata.ToolName) {
			modified = true
		}
	}

	if modified {
		_ = env.Save() // Ignorer l'erreur, pas critique
	}

	return env, nil
}
