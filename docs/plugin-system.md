# Plugin-System

Unterstützte Syntax wird zentral im Parser klassifiziert. `include` erzeugt einen Include-Node; `WRAP info|warning|note|tip` wird auf semantische Panel-Nodes abgebildet. Unbekannte Tags erzeugen einen unveränderten Raw-Node und eine Warnung. Neue Handler sollten nach `CanHandle` und `Parse` getrennt und in einer Registry registriert werden.
