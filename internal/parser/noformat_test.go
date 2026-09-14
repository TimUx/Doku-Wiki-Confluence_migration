package parser

import "testing"

func TestNoFormatDoesNotCollectLinksOrMedia(t *testing.T) {
	p := Parse("demo:page", `%%[[demo:target|kein Link]] {{image.png}}%%`)
	if len(p.Links) != 0 {
		t.Fatalf("no-format link was collected: %+v", p.Links)
	}
	if len(p.Media) != 0 {
		t.Fatalf("no-format media was collected: %+v", p.Media)
	}
}

func TestNoFormatDoesNotHideReferencesOutsideSpan(t *testing.T) {
	p := Parse("demo:page", `%%[[demo:literal]]%% [[demo:real]] {{real.png}}`)
	if len(p.Links) != 1 || p.Links[0].Target != "demo:real" {
		t.Fatalf("unexpected links: %+v", p.Links)
	}
	if len(p.Media) != 1 || p.Media[0].Target != "demo:real.png" {
		t.Fatalf("unexpected media: %+v", p.Media)
	}
}
