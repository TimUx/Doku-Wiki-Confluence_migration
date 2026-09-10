# Reparaturhinweise

Diese Fassung behebt die Fehler der ersten GitHub-Actions-Läufe:

- ungültige RegEx-Rückreferenz im DokuWiki-Heading-Parser entfernt
- identische Anzahl öffnender und schließender `=` im Parser explizit geprüft
- Go auf Version 1.26.8 aktualisiert, damit `govulncheck@latest` auch mit `GOTOOLCHAIN=local` installiert werden kann
- `modernc.org/sqlite` auf Version 1.58.0 aktualisiert
- vollständige, mit `go mod tidy` erzeugte Moduldateien aufgenommen
- CI-Modulprüfung auf das nicht verändernde `go mod tidy -diff` umgestellt
- GitHub Actions auf die von Dependabot vorgeschlagenen Hauptversionen aktualisiert
- reproduzierbaren Screenshot-Workflow mit fiktivem Test-Wiki ergänzt
- README und UI-Dokumentation an die automatisch erzeugten Screenshots angebunden
- Screenshot-Prüfung auf den tatsächlichen Testseitentitel `SAP Betriebshandbuch` korrigiert
- Diagnose-Artefakte und Fehler-Screenshot auch bei fehlgeschlagenen UI-Läufen aktiviert
- Quellcode mit `gofmt` formatiert

Nach dem Kopieren über das Repository sollten `go.mod`, `go.sum`, die Parser-Datei, Dokumentation und `.github/workflows/` gemeinsam committet werden. Offene Dependabot-PRs, deren Änderungen damit bereits enthalten sind, können anschließend geschlossen werden.
