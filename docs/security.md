# Sicherheit

Die Konfiguration erzwingt Read-only-Quellbetrieb. Pfade werden kanonisiert und gegen die erlaubten Roots geprüft. Page-IDs akzeptieren keine Pfadtrenner oder `..`. ZIP-Namen werden serverseitig bereinigt. Preview-Inhalte werden HTML-escaped; Security-Header sperren fremde Skripte und Ressourcen. Request-Bodies sind begrenzt.

Für den Produktivbetrieb soll der Service-User keine Schreibrechte im DokuWiki besitzen. Die systemd-Unit setzt `ReadOnlyPaths` für das Wiki und `ReadWritePaths` ausschließlich für den App-Arbeitsbereich.
