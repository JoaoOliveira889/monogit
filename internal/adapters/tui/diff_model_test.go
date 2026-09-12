package tui

import "testing"

const sampleDiff = `diff --git a/main.go b/main.go
index 1234567..89abcde 100644
--- a/main.go
+++ b/main.go
@@ -10,6 +10,7 @@ func main() {
 	setup()
-	oldCall()
+	newCall()
+	extraCall()
 	teardown()
@@ -40,3 +41,3 @@ func other() {
-	gone()
+	added()
 	rest()`

func TestParseUnifiedDiffNumbersLines(t *testing.T) {
	d := parseUnifiedDiff(sampleDiff)

	if d.Added != 3 || d.Removed != 2 {
		t.Errorf("added=%d removed=%d, want 3 and 2", d.Added, d.Removed)
	}
	if len(d.HunkStarts) != 2 {
		t.Fatalf("found %d hunks, want 2", len(d.HunkStarts))
	}

	// The first content row after the first hunk header is context at 10/10.
	first := d.Lines[d.HunkStarts[0]+1]
	if first.Kind != diffContext || first.OldLine != 10 || first.NewLine != 10 {
		t.Errorf("first context row = %+v, want context at 10/10", first)
	}

	// A removed row carries an old number only; an added row a new number only.
	removed := d.Lines[d.HunkStarts[0]+2]
	if removed.Kind != diffRemoved || removed.OldLine != 11 || removed.NewLine != 0 {
		t.Errorf("removed row = %+v, want old=11 new=0", removed)
	}
	added := d.Lines[d.HunkStarts[0]+3]
	if added.Kind != diffAdded || added.NewLine != 11 || added.OldLine != 0 {
		t.Errorf("added row = %+v, want new=11 old=0", added)
	}

	// Numbering resumes from the second hunk's header, not from the first.
	secondCtx := d.Lines[d.HunkStarts[1]+1]
	if secondCtx.OldLine != 40 {
		t.Errorf("second hunk starts at old line %d, want 40", secondCtx.OldLine)
	}
}

func TestParseUnifiedDiffMarksMetadata(t *testing.T) {
	d := parseUnifiedDiff(sampleDiff)

	for _, want := range []struct {
		row  int
		kind diffLineKind
	}{
		{0, diffMeta},       // diff --git
		{1, diffMeta},       // index
		{2, diffFileHeader}, // ---
		{3, diffFileHeader}, // +++
	} {
		if got := d.Lines[want.row].Kind; got != want.kind {
			t.Errorf("row %d kind = %v, want %v (%q)", want.row, got, want.kind, d.Lines[want.row].Text)
		}
	}
}

func TestParseUnifiedDiffHandlesEmptyAndGarbage(t *testing.T) {
	if got := parseUnifiedDiff(""); len(got.Lines) != 0 {
		t.Errorf("empty diff produced %d lines", len(got.Lines))
	}

	// A malformed header must not panic or drop the rest of the diff.
	d := parseUnifiedDiff("@@ nonsense @@\n context\n+added")
	if len(d.Lines) != 3 {
		t.Errorf("got %d lines from a malformed diff, want 3", len(d.Lines))
	}
	if d.Added != 1 {
		t.Errorf("added=%d, want 1", d.Added)
	}
}

func TestHunkMotions(t *testing.T) {
	d := parseUnifiedDiff(sampleDiff)
	first, second := d.HunkStarts[0], d.HunkStarts[1]

	if got := d.hunkAfter(0); got != first {
		t.Errorf("hunkAfter(0) = %d, want %d", got, first)
	}
	if got := d.hunkAfter(first); got != second {
		t.Errorf("hunkAfter(first) = %d, want %d", got, second)
	}
	if got := d.hunkAfter(second); got != -1 {
		t.Errorf("hunkAfter(last) = %d, want -1", got)
	}
	if got := d.hunkBefore(second); got != first {
		t.Errorf("hunkBefore(second) = %d, want %d", got, first)
	}
	if got := d.hunkBefore(first); got != -1 {
		t.Errorf("hunkBefore(first) = %d, want -1", got)
	}
}
