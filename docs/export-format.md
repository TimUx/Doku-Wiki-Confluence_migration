# Exportformat

Das ZIP enthält ein versionsgebundenes Manifest, je Seite HTML, Markdown, Originalquelle und JSON-Metadaten sowie die benötigten Original-Attachments.

## Confluence-Vorbereitung

Jede exportierte Seite enthält zusätzlich `confluence-storage.xml`. Diese Datei ist die technische, bereits auf Confluence ausgerichtete Fassung des Seiteninhalts. Medien werden als Confluence-Attachment-Referenzen mit dem **exakten exportierten Dateinamen** eingebunden.

Beispiel:

```xml
<ac:image><ri:attachment ri:filename="backup.png"/></ac:image>
```

Die Datei `attachments.csv` dokumentiert die Zuordnung von DokuWiki-Medien zu den tatsächlichen Exportdateien. Gleiche Dateinamen aus unterschiedlichen Quellpfaden werden deterministisch mit `_2`, `_3` usw. ergänzt.

## Includes

Includes werden **nicht** rekursiv als Zielseiten in Confluence aufgelöst. Die neue Confluence-Struktur ist zum Exportzeitpunkt unbekannt. Stattdessen bleibt die Position als eindeutiger Platzhalter erhalten:

```text
[DOKUWIKI INCLUDE: server:restore]
```

`includes.csv` listet die gefundenen Includes und ihre Platzhalter. Bei der späteren Übernahme wird der Platzhalter bewusst auf die endgültige Confluence-Seite bzw. das Include Page Macro abgebildet.

## Interne Links

Interne DokuWiki-Links werden in der Confluence-Fassung als `[DOKUWIKI LINK: ...]` markiert, sofern ihr späteres Confluence-Ziel nicht sicher bekannt ist. Dadurch werden keine falschen Seitenziele erfunden.

## Paketstruktur

```text
migration.zip
├── manifest.json
├── MIGRATION-GUIDE.md
├── MIGRATION-GUIDE.html
├── migration-report.html
├── migration-report.json
└── pages/
    └── 001_Titel/
        ├── page.html
        ├── page.md
        ├── confluence-storage.xml
        ├── source.txt
        ├── metadata.json
        ├── attachments.csv
        ├── includes.csv
        └── attachments/
```

Die Anleitung im ZIP beschreibt die manuelle Übernahme, insbesondere das unveränderte Hochladen der Anhänge und die bewusste Auflösung der Include-Platzhalter.
