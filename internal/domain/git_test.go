package domain

import "testing"

func TestRepositoryStruct(t *testing.T) {
	r := Repository{
		Name: "test-repo",
		Path: "/path/to/test-repo",
	}

	if r.Name != "test-repo" {
		t.Errorf("expected test-repo, got %s", r.Name)
	}
}

func TestFileStatus(t *testing.T) {
	tests := []struct {
		name      string
		fs        FileStatus
		wantStaged bool
		wantMod    bool
		wantUntr   bool
	}{
		{"staged", FileStatus{Name: "f1", Staged: true}, true, false, false},
		{"modified", FileStatus{Name: "f2", Modified: true}, false, true, false},
		{"untracked", FileStatus{Name: "f3", Untracked: true}, false, false, true},
		{"all", FileStatus{Name: "f4", Staged: true, Modified: true}, true, true, false},
		{"empty", FileStatus{}, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fs.Staged != tt.wantStaged {
				t.Errorf("Staged = %v, want %v", tt.fs.Staged, tt.wantStaged)
			}
			if tt.fs.Modified != tt.wantMod {
				t.Errorf("Modified = %v, want %v", tt.fs.Modified, tt.wantMod)
			}
			if tt.fs.Untracked != tt.wantUntr {
				t.Errorf("Untracked = %v, want %v", tt.fs.Untracked, tt.wantUntr)
			}
		})
	}
}

func TestBranchInfo(t *testing.T) {
	b := BranchInfo{Name: "main", IsLocal: true, IsRemote: true, IsCurrent: true}
	if b.Name != "main" || !b.IsLocal || !b.IsRemote || !b.IsCurrent {
		t.Errorf("unexpected BranchInfo: %+v", b)
	}
}

func TestRepositoryDefaults(t *testing.T) {
	r := Repository{}
	if r.Name != "" || r.Path != "" {
		t.Error("empty Repository should have zero values")
	}
}
