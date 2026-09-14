package parser

import (
    "strings"
    "testing"
)

func TestParseNowikiBlockPreservesLiteralContent(t *testing.T) {
    src := `<nowiki>[[not:a:link]] **not bold** {{image.png}}</nowiki>`
    p := Parse("demo:nowiki", src)
    if len(p.Nodes) != 1 { t.Fatalf("expected one node, got %d: %#v", len(p.Nodes), p.Nodes) }
    n := p.Nodes[0]
    if n.Type != "nowiki" { t.Fatalf("expected nowiki node, got %q", n.Type) }
    if n.Text != `[[not:a:link]] **not bold** {{image.png}}` { t.Fatalf("unexpected nowiki text: %q", n.Text) }
    if len(p.Links) != 0 || len(p.Media) != 0 || len(p.Includes) != 0 { t.Fatalf("nowiki content must not create references: links=%#v media=%#v includes=%#v", p.Links, p.Media, p.Includes) }
}

func TestParseNowikiMultilineBlock(t *testing.T) {
    src := "<nowiki>\n[[not:a:link]]\n**not bold**\n</nowiki>"
    p := Parse("demo:nowiki", src)
    if len(p.Nodes) != 1 || p.Nodes[0].Type != "nowiki" { t.Fatalf("expected one nowiki node, got %#v", p.Nodes) }
    if p.Nodes[0].Text != "[[not:a:link]]\n**not bold**" { t.Fatalf("unexpected multiline nowiki text: %q", p.Nodes[0].Text) }
    if strings.Contains(p.Nodes[0].Text, "</nowiki>") { t.Fatal("closing nowiki tag leaked into node text") }
}
