package parser

import "testing"

func TestParseCore(t *testing.T) {
 p:=Parse("betrieb:sap:handbuch",`===== Überschrift =====
**Fett** und //kursiv//
[[backup|Backup]] {{architecture.png}}
{{page>monitoring}}
<WRAP warning>
Produktionssystem
</WRAP>
<xyz value="1">`)
 if p.Title!="Überschrift"{t.Fatalf("title=%q",p.Title)}
 if len(p.Links)!=1||p.Links[0].Target!="betrieb:sap:backup"{t.Fatalf("links=%#v",p.Links)}
 if len(p.Media)!=1||p.Media[0].Target!="betrieb:sap:architecture.png"{t.Fatalf("media=%#v",p.Media)}
 if len(p.Includes)!=1||p.Includes[0].Target!="betrieb:sap:monitoring"{t.Fatalf("includes=%#v",p.Includes)}
 if len(p.Warnings)!=0{t.Fatalf("unexpected warnings=%#v",p.Warnings)}
 found:=false;for _,n:=range p.Nodes{if n.Type=="warning"&&n.Meta["kind"]=="warning"{found=true}};if !found{t.Fatal("wrap warning not preserved")}
}
func TestUTF8(t *testing.T){p:=Parse("ä:seite","===== Übergröße =====");if p.Title!="Übergröße"{t.Fatal(p.Title)}}
func TestParseListsTablesAndBlock(t *testing.T){p:=Parse("storage:arrays",`Name - Modell
* pure01 FA-X90R4
* pure02 FA-X90R4
- erster Schritt
- zweiter Schritt
  * unterpunkt
^ Name ^ Modell ^
| pure01 | FA-X90R4 |
<block important>
Snapshots stellen kein Backup dar.
</block>`);var bullets,ordered,rows int;for _,n:=range p.Nodes{switch n.Type{case "bullet_item":bullets++;case "ordered_item":ordered++;case "table_row":rows++;if len(n.Children)!=2{t.Fatalf("table row children=%d",len(n.Children))}}};if bullets!=3||ordered!=2{t.Fatalf("lists bullets=%d ordered=%d",bullets,ordered)};if rows!=2{t.Fatalf("table rows=%d",rows)};found:=false;for _,n:=range p.Nodes{if n.Type=="warning"&&n.Meta["kind"]=="important"{found=true}};if !found{t.Fatal("important wrap not parsed")}}
func TestRealArchitectureMediaSyntax(t *testing.T){src:=`====== Architektur ======
{{:cc33:storage:howto:block:pure:pure_architektur-uebericht.png?direct&800|}}
{{:cc33:storage:howto:block:pure:pure_architektur-test-prod.png?direct&800|}}
{{:cc33:storage:howto:block:pure:pure_r4_rearview.png?nolink&1200|}}
[[https://gb3portal.itscare.prod.dom/wiki/lib/exe/detail.php/cc33:storage:howto:block:pure:pod_1.png?id=cc33:storage:howto:block:pure:replication|{{https://gb3portal.itscare.prod.dom/wiki/lib/exe/fetch.php/cc33:storage:howto:block:pure:pod_1.png?400}}]]
{{cc33:storage:howto:block:pure:pure_storage_snap_erklaerung_englisch.png}}
<block important> Snapshots stellen kein Backup dar und ersetzen dies auch in keinem Fall! </block>`;p:=Parse("cc33:storage:howto:block:pure:architektur",src);if len(p.Media)!=5{t.Fatalf("media=%d %#v",len(p.Media),p.Media)};if p.Media[3].Target!="cc33:storage:howto:block:pure:pod_1.png"{t.Fatalf("media 4=%q",p.Media[3].Target)}}
func TestExtendedDokuWikiSyntax(t *testing.T){p:=Parse("demo:syntax",`====== - Erste Überschrift ======
===== - Unterpunkt =====
__unterstrichen__ ''code'' ~~gelöscht~~ ((Fußnote))
----
{{namespace>demo:manual&firstseconly}}
{{tagtopic>storage&nodate}}
<WRAP info #intro 50%>
Text im Wrap
<WRAP warning>Warnung</WRAP>
</WRAP>`);if len(p.Plugins)==0{t.Fatal("plugins missing")};if p.Plugins[0]!="numberedheadings"{t.Fatalf("plugins=%#v",p.Plugins)};if len(p.Includes)!=2{t.Fatalf("includes=%#v",p.Includes)};if p.Includes[0].Kind!="namespace"||p.Includes[1].Kind!="tagtopic"{t.Fatalf("include kinds=%#v",p.Includes)};if p.Nodes[0].Type!="heading"||p.Nodes[0].Text!="1 Erste Überschrift"{t.Fatalf("heading=%#v",p.Nodes[0])};foundWrap:=false;for _,n:=range p.Nodes{if n.Type=="wrap"{foundWrap=true;if len(n.Children)!=2{t.Fatalf("wrap children=%#v",n.Children)}}};if !foundWrap{t.Fatal("wrap node missing")}}
func TestNumberedHeadingExplicitNumber(t *testing.T){p:=Parse("demo:numbering",`====== -#5 Kapitel ======
===== - Unterkapitel =====
===== -#8 Sprung =====`);if p.Nodes[0].Text!="5 Kapitel"||p.Nodes[1].Text!="5.1 Unterkapitel"||p.Nodes[2].Text!="5.8 Sprung"{t.Fatalf("numbering=%#v",p.Nodes)}}
