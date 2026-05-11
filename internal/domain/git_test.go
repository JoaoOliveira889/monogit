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
