# Benutzerhandbuch – DokuWiki → Confluence Migrator

Dieses Handbuch beschreibt den kompletten Ablauf von der ersten Prüfung des DokuWiki bis zum kontrollierten Export für Confluence 10.2.17.

> **Wichtig:** Die Anwendung arbeitet read-only auf dem DokuWiki. Sie verändert keine Quelldateien. Die spätere Confluence-Seitenstruktur wird bewusst nicht erraten.

## 1. Voraussetzungen prüfen

Vor dem ersten Scan müssen mindestens folgende Punkte stimmen:

- Die Binary wurde auf dem Migrationsserver installiert.
- Der Service-User kann `pages_path` und `media_path` lesen.
- `security.read_only_source` steht auf `true`.
- Arbeits-, Datenbank- und Exportverzeichnisse liegen in den konfigurierten Storage-Pfaden.
- Vor der Migration existiert ein reguläres Backup des DokuWiki.

Die Konfiguration kann vor dem Start mit `check-config --config …` geprüft werden.

## 2. Anwendung starten

Nach dem Start ist die Weboberfläche standardmäßig unter Port 8080 erreichbar.

```text
http://<server>:8080
```

Die Oberfläche zeigt bereits auf der Startseite, dass die Quelle geschützt und nur lesbar eingebunden ist.

## 3. DokuWiki scannen

Öffne die **Migrationsübersicht** und klicke auf **Wiki scannen**.

![Migrationsübersicht](screenshots/01-dashboard.png)

Beim Scan werden unter anderem erfasst:

- DokuWiki-Seiten und Namespaces
- Originalmedien
- interne und externe Links
- Includes
- unterstützte und unbekannte Plugins
- Parser-Warnungen

Der Scan aktualisiert den lokalen SQLite-Index. Das Quell-Wiki wird dabei nicht verändert.

## 4. Seiten suchen und auswählen

Wechsle zu **Seiten & Namespaces**. Über das Suchfeld kannst du nach Page-ID, Titel, Namespace oder Inhalt suchen.

![Seiten und Namespaces](screenshots/02-pages.png)

Setze bei den gewünschten Seiten den Auswahlhaken. Der Migrationsassistent zeigt anschließend eine Zusammenfassung der Auswahl.

![Ausgewählte Seite](screenshots/03-selection.png)

Die Auswahl ist bewusst explizit: Es werden keine zusätzlichen Seiten nur deshalb automatisch als zukünftige Confluence-Seiten angelegt, weil sie von einem Include referenziert werden.

## 5. Seite und Quellinhalt prüfen

Klicke eine ausgewählte Seite an. Im Seitendetail stehen mehrere Ansichten zur Verfügung:

- **Vorschau** – gerenderte Darstellung des erkannten Inhalts
- **Source** – unveränderter DokuWiki-Quelltext
- **Migration** – Hinweise auf die für Confluence vorbereiteten Bestandteile

![Für die Migration vorbereitete Seite](screenshots/04-migration-preview.png)

Prüfe hier insbesondere Warnungen, Includes, Medien, Tabellen, Codeblöcke und Links.

## 6. Übernahme vorbereiten

Im Migrationsassistenten startet **Prüfung starten →** den zweiten Schritt.

Hier wird angezeigt, welche Bestandteile beim Export berücksichtigt werden. Die Anwendung erzeugt bereits die technischen Confluence-Referenzen für Attachments.

### Bilder

Für jedes exportierte Medium wird ein deterministischer Attachment-Dateiname erzeugt. Die Confluence-Ausgabe referenziert genau diesen Namen. Beim späteren Upload in Confluence darf der Dateiname deshalb **nicht geändert** werden.

Die Datei `attachments.csv` dokumentiert die Zuordnung:

```text
DokuWiki-Ziel → Confluence-Anhangsname → MIME
```

### Includes

Includes werden **nicht** auf eine vermutete Confluence-Seite aufgelöst. Stattdessen bleibt eine eindeutige Markierung erhalten:

```text
[DOKUWIKI INCLUDE: server:backup]
```

Damit kann die endgültige Zielseite später festgelegt werden, sobald die neue Confluence-Struktur bekannt ist.

## 7. ZIP-Export erstellen

Nach der Prüfung wählst du **Export vorbereiten →** und anschließend **ZIP erstellen und herunterladen**.

![Export vorbereiten](screenshots/05-export.png)

Das ZIP enthält pro ausgewählter Seite unter anderem:

```text
pages/001_<Titel>/
├── page.html
├── page.md
├── confluence-storage.xml
├── source.txt
├── metadata.json
├── attachments/
├── attachments.csv
└── includes.csv
```

Zusätzlich werden globale Manifest-, Report- und Migrationsanleitungsdateien erzeugt.

## 8. Export kontrollieren

Vor der Übernahme nach Confluence sollte das ZIP stichprobenartig geprüft werden.

### Pro Seite prüfen

1. `source.txt` enthält den erwarteten Originalinhalt.
2. `page.html` und `page.md` enthalten die aufbereiteten Inhalte.
3. `confluence-storage.xml` enthält die vorbereiteten Confluence-Storage-Elemente.
4. Alle Dateien unter `attachments/` sind vorhanden.
5. `attachments.csv` stimmt mit den tatsächlichen Attachment-Dateien überein.
6. `includes.csv` enthält alle Include-Platzhalter.
7. `metadata.json` und die Warnungen wurden geprüft.

## 9. Inhalte nach Confluence übernehmen

Die eigentliche Übernahme ist ein kontrollierter manueller Schritt:

1. Ziel-Space und Zielseite in Confluence bestimmen.
2. Zielseite anlegen und Titel prüfen.
3. Alle benötigten Dateien aus `attachments/` hochladen.
4. **Attachment-Dateinamen unverändert lassen.**
5. Den vorbereiteten Inhalt bzw. das Confluence-Storage-Format entsprechend dem vorgesehenen Importweg übernehmen.
6. Include-Platzhalter durch die endgültige Confluence-Zielseite bzw. das passende **Include Page Macro** ersetzen.
7. Interne DokuWiki-Links auf die endgültigen Confluence-Seiten abbilden.
8. Bilder, Tabellen, Codeblöcke und Formatierungen kontrollieren.

> Die Anwendung entscheidet absichtlich nicht automatisch, welche Confluence-Seite ein Include oder ein interner DokuWiki-Link künftig referenzieren soll.

## 10. Abschlusskontrolle

Nach der Übernahme sollte jede Seite mindestens einmal fachlich geprüft werden:

- [ ] Titel und Namespace/Zielposition stimmen.
- [ ] Bilder werden angezeigt.
- [ ] Tabellen sind korrekt.
- [ ] Codeblöcke sind lesbar.
- [ ] interne Links zeigen auf die richtigen Confluence-Seiten.
- [ ] Include-Platzhalter wurden vollständig bearbeitet.
- [ ] externe Links funktionieren.
- [ ] unbekannte Plugins wurden fachlich bewertet.
- [ ] keine unbeabsichtigten Inhalte fehlen.

## 11. Plugin-Inventar

Unter **Plugin-Inventar** werden erkannte Plugins zusammengefasst. Unterstützte Syntax kann automatisch verarbeitet werden; unbekannte Plugins werden sichtbar gemacht und mit einem Prüfhinweis versehen.

![Plugin-Inventar](screenshots/06-plugins.png)

Das Inventar ist besonders vor einer Serienmigration wichtig, weil DokuWiki-Plugins installationsabhängig sein können.

## 12. Typische Probleme

### Der Scan findet keine Seiten

Prüfe `pages_path`, Dateirechte und ob der Service-User tatsächlich auf `data/pages` zugreifen kann.

### Ein Medium fehlt

Prüfe die Media-ID, den Namespace sowie Groß-/Kleinschreibung im DokuWiki. Zusätzlich sollte `media_path` auf das richtige `data/media` zeigen.

### Ein Include bleibt sichtbar

Das ist bei diesem Migrationskonzept zunächst **korrekt**. Der Platzhalter soll verhindern, dass die Anwendung eine falsche zukünftige Confluence-Struktur annimmt. Er wird nach Festlegung der Zielstruktur ersetzt.

### Ein Plugin wird als unbekannt angezeigt

Den Originalquelltext und die Warnung prüfen. Unbekannte Plugin-Syntax wird bewusst nicht geraten oder stillschweigend in eine andere Funktion umgewandelt.

### SQLite ist gesperrt

Sicherstellen, dass nicht mehrere Instanzen gleichzeitig dieselbe Datenbank verwenden.

## 13. Was das Tool bewusst nicht macht

Der Migrator ist kein automatisches Confluence-Space-Backup und kein Blindkonverter. Er:

- verändert das DokuWiki nicht,
- errät keine zukünftige Confluence-Hierarchie,
- löst Includes nicht gegen unbekannte Zielseiten auf,
- verschweigt unbekannte Plugins nicht,
- verändert Attachment-Dateinamen nicht nachträglich,
- ersetzt keine fachliche Abnahme der migrierten Inhalte.

Das Ziel ist eine **nachvollziehbare, reproduzierbare und kontrollierbare Migration**.
