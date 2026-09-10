# DokuWiki Parser

Der Parser arbeitet zeilenorientiert mit begrenzten Token-Erkennern und erzeugt ein neutrales Modell. Überschriften, Absätze, Listen, Code, Zitate, Links, Medien, Includes und WRAP-Blöcke werden semantisch gespeichert. Fehlerhafte oder unbekannte Plugin-Syntax wird als `unknown_plugin` samt Originalsyntax und Position erhalten.

Die Parserarchitektur ist bewusst tolerant: erhalten, warnen und berichten hat Vorrang vor Verwerfen.
