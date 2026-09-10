package parser
import "testing"
func TestParseCore(t *testing.T){p:=Parse("betrieb:sap:handbuch",`===== Überschrift =====
**Fett** und //kursiv//
[[backup|Backup]] {{architecture.png}}
{{page>monitoring}}
<WRAP warning>
Produktionssystem
</WRAP>
<xyz value="1">`);if p.Title!="Überschrift"{t.Fatalf("title=%q",p.Title)};if len(p.Links)!=1||p.Links[0].Target!="betrieb:sap:backup"{t.Fatalf("links=%#v",p.Links)};if len(p.Media)!=1||p.Media[0].Target!="betrieb:sap:architecture.png"{t.Fatalf("media=%#v",p.Media)};if len(p.Includes)!=1||p.Includes[0].Target!="betrieb:sap:monitoring"{t.Fatalf("includes=%#v",p.Includes)};if len(p.Warnings)!=1{t.Fatalf("warnings=%#v",p.Warnings)}}
func TestUTF8(t *testing.T){p:=Parse("ä:seite","===== Übergröße =====");if p.Title!="Übergröße"{t.Fatal(p.Title)}}
