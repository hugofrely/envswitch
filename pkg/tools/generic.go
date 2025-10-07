package tools

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// GenericTool est un tool générique qui copie des fichiers de configuration
// basé sur des conventions de nommage (ex: ~/.TOOLRC pour l'outil TOOL)
type GenericTool struct {
	toolName   string
	configPath string
}

// NewGenericTool crée un tool générique pour un outil donné
func NewGenericTool(toolName, configPath string) *GenericTool {
	return &GenericTool{
		toolName:   toolName,
		configPath: configPath,
	}
}

func (g *GenericTool) Name() string {
	return g.toolName
}

func (g *GenericTool) IsInstalled() bool {
	_, err := exec.LookPath(g.toolName)
	return err == nil
}

func (g *GenericTool) Snapshot(snapshotPath string) error {
	// Créer le dossier de destination
	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Vérifier si le fichier de config existe
	if _, err := os.Stat(g.configPath); os.IsNotExist(err) {
		// Pas de config, rien à sauvegarder
		return nil
	}

	// Déterminer si c'est un fichier ou un dossier
	info, err := os.Stat(g.configPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Copier le dossier entier
		return copyDir(g.configPath, filepath.Join(snapshotPath, filepath.Base(g.configPath)))
	}

	// Copier le fichier
	return copyFile(g.configPath, filepath.Join(snapshotPath, filepath.Base(g.configPath)))
}

func (g *GenericTool) Restore(snapshotPath string) error {
	// Déterminer le nom du fichier/dossier à restaurer
	baseName := filepath.Base(g.configPath)
	sourcePath := filepath.Join(snapshotPath, baseName)

	// Vérifier si le snapshot existe
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// Pas de snapshot, rien à restaurer
		return nil
	}

	// Vérifier si c'est un fichier ou un dossier
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Supprimer le dossier et le remplacer
		os.RemoveAll(g.configPath)
		return copyDir(sourcePath, g.configPath)
	}

	// Copier le fichier
	return copyFile(sourcePath, g.configPath)
}

func (g *GenericTool) GetMetadata() (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	// Vérifier si le fichier de config existe
	if info, err := os.Stat(g.configPath); err == nil {
		metadata["config_exists"] = true
		metadata["config_path"] = g.configPath
		if info.IsDir() {
			metadata["config_type"] = "directory"
		} else {
			metadata["config_type"] = "file"
			metadata["config_size"] = info.Size()
		}
	} else {
		metadata["config_exists"] = false
	}

	return metadata, nil
}

func (g *GenericTool) ValidateSnapshot(snapshotPath string) error {
	// Vérifier que le dossier de snapshot existe
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot path does not exist: %s", snapshotPath)
	}

	return nil
}

func (g *GenericTool) Diff(snapshotPath string) ([]Change, error) {
	var changes []Change

	baseName := filepath.Base(g.configPath)
	snapshotFile := filepath.Join(snapshotPath, baseName)

	currentExists := fileExists(g.configPath)
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
		// Comparer les contenus
		if !filesEqual(g.configPath, snapshotFile) {
			changes = append(changes, Change{
				Type: ChangeTypeModified,
				Path: baseName,
			})
		}
	}

	return changes, nil
}

// Fonctions helper
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Créer le dossier parent si nécessaire
	if mkdirErr := os.MkdirAll(filepath.Dir(dst), 0755); mkdirErr != nil {
		return mkdirErr
	}

	destFile, createErr := os.Create(dst)
	if createErr != nil {
		return createErr
	}
	defer destFile.Close()

	if _, copyErr := io.Copy(destFile, sourceFile); copyErr != nil {
		return copyErr
	}

	// Copier les permissions
	sourceInfo, statErr := os.Stat(src)
	if statErr != nil {
		return statErr
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculer le chemin relatif
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath)
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func filesEqual(file1, file2 string) bool {
	content1, err1 := os.ReadFile(file1)
	content2, err2 := os.ReadFile(file2)

	if err1 != nil || err2 != nil {
		return false
	}

	return bytes.Equal(content1, content2)
}
