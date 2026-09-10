# Abhängigkeitsmodell

Seiten speichern Includes, interne Links und Medienreferenzen getrennt. Beim Export werden Includes transitiv ergänzt. Eine `seen`-Menge verhindert Endlosschleifen bei Zyklen. Fehlende Zielseiten werden als Warnungen behandelt; bereits besuchte Seiten werden nicht erneut exportiert.

Im dynamischen Zielmodus repräsentiert die Preview Includes als semantische „Include Page“-Elemente. Ein statischer Renderer kann später denselben Graphen nutzen, um Inhalte physisch einzusetzen.
