# Abhängigkeitsmodell

Seiten speichern Includes, interne Links und Medienreferenzen getrennt. Der Scan erfasst diese Beziehungen als Abhängigkeiten, ohne die DokuWiki-Quelle zu verändern.

## Exportverhalten

Beim aktuellen Migrationsexport werden Include-Ziele **nicht transitiv als zusätzliche Zielseiten exportiert**. Ein Include bleibt auf der ausgewählten Quellseite als eindeutiger Platzhalter erhalten, zum Beispiel:

```text
[DOKUWIKI INCLUDE: server:backup]
```

Der Grund ist die bewusst offene Zielstruktur: Die Anwendung kennt beim Export noch nicht, unter welcher Seite bzw. in welchem Space das betreffende Dokument in Confluence liegen wird. Eine automatische Zuordnung könnte deshalb falsche Include-Beziehungen erzeugen.

Die Datei `includes.csv` dokumentiert pro Include die Quellseite, das ursprüngliche DokuWiki-Ziel und den erzeugten Platzhalter.

## Medien und Links

Medien werden separat erfasst und beim Export als konkrete Attachment-Dateien mit deterministischen Namen bereitgestellt. Die Confluence-Storage-Ausgabe referenziert genau diese Namen.

Interne Links werden analysiert und können in der Migration anhand der endgültigen Confluence-Seitenstruktur aufgelöst werden. Externe Links bleiben als externe Verweise erhalten.

## Zyklen

Der Dependency-Graph darf Zyklen enthalten. Die Analyse muss deshalb zyklische Beziehungen erkennen können, ohne in einer Endlosschleife zu laufen. Eine fachliche Auflösung der Zielstruktur findet erst im Migrationsschritt statt.
