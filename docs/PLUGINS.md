# EnvSwitch Plugin System

Les plugins permettent d'ajouter le support de nouveaux outils à EnvSwitch (en plus de gcloud, kubectl, aws, docker, git).

## Table des matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture d'un plugin](#architecture-dun-plugin)
- [Créer un plugin simple](#créer-un-plugin-simple)
- [Exemple complet: Plugin NPM](#exemple-complet-plugin-npm)
- [Installer et gérer les plugins](#installer-et-gérer-les-plugins)
- [Structure du manifest](#structure-du-manifest)
- [Tester votre plugin](#tester-votre-plugin)

---

## Vue d'ensemble

Un plugin EnvSwitch est simplement:

1. Un dossier avec un fichier `plugin.yaml`
2. Le fichier décrit quel outil le plugin supporte
3. Le plugin capture et restaure la configuration de l'outil

**Où sont stockés les plugins:**
```
~/.envswitch/plugins/
├── npm/
│   └── plugin.yaml
├── terraform/
│   └── plugin.yaml
└── ansible/
    └── plugin.yaml
```

---

## Architecture d'un plugin

### Plugin simple (YAML seulement)

```
npm-plugin/
├── plugin.yaml          # Manifest (obligatoire)
├── README.md           # Documentation
└── LICENSE             # Licence (optionnel)
```

### Plugin avancé (avec code Go)

```
terraform-plugin/
├── plugin.yaml         # Manifest (obligatoire)
├── main.go            # Point d'entrée
├── terraform.go       # Implémentation du plugin
├── go.mod             # Dépendances Go
├── go.sum
├── README.md          # Documentation
├── LICENSE
└── examples/          # Exemples d'usage
    └── example.yaml
```

### Architecture détaillée d'un plugin complet

```
my-tool-plugin/
│
├── plugin.yaml              # Manifest principal
│
├── src/                     # Code source (si plugin avancé)
│   ├── main.go             # Entry point
│   ├── plugin.go           # Implémentation Plugin interface
│   ├── snapshot.go         # Logique de snapshot
│   ├── restore.go          # Logique de restore
│   └── utils.go            # Fonctions helpers
│
├── tests/                   # Tests
│   ├── plugin_test.go
│   └── integration_test.go
│
├── examples/                # Exemples et templates
│   ├── basic/
│   │   └── config.yaml
│   └── advanced/
│       └── config.yaml
│
├── docs/                    # Documentation
│   ├── USAGE.md
│   └── TROUBLESHOOTING.md
│
├── scripts/                 # Scripts utilitaires
│   ├── install.sh
│   └── test.sh
│
├── go.mod                   # Si plugin Go
├── go.sum
├── Makefile                # Commandes de build
├── README.md               # Documentation principale
├── LICENSE
└── .gitignore
```

### Structure minimale recommandée

Pour démarrer rapidement:

```
my-plugin/
├── plugin.yaml       # Obligatoire
└── README.md        # Recommandé
```

---

## Créer un plugin simple

### Étape 1: Créer le dossier

```bash
mkdir npm-plugin
cd npm-plugin
```

### Étape 2: Créer le manifest `plugin.yaml`

```yaml
metadata:
  name: npm
  version: 1.0.0
  description: NPM registry and authentication
  author: Votre Nom
  tool_name: npm
```

### Étape 3: Créer le README.md

```markdown
# NPM Plugin for EnvSwitch

Capture et restaure la configuration NPM.

## Installation

envswitch plugin install ./npm-plugin

## Ce qui est capturé

- ~/.npmrc (configuration et auth)
```

C'est tout! Le plugin est prêt à être installé.

---

## Exemple complet: Plugin NPM

Créons un plugin qui capture la configuration NPM (registry, authentification, etc.)

### Structure du projet

```
npm-plugin/
├── plugin.yaml
└── README.md
```

### plugin.yaml

```yaml
metadata:
  name: npm
  version: 1.0.0
  description: NPM registry and authentication management
  author: EnvSwitch Community
  homepage: https://github.com/envswitch/npm-plugin
  license: MIT
  tool_name: npm
  tags:
    - npm
    - nodejs
    - registry
```

### README.md

```markdown
# EnvSwitch NPM Plugin

Plugin pour gérer les configurations NPM avec EnvSwitch.

## Installation

git clone https://github.com/username/npm-plugin
envswitch plugin install ./npm-plugin

## Ce qui est capturé

- `~/.npmrc` - Configuration NPM globale
- Registry URL
- Authentification tokens
- Proxy settings

## Usage

### Créer des environnements

bash
# Environnement travail
npm config set registry https://npm.company.com
npm login
envswitch create work --from-current

# Environnement personnel
npm config set registry https://registry.npmjs.org/
npm login
envswitch create personal --from-current


### Switcher entre environnements

bash
envswitch switch work      # Registry d'entreprise
envswitch switch personal  # Registry public


## Prérequis

- EnvSwitch v0.1.0+
- NPM installé

## Support

Ouvrez une issue sur: https://github.com/username/npm-plugin/issues
```

### Ce que le plugin capture

Le plugin NPM capture automatiquement:

- `~/.npmrc` - Configuration NPM globale
- Registry URL
- Authentification tokens
- Proxy settings
- Configuration custom

### Exemple de configuration NPM

Voici ce que contient typiquement un `.npmrc`:

```ini
registry=https://registry.npmjs.org/
//registry.npmjs.org/:_authToken=npm_xxxxxxxxxxxxx
@mycompany:registry=https://npm.mycompany.com/
//npm.mycompany.com/:_authToken=xxxxxxxxxxxxx
```

### Comment ça fonctionne

Quand vous faites `envswitch switch`:

1. **Snapshot**: EnvSwitch copie `~/.npmrc` dans l'environnement actuel
2. **Restore**: EnvSwitch restaure le `~/.npmrc` de l'environnement cible

### Cas d'usage réels

**Scénario 1: Travail vs Personnel**

```bash
# Environnement travail
cat ~/.npmrc
# registry=https://npm.company.com
# //npm.company.com/:_authToken=work_token

envswitch create work --from-current

# Environnement personnel
npm config set registry https://registry.npmjs.org/
npm login  # Avec vos credentials personnels

envswitch create personal --from-current

# Maintenant vous pouvez switcher instantanément:
envswitch switch work      # Registry d'entreprise + auth
envswitch switch personal  # Registry public + auth perso
```

**Scénario 2: Multiple clients**

```bash
# Client A
npm config set registry https://npm.clientA.com
npm config set //npm.clientA.com/:_authToken token_A
envswitch create clientA --from-current

# Client B
npm config set registry https://npm.clientB.com
npm config set //npm.clientB.com/:_authToken token_B
envswitch create clientB --from-current

# Switch facilement entre les clients
envswitch switch clientA
envswitch switch clientB
```

---

## Installer et gérer les plugins

### Lister les plugins installés

```bash
envswitch plugin list
```

Exemple de sortie:
```
Installed plugins:

  • npm v1.0.0
    NPM registry and authentication management
    Tool: npm

  • terraform v1.0.0
    Terraform workspace management
    Tool: terraform

Total: 2 plugin(s)
```

### Installer un plugin

```bash
# Depuis un dossier local
envswitch plugin install ./npm-plugin

# Depuis un repo Git
git clone https://github.com/user/npm-plugin
envswitch plugin install ./npm-plugin
```

Sortie:
```
✅ Plugin 'npm' v1.0.0 installed successfully
   NPM registry and authentication management
```

### Voir les infos d'un plugin

```bash
envswitch plugin info npm
```

Sortie:
```
Plugin: npm
Version: 1.0.0
Description: NPM registry and authentication management
Author: EnvSwitch Community
Homepage: https://github.com/envswitch/npm-plugin
License: MIT
Tool: npm
Tags: [npm nodejs registry]
```

### Supprimer un plugin

```bash
envswitch plugin remove npm
```

---

## Structure du manifest

Le fichier `plugin.yaml` contient les métadonnées de votre plugin.

### Champs obligatoires

```yaml
metadata:
  name: npm              # Nom unique (lowercase, tirets)
  version: 1.0.0         # Version sémantique
  tool_name: npm         # Nom de l'outil
```

### Champs optionnels

```yaml
metadata:
  description: Description courte du plugin
  author: Votre Nom
  homepage: https://github.com/user/plugin
  license: MIT
  tags:
    - npm
    - nodejs
```

### Exemples de manifests

**Plugin Terraform:**
```yaml
metadata:
  name: terraform
  version: 1.0.0
  description: Terraform workspace and state management
  tool_name: terraform
  author: Community
  tags:
    - terraform
    - iac
```

**Plugin Ansible:**
```yaml
metadata:
  name: ansible
  version: 1.0.0
  description: Ansible inventory and vault management
  tool_name: ansible
  author: Community
  tags:
    - ansible
    - automation
```

**Plugin Helm:**
```yaml
metadata:
  name: helm
  version: 1.0.0
  description: Helm repositories and configuration
  tool_name: helm
  author: Community
  tags:
    - helm
    - kubernetes
```

---

## Tester votre plugin

### Test manuel rapide

1. **Créer un environnement test:**
```bash
envswitch create test-npm --empty
```

2. **Configurer NPM:**
```bash
npm config set registry https://test-registry.com
echo "Test config" >> ~/.npmrc
```

3. **Activer le plugin:**

Éditer `~/.envswitch/environments/test-npm/metadata.yaml`:
```yaml
tools:
  npm:
    enabled: true
    snapshot_path: ""
```

4. **Tester le snapshot:**
```bash
envswitch switch test-npm
```

5. **Vérifier le snapshot:**
```bash
ls ~/.envswitch/environments/test-npm/snapshots/npm/
cat ~/.envswitch/environments/test-npm/snapshots/npm/.npmrc
```

6. **Tester la restauration:**
```bash
# Modifier la config
npm config set registry https://different-registry.com

# Switcher pour restaurer
envswitch switch test-npm

# Vérifier que c'est restauré
npm config get registry
# Devrait afficher: https://test-registry.com
```

---

## Fichiers capturés par outil

Voici ce que les plugins typiques capturent:

### NPM
- `~/.npmrc` - Configuration globale

### Yarn
- `~/.yarnrc` - Configuration Yarn 1.x
- `~/.yarnrc.yml` - Configuration Yarn 2+

### Python/Pip
- `~/.pip/pip.conf` - Configuration pip
- `~/.pypirc` - Credentials PyPI

### Terraform
- `~/.terraform.d/` - Configuration CLI
- `~/.terraformrc` - Credentials

### Ansible
- `~/.ansible.cfg` - Configuration
- `~/.ansible/` - Collections, plugins

### Helm
- `~/.config/helm/` - Configuration
- `~/.cache/helm/` - Repository cache

### Git
- `~/.gitconfig` - Configuration globale
- `~/.git-credentials` - Credentials

### Docker
- `~/.docker/config.json` - Configuration et auth

### Maven
- `~/.m2/settings.xml` - Configuration et repositories

### Gradle
- `~/.gradle/gradle.properties` - Configuration

---

## Créer un plugin avancé (optionnel)

Si vous avez besoin de logique custom (par exemple, exécuter des commandes), vous pouvez créer un plugin en Go.

### Structure complète avec Go

```
npm-plugin/
├── plugin.yaml           # Manifest
├── main.go              # Entry point
├── plugin.go            # Implémentation
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── tests/
    └── plugin_test.go
```

### Exemple minimal en Go

**main.go:**
```go
package main

import (
    "os"
    "path/filepath"
    "os/exec"
)

type NPMPlugin struct {
    ConfigFile string // ~/.npmrc
}

func NewNPMPlugin() *NPMPlugin {
    home, _ := os.UserHomeDir()
    return &NPMPlugin{
        ConfigFile: filepath.Join(home, ".npmrc"),
    }
}

func (n *NPMPlugin) Name() string {
    return "npm"
}

func (n *NPMPlugin) IsInstalled() bool {
    _, err := exec.LookPath("npm")
    return err == nil
}

// Snapshot copie ~/.npmrc vers le snapshot
func (n *NPMPlugin) Snapshot(destPath string) error {
    os.MkdirAll(destPath, 0755)

    // Copier .npmrc
    data, err := os.ReadFile(n.ConfigFile)
    if err != nil {
        return err
    }

    return os.WriteFile(
        filepath.Join(destPath, ".npmrc"),
        data,
        0644,
    )
}

// Restore restaure ~/.npmrc depuis le snapshot
func (n *NPMPlugin) Restore(sourcePath string) error {
    snapshotFile := filepath.Join(sourcePath, ".npmrc")

    data, err := os.ReadFile(snapshotFile)
    if err != nil {
        return err
    }

    return os.WriteFile(n.ConfigFile, data, 0644)
}
```

**go.mod:**
```go
module github.com/username/npm-plugin

go 1.21
```

Mais dans la plupart des cas, un simple `plugin.yaml` suffit!

---

## Exemples de structure de projets réels

### 1. Plugin simple (NPM)

```
npm-plugin/
├── plugin.yaml
└── README.md
```

**Usage:** Capture automatique de `~/.npmrc`

### 2. Plugin moyen (Terraform)

```
terraform-plugin/
├── plugin.yaml
├── README.md
├── LICENSE
└── examples/
    └── terraform.tfvars.example
```

**Usage:** Capture `~/.terraform.d/` + doc détaillée

### 3. Plugin avancé (Custom Tool)

```
custom-plugin/
├── plugin.yaml
├── main.go
├── plugin.go
├── snapshot.go
├── restore.go
├── utils.go
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── Makefile
├── tests/
│   ├── plugin_test.go
│   └── integration_test.go
├── examples/
│   ├── basic.yaml
│   └── advanced.yaml
└── docs/
    ├── USAGE.md
    └── TROUBLESHOOTING.md
```

**Usage:** Logique custom en Go pour outils complexes

---

## Partager votre plugin

### Sur GitHub

1. **Créer un repo:**
```bash
mkdir npm-plugin
cd npm-plugin

# Créer les fichiers
cat > plugin.yaml << 'EOF'
metadata:
  name: npm
  version: 1.0.0
  description: NPM registry and authentication
  tool_name: npm
  author: Your Name
  homepage: https://github.com/yourusername/npm-plugin
  license: MIT
EOF

cat > README.md << 'EOF'
# NPM Plugin for EnvSwitch

Capture et restaure la configuration NPM.

## Installation
git clone https://github.com/yourusername/npm-plugin
envswitch plugin install ./npm-plugin
EOF

git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/yourusername/npm-plugin
git push -u origin main
```

2. **Les utilisateurs peuvent installer:**
```bash
git clone https://github.com/yourusername/npm-plugin
envswitch plugin install ./npm-plugin
```

---

## Exemples de plugins utiles

Voici des idées de plugins que vous pourriez créer:

### Outils de développement
- **npm** - Registries et auth ✅ (exemple ci-dessus)
- **yarn** - Configuration yarn
- **pnpm** - Configuration pnpm
- **pip** - Python package index
- **gem** - Ruby gems
- **cargo** - Rust packages
- **composer** - PHP packages

### Infrastructure
- **terraform** - Workspaces et state
- **ansible** - Inventories et vaults
- **pulumi** - Stacks et configs
- **CDK** - AWS CDK configuration

### Cloud & Kubernetes
- **helm** - Repositories et charts
- **kustomize** - Configuration
- **eksctl** - EKS clusters

### Autres
- **ssh** - SSH keys et config
- **gpg** - GPG keys
- **vscode** - Settings et extensions
- **postgres** - psql configuration

---

## Support et contribution

### Besoin d'aide?

- **Documentation**: [README principal](../README.md)
- **Issues**: [GitHub Issues](https://github.com/hugofrely/envswitch/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hugofrely/envswitch/discussions)

### Contribuer

Votre plugin pourrait aider d'autres développeurs! N'hésitez pas à:

1. Créer un repo GitHub pour votre plugin
2. Documenter l'installation et l'usage
3. Partager le lien dans les discussions EnvSwitch

---

## FAQ

**Q: Est-ce que je dois savoir Go pour créer un plugin?**

Non! Un simple fichier `plugin.yaml` suffit pour la plupart des cas. EnvSwitch gère automatiquement la copie des fichiers de configuration.

**Q: Combien de temps ça prend pour créer un plugin?**

5 minutes pour un plugin simple (juste le YAML). Le plugin NPM ci-dessus peut être créé en moins de 5 minutes.

**Q: Quelle est la structure minimale?**

Juste un fichier `plugin.yaml` avec les 3 champs obligatoires (name, version, tool_name).

**Q: Est-ce que les plugins peuvent contenir des credentials?**

Oui, mais faites attention. Les snapshots contiennent vos fichiers de config incluant les tokens. C'est justement le but (garder vos auth séparées par environnement).

**Q: Comment contribuer un plugin officiel?**

Créez votre plugin, testez-le, puis ouvrez une issue dans le repo EnvSwitch pour proposer de l'ajouter aux plugins officiels.

**Q: Les plugins fonctionnent sur tous les OS?**

Oui, tant que l'outil lui-même est supporté sur cet OS.

**Q: Puis-je avoir plusieurs versions d'un plugin?**

Pour l'instant non, un seul plugin par outil. Mais vous pouvez mettre à jour un plugin existant.
