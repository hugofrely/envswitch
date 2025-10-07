package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// MultiPathTool gère plusieurs fichiers/dossiers de configuration
type MultiPathTool struct {
	toolName    string
	configPaths []string
}

// NewMultiPathTool crée un tool qui gère plusieurs chemins
func NewMultiPathTool(toolName string, configPaths []string) *MultiPathTool {
	return &MultiPathTool{
		toolName:    toolName,
		configPaths: configPaths,
	}
}

func (m *MultiPathTool) Name() string {
	return m.toolName
}

func (m *MultiPathTool) IsInstalled() bool {
	// Considérer installé si au moins un fichier existe
	for _, path := range m.configPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func (m *MultiPathTool) Snapshot(snapshotPath string) error {
	// Créer le dossier de destination
	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Copier chaque fichier/dossier
	for _, configPath := range m.configPaths {
		// Vérifier si le fichier existe
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Fichier n'existe pas, on continue
			continue
		}

		// Déterminer si c'est un fichier ou un dossier
		info, err := os.Stat(configPath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", configPath, err)
		}

		baseName := filepath.Base(configPath)
		destPath := filepath.Join(snapshotPath, baseName)

		if info.IsDir() {
			// Copier le dossier entier
			if err := copyDir(configPath, destPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", configPath, err)
			}
		} else {
			// Copier le fichier
			if err := copyFile(configPath, destPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", configPath, err)
			}
		}
	}

	return nil
}

func (m *MultiPathTool) Restore(snapshotPath string) error {
	// Restaurer chaque fichier/dossier
	for _, configPath := range m.configPaths {
		baseName := filepath.Base(configPath)
		sourcePath := filepath.Join(snapshotPath, baseName)

		// Vérifier si le snapshot existe
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			// Pas de snapshot pour ce fichier, on continue
			continue
		}

		// Déterminer si c'est un fichier ou un dossier
		info, err := os.Stat(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to stat snapshot %s: %w", sourcePath, err)
		}

		if info.IsDir() {
			// Supprimer le dossier et le remplacer
			os.RemoveAll(configPath)
			if err := copyDir(sourcePath, configPath); err != nil {
				return fmt.Errorf("failed to restore directory %s: %w", configPath, err)
			}
		} else {
			// Copier le fichier
			if err := copyFile(sourcePath, configPath); err != nil {
				return fmt.Errorf("failed to restore file %s: %w", configPath, err)
			}
		}
	}

	return nil
}

func (m *MultiPathTool) GetMetadata() (map[string]interface{}, error) {
	metadata := make(map[string]interface{})
	metadata["config_paths"] = m.configPaths

	existingPaths := []string{}
	for _, path := range m.configPaths {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}
	metadata["existing_paths"] = existingPaths
	metadata["path_count"] = len(m.configPaths)

	return metadata, nil
}

func (m *MultiPathTool) ValidateSnapshot(snapshotPath string) error {
	// Vérifier que le dossier de snapshot existe
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot path does not exist: %s", snapshotPath)
	}

	return nil
}

func (m *MultiPathTool) Diff(snapshotPath string) ([]Change, error) {
	var changes []Change

	for _, configPath := range m.configPaths {
		baseName := filepath.Base(configPath)
		snapshotFile := filepath.Join(snapshotPath, baseName)

		currentExists := fileExists(configPath)
		snapshotExists := fileExists(snapshotFile)

		if snapshotExists && !currentExists {
			changes = append(changes, Change{
				Type: ChangeTypeRemoved,
				Path: baseName,
			})
		} else if !snapshotExists && currentExists {
			changes = append(changes, Change{
				Type: ChangeTypeAdded,
				Path: baseName,
			})
		} else if snapshotExists && currentExists {
			// Comparer les contenus (simple check, pas de diff profond)
			if !filesEqual(configPath, snapshotFile) {
				changes = append(changes, Change{
					Type: ChangeTypeModified,
					Path: baseName,
				})
			}
		}
	}

	return changes, nil
}
