# KzNginxGenerator

**by Karoza** — un générateur de configurations Nginx robustes et optimisées,
du simple reverse proxy aux architectures complexes (load balancing, SSL/TLS,
FastCGI, cache, rate limiting), via une interface web locale ou en ligne de
commande.

[![CI](https://github.com/karoza/kz-nginx-generator/actions/workflows/ci.yml/badge.svg)](https://github.com/karoza/kz-nginx-generator/actions/workflows/ci.yml)
[![Release](https://github.com/karoza/kz-nginx-generator/actions/workflows/release.yml/badge.svg)](https://github.com/karoza/kz-nginx-generator/actions/workflows/release.yml)
[![Coverage](https://codecov.io/gh/karoza/kz-nginx-generator/branch/main/graph/badge.svg)](https://codecov.io/gh/karoza/kz-nginx-generator)
[![Latest Release](https://img.shields.io/github/v/release/karoza/kz-nginx-generator?include_prereleases&sort=semver)](https://github.com/karoza/kz-nginx-generator/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/karoza/kz-nginx-generator)](https://goreportcard.com/report/github.com/karoza/kz-nginx-generator)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## Sommaire

- [Pourquoi KzNginxGenerator ?](#pourquoi-kznginxgenerator-)
- [Fonctionnalités](#fonctionnalités)
- [Installation](#installation)
- [Utilisation en ligne de commande](#utilisation-en-ligne-de-commande)
- [Interface Web](#interface-web)
- [Modèle de configuration avancé](#modèle-de-configuration-avancé)
- [Développement local](#développement-local)
- [Contribuer](#contribuer)
- [Licence](#licence)

## Pourquoi KzNginxGenerator ?

Écrire une configuration Nginx correcte à la main est source d'erreurs :
oubli d'un header de sécurité, syntaxe SSL incomplète, cache FastCGI mal
déclaré... **KzNginxGenerator** encapsule ces bonnes pratiques dans un
modèle de données Go extensible et un moteur de rendu testé, exposé à la
fois en CLI (pour l'intégration dans vos scripts/CI) et via une interface
web locale (pour construire visuellement des configurations complexes).

## Fonctionnalités

- ⚡ **Reverse Proxy** simple en une commande.
- 🧭 **Load Balancing** : `round_robin`, `least_conn`, `ip_hash`, poids,
  backups, `keepalive`.
- 🔒 **SSL/TLS** complet : certificats, protocoles, ciphers, HSTS, OCSP
  Stapling, redirection HTTP → HTTPS automatique.
- 🛡️ **Security Headers** : HSTS, CSP, X-Frame-Options, X-Content-Type-Options,
  Referrer-Policy, Permissions-Policy.
- 🚀 **HTTP/2** et **HTTP/3 (QUIC)**.
- 🔌 **WebSockets** en un clic.
- 🐘 **PHP-FPM / FastCGI**, avec **cache FastCGI** configurable par location.
- 🚦 **Rate Limiting** par zone, avec `burst`/`nodelay`.
- ✍️ **Directives brutes** (`CustomDirectives`) pour tout ce qui n'est pas
  encore couvert par le modèle structuré.
- 🖥️ **CLI** (`kznginx generate`) et **UI Web locale** (`kznginx ui`) avec
  aperçu et coloration syntaxique en temps réel.

## Installation

### Installation rapide (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/karoza/kz-nginx-generator/main/install.sh | sh
```

Le script détecte automatiquement votre OS et votre architecture, télécharge
le dernier binaire stable depuis les GitHub Releases et l'installe dans
`/usr/local/bin/kznginx`.

### Via `go install`

```bash
go install github.com/karoza/kz-nginx-generator@latest
```

### Depuis les sources

```bash
git clone https://github.com/karoza/kz-nginx-generator.git
cd kz-nginx-generator
make build
./bin/kznginx version
```

### Docker

```bash
docker run --rm -p 8080:8080 ghcr.io/karoza/kz-nginx-generator:latest ui --port 8080
```

## Utilisation en ligne de commande

Générer un reverse proxy simple, affiché sur stdout :

```bash
kznginx generate --domain example.com --proxy http://localhost:8000
```

Écrire directement dans un fichier de configuration Nginx :

```bash
kznginx generate \
  --domain example.com \
  --proxy http://localhost:8000 \
  --out /etc/nginx/sites-available/karoza-app.conf
```

Activer HTTPS (redirection automatique + HSTS) et le support WebSocket :

```bash
kznginx generate \
  --domain example.com \
  --proxy http://localhost:8000 \
  --ssl-cert /etc/ssl/certs/example.com.crt \
  --ssl-key /etc/ssl/private/example.com.key \
  --websocket
```

Afficher la version installée :

```bash
kznginx version
# kznginx v1.0.0 (a1b2c3d)
```

## Interface Web

Pour construire visuellement des configurations plus complexes (plusieurs
upstreams, plusieurs locations, cache FastCGI, rate limiting...), lancez
l'interface web locale :

```bash
kznginx ui --port 8080
```

Puis ouvrez [http://localhost:8080](http://localhost:8080) : ajoutez des
blocs `upstream` et `location` dynamiquement, configurez le SSL et les
headers de sécurité, et copiez la configuration générée en temps réel
directement dans le presse-papier. Aucune donnée ne quitte votre machine —
tout est généré localement par le binaire `kznginx`.

## Modèle de configuration avancé

Le paquet `internal/nginx` peut aussi être utilisé comme bibliothèque Go
dans vos propres outils :

```go
cfg := nginx.Config{
    Upstreams: []nginx.Upstream{
        {
            Name:   "backend",
            Method: nginx.LoadBalanceLeastConn,
            Servers: []nginx.UpstreamServer{
                {Address: "10.0.0.1:8000", Weight: 3},
                {Address: "10.0.0.2:8000", Backup: true},
            },
            KeepAlive: 32,
        },
    },
    Servers: []nginx.Server{
        {
            ServerNames:     []string{"example.com"},
            HTTP2:           true,
            RedirectToHTTPS: true,
            SSL: nginx.SSL{
                Enabled:            true,
                CertificatePath:    "/etc/ssl/example.com.crt",
                CertificateKeyPath: "/etc/ssl/example.com.key",
                HSTS:               true,
            },
            Locations: []nginx.Location{
                {Path: "/", ProxyPass: "http://backend"},
            },
        },
    },
}

output, err := nginx.Render(cfg)
```

## Développement local

Prérequis : [Go 1.23+](https://go.dev/dl/).

```bash
git clone https://github.com/karoza/kz-nginx-generator.git
cd kz-nginx-generator

make test        # tests unitaires + couverture
make lint        # golangci-lint
make run-ui      # lance l'UI web sur http://localhost:8080
make run-generate
```

Structure du projet :

```
cmd/                CLI Cobra (kznginx ui, generate, version)
internal/nginx/      modèle de données + moteur de rendu (text/template)
internal/server/     serveur HTTP de l'UI web + API JSON
web/                 assets statiques de l'UI (embarqués via go:embed)
.github/workflows/   pipelines CI (lint/test/build) et Release (GoReleaser)
install.sh           script d'installation one-liner
```

## Contribuer

Les contributions sont les bienvenues ! Consultez
[CONTRIBUTING.md](CONTRIBUTING.md) pour le workflow GitFlow, la convention
de commits (Conventional Commits) et la procédure de Pull Request.

## Licence

Distribué sous licence [MIT](LICENSE). © 2026 Karoza.
