# Reparaturhinweise

Diese Fassung behebt die Fehler der ersten GitHub-Actions-Läufe:

- ungültige RegEx-Rückreferenz im DokuWiki-Heading-Parser entfernt
- identische Anzahl öffnender und schließender `=` im Parser explizit geprüft
- Go auf Version 1.25.13 aktualisiert
- vollständige, mit `go mod tidy` erzeugte Moduldateien aufgenommen
- CI-Modulprüfung auf das nicht verändernde `go mod tidy -diff` umgestellt
- GitHub Actions auf die von Dependabot vorgeschlagenen Hauptversionen aktualisiert
- Quellcode mit `gofmt` formatiert

Nach dem Kopieren über das Repository sollten `go.mod`, `go.sum`, die Parser-Datei, Dokumentation und `.github/workflows/` gemeinsam committet werden. Offene Dependabot-PRs, deren Änderungen damit bereits enthalten sind, können anschließend geschlossen werden.
