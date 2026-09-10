# Confluence 10.2.17 – manueller Workflow

Ohne REST-API ist der Export ein Assistenzpaket, kein nativer Space-Import. Der vorgesehene Pilotablauf ist:

1. Zielseiten und Hierarchie in Confluence anlegen.
2. Attachments je Seite hochladen, bevor Bildreferenzen final geprüft werden.
3. gerenderten Inhalt kontrolliert einfügen und Tabellen, Codeblöcke sowie Panels prüfen.
4. dynamische Includes als Confluence „Include Page“-Makro nachbauen.
5. interne Links auf die tatsächlich angelegten Confluence-Seiten setzen.
6. unbekannte Plugins anhand des Reports manuell abbilden.

Welche Importformate, Makros und Attachment-Workflows konkret verfügbar sind, hängt von installierten Apps, Berechtigungen und Instanzkonfiguration ab. Dies muss mit einer Pilotseite in der Zielinstanz verifiziert werden. Ein generischer HTML- oder Word-Import darf nicht als verlustfreie Übernahme von Makros, internen Links oder Attachments angenommen werden.
