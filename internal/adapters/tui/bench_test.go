package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

func benchModel(repoCount int) *Model {
	m := NewModel("/tmp", 0, nil)
	m.repos = make([]domain.Repository, repoCount)
	for i := range m.repos {
		m.repos[i] = domain.Repository{
			Name:           fmt.Sprintf("repo-%03d", i),
			Path:           fmt.Sprintf("/tmp/repo-%03d", i),
			Branch:         "main",
			Ahead:          i % 3,
			Behind:         i % 5,
			IsDirty:        i%2 == 0,
			ModifiedCount:  i % 7,
			UntrackedCount: i % 4,
		}
	}
	m.showSplash = false
	m.invalidateFilterCache()
	m.handleResize(tea.WindowSizeMsg{Width: 160, Height: 48})
	return &m
}

// BenchmarkStartupStatusBurst mirrors what happens when a scan completes: one
// repoStatusMsg per repository arrives back to back.
func BenchmarkStartupStatusBurst(b *testing.B) {
	for _, n := range []int{25, 100} {
		b.Run(fmt.Sprintf("repos=%d", n), func(b *testing.B) {
			m := benchModel(n)
			b.ResetTimer()
			for range b.N {
				for i := range m.repos {
					m.Update(repoStatusMsg{
						index:  i,
						branch: "main",
						dirty:  true,
					})
				}
				m.View()
			}
		})
	}
}

func BenchmarkCursorMove(b *testing.B) {
	m := benchModel(100)
	b.ResetTimer()
	for i := range b.N {
		delta := 1
		if i%2 == 0 {
			delta = -1
		}
		m.handleCursorMove(delta)
		m.View()
	}
}

func BenchmarkViewWithModalOpen(b *testing.B) {
	m := benchModel(100)
	m.promptConfirm("Push all repositories?", "Only repositories ahead will be pushed.", "push_all")
	b.ResetTimer()
	for range b.N {
		m.View()
	}
}

func BenchmarkRenderRepoViewportContent(b *testing.B) {
	m := benchModel(100)
	b.ResetTimer()
	for range b.N {
		m.renderRepoViewportContent()
	}
}

func BenchmarkRenderBody(b *testing.B) {
	m := benchModel(100)
	m.syncViewports()
	b.ResetTimer()
	for range b.N {
		m.renderBody()
	}
}

func BenchmarkRenderHeaderFooter(b *testing.B) {
	m := benchModel(100)
	b.ResetTimer()
	for range b.N {
		m.renderHeader()
		m.renderFooter()
	}
}
