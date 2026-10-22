package placeholder

import "testing"

func ptr(b bool) *bool { return &b }

func TestBR2_2_ContentEditClears(t *testing.T) {
	before := map[string]string{"quote_en": "[Client quote]", "published": "false", "sort": "1"}
	after := map[string]string{"quote_en": "Sunshine answered in ten minutes.", "published": "false", "sort": "1"}
	changed := ContentChanged(before, after, []string{"quote_en", "quote_id", "name"})
	if !changed {
		t.Fatal("a changed quote was not detected")
	}
	if Next(true, changed, nil) {
		t.Fatal("the placeholder flag survived a content edit")
	}
}

func TestBR2_2_OrderAndPublishDoNotClear(t *testing.T) {
	before := map[string]string{"quote_en": "[Client quote]", "published": "false", "sort": "1"}
	after := map[string]string{"quote_en": "[Client quote]", "published": "true", "sort": "3"}
	changed := ContentChanged(before, after, []string{"quote_en"})
	if changed {
		t.Fatal("publishing or reordering was counted as a content edit")
	}
	if !Next(true, changed, nil) {
		t.Fatal("the flag cleared without a content edit")
	}
}

func TestBR2_3_ExplicitWins(t *testing.T) {
	if !Next(true, true, ptr(true)) {
		t.Fatal("an explicit 'still a placeholder' was overridden by the edit")
	}
	if Next(true, false, ptr(false)) {
		t.Fatal("an explicit clear was ignored")
	}
	if !Next(false, false, ptr(true)) {
		t.Fatal("an editor could not mark a row placeholder")
	}
}

func TestNonPlaceholderStaysClear(t *testing.T) {
	if Next(false, true, nil) {
		t.Fatal("a real row became a placeholder by being edited")
	}
}
