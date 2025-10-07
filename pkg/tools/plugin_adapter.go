package tools

import (
	"fmt"

	"github.com/hugofrely/envswitch/pkg/plugin"
)

// PluginAdapter adapte un Plugin pour qu'il implémente l'interface Tool
type PluginAdapter struct {
	plugin plugin.Plugin
}

// NewPluginAdapter crée un adaptateur pour un plugin
func NewPluginAdapter(p plugin.Plugin) *PluginAdapter {
	return &PluginAdapter{
		plugin: p,
	}
}

func (p *PluginAdapter) Name() string {
	return p.plugin.Name()
}

func (p *PluginAdapter) IsInstalled() bool {
	return p.plugin.IsInstalled()
}

func (p *PluginAdapter) Snapshot(snapshotPath string) error {
	return p.plugin.Snapshot(snapshotPath)
}

func (p *PluginAdapter) Restore(snapshotPath string) error {
	return p.plugin.Restore(snapshotPath)
}

func (p *PluginAdapter) GetMetadata() (map[string]interface{}, error) {
	return p.plugin.GetMetadata()
}

func (p *PluginAdapter) ValidateSnapshot(snapshotPath string) error {
	return p.plugin.Validate(snapshotPath)
}

// Diff implémente une différence basique pour les plugins
// Les plugins n'implémentent pas forcément Diff, donc on retourne une implémentation simple
func (p *PluginAdapter) Diff(snapshotPath string) ([]Change, error) {
	// Pour l'instant, on ne peut pas faire de diff détaillé sans que le plugin l'implémente
	// On retourne juste si le snapshot existe ou pas
	var changes []Change

	// Vérifier si le snapshot a changé en comparant les métadonnées
	currentMeta, err := p.plugin.GetMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get current metadata: %w", err)
	}

	// Si on ne peut pas obtenir les métadonnées, on ne peut pas faire de diff
	if len(currentMeta) == 0 {
		return changes, nil
	}

	// Pour une implémentation basique, on signale juste qu'il y a potentiellement des changements
	// Un vrai diff nécessiterait que le plugin expose plus d'informations
	return changes, nil
}
