# Repository auf GitHub veröffentlichen

## Empfohlener Weg mit Git

Lege auf GitHub ein leeres Repository ohne automatisch erzeugte README, Lizenz oder `.gitignore` an. Entpacke dieses Paket und führe im Projektordner aus:

```bash
git init
git branch -M main
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/TimUx/Doku-Wiki-Confluence_migration.git
git push -u origin main
```

Der Go-Modulpfad, alle internen Imports, Badges und Repository-Links sind bereits auf `github.com/TimUx/Doku-Wiki-Confluence_migration` eingestellt. Führe mit installierter Go-Toolchain vor dem ersten Push `go mod tidy` aus und committe die erzeugte `go.sum`.

## Erster CI-Lauf

Der erste Push startet CI und Sicherheitsprüfung automatisch. Prüfe unter **Actions**, ob beide Workflows erfolgreich sind. Der CI-Lauf erzeugt außerdem eine Testabdeckung und eine Linux-AMD64-Binary als temporäre Artefakte.

## Branch-Schutz

Aktiviere unter **Settings → Branches** für `main` mindestens:

- Pull Request vor Merge erforderlich
- Statuscheck `Test, Vet & Build` erforderlich
- Statuscheck `Go Vulnerability Check` erforderlich
- Löschen und Force-Pushes verbieten

## Release erzeugen

Nach einem erfolgreichen CI-Lauf:

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

Der Release-Workflow veröffentlicht Linux-Binaries für AMD64 und ARM64 zusammen mit `SHA256SUMS`.
