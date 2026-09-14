package parser

import "testing"

func TestParseCore(t *testing.T) {
	p := Parse("betrieb:sap:handbuch", `===== Überschrift =====
**Fett** und //kursiv//
[[backup|Backup]] {{architecture.png}}
{{page>monitoring}}
<WRAP warning>
Produktionssystem
</WRAP>
<xyz value="1">`)
	if p.Title != "Überschrift" { t.Fatalf("title=%q", p.Title) }
	if len(p.Links) != 1 || p.Links[0].Target != "betrieb:sap:backup" { t.Fatalf("links=%#v", p.Links) }
	if len(p.Media) != 1 || p.Media[0].Target != "betrieb:sap:architecture.png" { t.Fatalf("media=%#v", p.Media) }
	if len(p.Includes) != 1 || p.Includes[0].Target != "betrieb:sap:monitoring" { t.Fatalf("includes=%#v", p.Includes) }
	if len(p.Warnings) != 1 { t.Fatalf("warnings=%#v", p.Warnings) }
}

func TestUTF8(t *testing.T) {
	p := Parse("ä:seite", "===== Übergröße =====")
	if p.Title != "Übergröße" { t.Fatal(p.Title) }
}

func TestParseListsTablesAndBlock(t *testing.T) {
	p := Parse("storage:arrays", `Name - Modell
* pure01 FA-X90R4
* pure02 FA-X90R4
- erster Schritt
- zweiter Schritt
^ Name ^ Modell ^
| pure01 | FA-X90R4 |
<block important>
Snapshots stellen kein Backup dar.
</block>`)
	var bullets, ordered, rows int
	for _, n := range p.Nodes { switch n.Type { case "bullet_item": bullets++; case "ordered_item": ordered++; case "table_row": rows++; if len(n.Children)!=2 { t.Fatalf("table row children=%d",len(n.Children)) } } }
	if bullets!=2 || ordered!=2 { t.Fatalf("lists bullets=%d ordered=%d",bullets,ordered) }
	if rows!=2 { t.Fatalf("table rows=%d",rows) }
	foundImportant:=false; for _,n:=range p.Nodes { if n.Type=="warning"&&n.Meta!=nil&&n.Meta["kind"]=="important" {foundImportant=true} }; if !foundImportant {t.Fatal("important block was not parsed as warning")}
}

func TestRealArchitectureMediaSyntax(t *testing.T) {
	src := `====== Architektur ======
{{:cc33:storage:howto:block:pure:pure_architektur-uebericht.png?direct&800|}}
{{:cc33:storage:howto:block:pure:pure_architektur-test-prod.png?direct&800|}}
{{:cc33:storage:howto:block:pure:pure_r4_rearview.png?nolink&1200|}}
[[https://gb3portal.itscare.prod.dom/wiki/lib/exe/detail.php/cc33:storage:howto:block:pure:pod_1.png?id=cc33:storage:howto:block:pure:replication|{{https://gb3portal.itscare.prod.dom/wiki/lib/exe/fetch.php/cc33:storage:howto:block:pure:pod_1.png?400}}]]
{{cc33:storage:howto:block:pure:pure_storage_snap_erklaerung_englisch.png}}
<block important> Snapshots stellen kein Backup dar und ersetzen dies auch in keinem Fall! </block>`
	p := Parse("cc33:storage:howto:block:pure:architektur", src)
	if len(p.Media) != 5 { t.Fatalf("media=%d %#v", len(p.Media), p.Media) }
	if p.Media[0].Target != "cc33:storage:howto:block:pure:pure_architektur-uebericht.png" { t.Fatalf("media 1=%q", p.Media[0].Target) }
	if p.Media[3].Target != "cc33:storage:howto:block:pure:pod_1.png" { t.Fatalf("media 4=%q", p.Media[3].Target) }
	if p.Media[4].Target != "cc33:storage:howto:block:pure:pure_storage_snap_erklaerung_englisch.png" { t.Fatalf("media 5=%q", p.Media[4].Target) }
	foundImportant:=false; for _,n:=range p.Nodes { if n.Type=="warning"&&n.Meta!=nil&&n.Meta["kind"]=="important" {foundImportant=true} }; if !foundImportant {t.Fatal("inline important block not parsed")}
}
