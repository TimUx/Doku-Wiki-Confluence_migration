# Architektur

Die Binary besitzt vier Schichten: HTTP/UI, Anwendungsdienste, neutrales Dokumentmodell und Infrastruktur. Der Scanner liest DokuWiki-Dateien, der Parser erzeugt typisierte Nodes und Referenzen, SQLite hält den wiederverwendbaren Index. Renderer erzeugen Preview- bzw. Exportformate. Quell- und Arbeitsverzeichnis sind strikt getrennt.

Der Scan ist fehlertolerant: ein Lesefehler an einer einzelnen Seite wird als Warnung gespeichert. Nur ein nicht begehbares Quellverzeichnis bricht den Gesamtlauf ab. Gleichzeitige Scans werden abgewiesen.
