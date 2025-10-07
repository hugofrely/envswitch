# Exemples EnvSwitch

Ce dossier contient des exemples complets pour utiliser EnvSwitch.

## Plugins

### NPM Plugin

Plugin complet pour gérer les configurations NPM.

**Location:** `npm-plugin-example/`

**Contenu:**
- `plugin.yaml` - Manifest du plugin
- `main.go` - Code complet du plugin (170 lignes)
- `go.mod` - Dépendances Go
- `README.md` - Documentation complète
- `.gitignore`

**Tester:**
```bash
cd npm-plugin-example
go build
./npm-plugin
```

**Installer:**
```bash
envswitch plugin install ./npm-plugin-example
```

**Documentation:** Voir [docs/PLUGINS.md](../docs/PLUGINS.md)

---

## Utilisation

Chaque exemple peut être utilisé tel quel ou comme base pour créer vos propres plugins.

### Structure typique d'un plugin

```
my-plugin/
├── plugin.yaml      # Manifest (obligatoire)
├── main.go         # Code du plugin (optionnel)
├── go.mod          # Dépendances Go (si code Go)
├── README.md       # Documentation
└── .gitignore
```

### Créer un nouveau plugin

1. Copier un exemple comme base:
   ```bash
   cp -r npm-plugin-example my-plugin
   ```

2. Modifier les fichiers:
   - `plugin.yaml` - Changer le nom, version, tool_name
   - `main.go` - Adapter la logique pour votre outil
   - `README.md` - Documentation

3. Tester:
   ```bash
   cd my-plugin
   go build
   ./my-plugin
   ```

4. Installer:
   ```bash
   envswitch plugin install .
   ```

---

## Documentation

- **Guide complet**: [docs/PLUGINS.md](../docs/PLUGINS.md)
- **README principal**: [README.md](../README.md)
