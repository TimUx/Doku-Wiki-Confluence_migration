# Mitwirken

## Lokale Entwicklung

Benötigt werden Go 1.24 oder neuer und GNU Make.

```bash
go mod download
make check
make build
```

Änderungen sollten klein und nachvollziehbar bleiben. Neue Parser-Regeln benötigen passende Tests; sicherheitsrelevante Pfad- oder Exportänderungen benötigen negative Tests für manipulierte Eingaben.

## Pull Requests

1. Branch von `main` erstellen.
2. Änderung implementieren und dokumentieren.
3. `make check` und `make build` ausführen.
4. Pull Request mit Motivation, Testnachweis und bekannten Grenzen eröffnen.

## Repository

Der Modulpfad und alle Repository-Verweise sind auf `github.com/TimUx/Doku-Wiki-Confluence_migration` eingestellt. Nach Änderungen an Abhängigkeiten ist `go mod tidy` auszuführen und die aktualisierte `go.sum` mit zu committen.
