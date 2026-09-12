package git

import (
	"strings"
	"testing"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

func TestValidateRebaseItem(t *testing.T) {
	tests := []struct {
		name  string
		item  domain.RebaseItem
		valid bool
	}{
		{"pick", domain.RebaseItem{Hash: "abc1234", Action: "pick", Message: "feat: x"}, true},
		{"empty action defaults to pick", domain.RebaseItem{Hash: "abc1234", Message: "feat: x"}, true},
		{"bad hash", domain.RebaseItem{Hash: "zzz", Action: "pick"}, false},
		{"unknown action", domain.RebaseItem{Hash: "abc1234", Action: "exec"}, false},
		{"newline injects a todo line", domain.RebaseItem{
			Hash:    "abc1234",
			Action:  "pick",
			Message: "x\nexec touch /tmp/pwned",
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRebaseItem(tt.item)
			if tt.valid && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("expected the item to be rejected")
			}
		})
	}
}

func TestSequenceEditorCommandQuotesPaths(t *testing.T) {
	got := sequenceEditorCommand("/tmp/dir with space/todo'; touch pwned; '.txt")

	if strings.Contains(got, "; touch pwned") && !strings.Contains(got, `'"'"'`) {
		t.Errorf("path was not quoted: %s", got)
	}
	if !strings.HasSuffix(got, "'") {
		t.Errorf("quoted path is not terminated: %s", got)
	}
}

func TestGetStatusFilesKeepsRenamedTarget(t *testing.T) {
	// git status --porcelain -z emits "R  <new>\0<old>\0" for renames.
	raw := "R  b.txt\x00a.txt\x00 M c.txt\x00"

	files := parseStatusFiles(raw)

	if len(files) != 2 {
		t.Fatalf("got %d files, want 2: %+v", len(files), files)
	}
	if files[0].Name != "b.txt" {
		t.Errorf("rename resolved to %q, want the new path %q", files[0].Name, "b.txt")
	}
	if files[1].Name != "c.txt" {
		t.Errorf("second entry = %q, want %q", files[1].Name, "c.txt")
	}
}
