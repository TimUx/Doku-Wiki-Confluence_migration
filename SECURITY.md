# Sicherheitsrichtlinie

## Unterstützte Versionen

Bis zur ersten stabilen Version wird ausschließlich der aktuelle Stand von `main` gepflegt.

## Sicherheitslücken melden

Bitte nutze eine private GitHub Security Advisory des Repositorys. Veröffentliche keine Zugangsdaten, absoluten Serverpfade, vertraulichen Wiki-Inhalte oder produktiven Exportdateien in Issues.

Eine Meldung sollte betroffene Versionen, Reproduktionsschritte, mögliche Auswirkungen und – sofern bekannt – einen Lösungsvorschlag enthalten. Nach Eingang werden Erhalt und weiteres Vorgehen zeitnah bestätigt.

## Sicherheitsgrenzen

Die Anwendung behandelt das DokuWiki als nicht vertrauenswürdige, ausschließlich lesbare Quelle. Einzelheiten zu Pfadvalidierung, Symlinks, XSS und ZIP-Inhalten stehen in [docs/security.md](docs/security.md).
