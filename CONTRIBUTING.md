# Contribuer à KzNginxGenerator

Merci de votre intérêt pour **KzNginxGenerator** (by Karoza) ! Ce document explique
le workflow de développement, les conventions de commit et comment lancer les
tests et soumettre une Pull Request.

## Workflow Git (GitFlow)

Le projet suit un modèle **GitFlow** simplifié :

| Branche         | Rôle                                                                 |
|------------------|-----------------------------------------------------------------------|
| `main`          | Code stable, chaque commit correspond à une release taguée (`vX.Y.Z`).|
| `develop`       | Branche d'intégration, base de toutes les nouvelles fonctionnalités.  |
| `feature/*`     | Une fonctionnalité ou un correctif, créée à partir de `develop`.      |
| `release/*`     | Stabilisation d'une prochaine version (souvent des pré-releases beta).|

Workflow type pour une contribution :

1. Créez votre branche à partir de `develop` :
   ```bash
   git checkout develop
   git pull
   git checkout -b feature/mon-super-ajout
   ```
2. Développez, committez, testez.
3. Ouvrez une Pull Request vers `develop`.
4. Lorsqu'une version est prête à être publiée, une branche `release/vX.Y.Z`
   est créée depuis `develop` pour la stabilisation (corrections mineures,
   mise à jour du changelog), généralement publiée d'abord en pré-release
   (`vX.Y.Z-beta.1`, `vX.Y.Z-rc.1`, ...) avant d'être fusionnée dans `main`
   et taguée en version stable.

## Convention de commits

Le projet utilise [Conventional Commits](https://www.conventionalcommits.org/) :

```
<type>(<scope optionnel>): <description>

[corps optionnel]

[footer optionnel]
```

Types principaux :

- `feat:` — nouvelle fonctionnalité
- `fix:` — correction de bug
- `docs:` — documentation uniquement
- `test:` — ajout ou modification de tests
- `refactor:` — changement de code sans impact fonctionnel
- `chore:` — maintenance (dépendances, outillage, CI...)

Exemples :

```
feat(nginx): ajoute le support du cache FastCGI par location
fix(cli): corrige l'écriture du fichier de sortie avec --out
docs(readme): ajoute un exemple d'upstream avec least_conn
```

## Lancer les tests localement

Le projet nécessite [Go 1.23+](https://go.dev/dl/).

```bash
# Tous les tests, avec couverture
make test

# Couverture détaillée par fonction
make test-cover

# Lint (nécessite golangci-lint installé localement)
make lint

# Build du binaire dans ./bin/kznginx
make build
```

Le paquet `internal/nginx` (modèle de données + moteur de rendu) doit
conserver une couverture de tests d'au moins **80 %** : toute nouvelle
directive ou structure ajoutée au modèle doit être accompagnée de tests
dans `internal/nginx/*_test.go`.

## Style de code

- Le code est formaté avec `gofmt` (vérifié en CI via `golangci-lint`).
- Pas de dépendances inutiles : privilégiez la bibliothèque standard.
- Toute nouvelle directive Nginx supportée doit :
  1. Être ajoutée au modèle de données dans `internal/nginx/model.go`.
  2. Être validée si nécessaire dans `internal/nginx/validate.go`.
  3. Être rendue dans le template correspondant sous
     `internal/nginx/templates/`.
  4. Être couverte par un test dans `internal/nginx/render_test.go`.

## Soumettre une Pull Request

1. Assurez-vous que `make test` et `make lint` passent localement.
2. Rebasez votre branche sur la dernière version de `develop` si nécessaire.
3. Ouvrez la Pull Request avec une description claire du changement et de
   sa motivation (le "pourquoi", pas seulement le "quoi").
4. La CI (lint, tests, build multi-plateforme) doit passer au vert avant
   toute fusion.
5. Au moins une review est requise avant le merge.

Merci de contribuer à KzNginxGenerator !
