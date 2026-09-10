# Weboberfläche und automatische Screenshots

Die Screenshots werden nicht manuell gepflegt. Der Workflow `UI Screenshots` baut die Anwendung, startet eine isolierte Testinstanz mit `testdata/dokuwiki`, führt einen read-only Scan aus und öffnet die Oberfläche in einem Headless-Chromium-Browser.

## Migrationsübersicht

![Migrationsübersicht mit Testdaten](screenshots/dashboard.png)

## Seitenansicht

![Seitendetail mit Vorschau](screenshots/page-detail.png)

## Plugin-Inventar

![Plugin-Inventar](screenshots/plugins.png)

Der Workflow speichert die drei PNG-Dateien außerdem als Actions-Artefakt und aktualisiert sie auf `main`, wenn sich die Darstellung geändert hat. Der Screenshot-Commit löst wegen der gesetzten Pfadfilter keinen weiteren Screenshot-Lauf aus.

Schlägt eine Browserprüfung fehl, wird nach Möglichkeit zusätzlich `ui-failure.png` aufgenommen. Zusammen mit `server.log` und `scan.json` wird die Datei auch bei einem fehlgeschlagenen Lauf als Diagnose-Artefakt hochgeladen.

Falls der direkte Bot-Push durch Branch-Schutz blockiert wird, muss unter **Settings → Actions → General → Workflow permissions** Schreibzugriff erlaubt und die Branch-Regel für den GitHub-Actions-Bot passend konfiguriert werden. Alternativ können die PNG-Dateien aus dem Workflow-Artefakt manuell übernommen werden.
