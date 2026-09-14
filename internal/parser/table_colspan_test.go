package parser

import "testing"

func TestParseTableRowColspanMarker(t *testing.T) {
	p := Parse("demo:table", "^ Name ^ Value ^\n| Alpha | ::: |\n")
	if len(p.Nodes) != 2 {
		t.Fatalf("expected heading and table row, got %#v", p.Nodes)
	}
	row := p.Nodes[1]
	if len(row.Children) != 1 {
		t.Fatalf("expected one cell after colspan marker, got %#v", row.Children)
	}
	if got := row.Children[0].Meta["colspan"]; got != "2" {
		t.Fatalf("expected colspan=2, got %q", got)
	}
}

func TestParseTableRowMultipleColspanMarkers(t *testing.T) {
	p := Parse("demo:table", "| Alpha | ::: | ::: |\n")
	if len(p.Nodes) != 1 || len(p.Nodes[0].Children) != 1 {
		t.Fatalf("expected one cell after colspan markers, got %#v", p.Nodes)
	}
	if got := p.Nodes[0].Children[0].Meta["colspan"]; got != "3" {
		t.Fatalf("expected colspan=3, got %q", got)
	}
}
