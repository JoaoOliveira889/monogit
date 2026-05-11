package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"monogit/internal/pkg/ui"
)

func (m *Model) renderHeader() string {
	title := " MonoGit "
	stats := fmt.Sprintf("%d repos ", len(m.repos))

	loading := ""
	if m.isBusy() {
		loading = ui.SpinnerStyle.Render(m.spinnerView() + " Loading...")
	}

	header := ui.HeaderStyle.Render(title)
	statsStr := ui.SubtleStyle.Render(stats)

	spacerLen := m.width - lipgloss.Width(header) - lipgloss.Width(statsStr) - lipgloss.Width(loading)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		header,
		spacer,
		loading,
		statsStr,
	)
}

func (m *Model) renderFooter() string {
	var parts []string
	sep := ui.SubtleStyle.Render(" • ")

	if m.showConfirmModal {
		parts = []string{
			m.fmtKey("y", "yes"),
			m.fmtKey("n", "no"),
			m.fmtKey("esc", "cancel"),
		}
	} else if m.showHelp {
		parts = []string{
			m.fmtKey("esc/ctrl+p", "close"),
		}
	} else if m.activePanel == CommitWizardPanel {
		switch m.commitStep {
		case StepAddOption:
			parts = []string{
				m.fmtKey("a", "add all"),
				m.fmtKey("s", "select files"),
				m.fmtKey("esc", "cancel"),
			}
		case StepSelectFiles:
			parts = []string{
				m.fmtKey("space", "toggle"),
				m.fmtKey("enter", "done"),
				m.fmtKey("x", "discard"),
				m.fmtKey("esc", "back"),
			}
		case StepMessage:
			parts = []string{
				m.fmtKey("enter", "commit"),
				m.fmtKey("esc", "cancel"),
			}
		case StepPull:
			parts = []string{
				m.fmtKey("y", "pull"),
				m.fmtKey("n", "skip"),
			}
		case StepPush:
			parts = []string{
				m.fmtKey("y", "push"),
				m.fmtKey("n", "skip"),
			}
		}
	} else if m.showFiles {
		if m.activePanel == DiffPanel {
			parts = []string{
				m.fmtKey("jk", "scroll"),
				m.fmtKey("tab/2", "files"),
				m.fmtKey("1", "repos"),
				m.fmtKey("ctrl+p", "help"),
			}
		} else {
			parts = []string{
				m.fmtKey("jk", "nav"),
				m.fmtKey("space", "select"),
				m.fmtKey("x", "discard"),
				m.fmtKey("a/n", "all/none"),
				m.fmtKey("enter", "done"),
				m.fmtKey("tab/3", "diff"),
			}
		}
	} else if m.showBranches {
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey("enter", "checkout"),
			m.fmtKey("n", "new"),
			m.fmtKey("esc", "back"),
		}
	} else {
		if m.activePanel == RepoPanel {
			parts = []string{
				m.fmtKey("hjkl", "nav"),
				m.fmtKey("f/F", "fetch"),
				m.fmtKey("p/P", "pull"),
				m.fmtKey("u/U", "push"),
				m.fmtKey("c", "commit"),
				m.fmtKey("b", "branches"),
				m.fmtKey("o", "logs"),
				m.fmtKey("ctrl+p", "help"),
			}
		} else {
			parts = []string{
				m.fmtKey("jk", "scroll"),
				m.fmtKey("x", "undo"),
				m.fmtKey("g", "graph"),
				m.fmtKey("z/Z", "stash"),
				m.fmtKey("1", "repos"),
			}
		}
	}

	keys := strings.Join(parts, sep)

	versionStr := ui.SubtleStyle.Render(Version)
	rightSide := versionStr

	spacerWidth := m.width - lipgloss.Width(keys) - lipgloss.Width(rightSide) - 4
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := strings.Repeat(" ", spacerWidth)

	return ui.FooterStyle.Width(m.width).Render(
		fmt.Sprintf(" %s%s%s ", keys, spacer, rightSide),
	)
}

func (m *Model) fmtKey(k, action string) string {
	return ui.FooterKeyStyle.Render(k) + " " + ui.FooterActionStyle.Render(action)
}

func (m *Model) renderConfirmationModal() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		ui.PanelTitleStyle.Render(" Confirmation "),
		"",
		ui.ValueStyle.Render(m.confirmModalTitle),
		"",
		lipgloss.JoinHorizontal(lipgloss.Center,
			m.fmtKey("y", "Yes"),
			"   ",
			m.fmtKey("n", "No"),
		),
	)
}

func (m *Model) renderCommitWizardModal() string {
	var content string
	title := " Commit Wizard "

	switch m.commitStep {
	case StepAddOption:
		content = "How would you like to stage changes?\n\n" +
			m.fmtKey("a", "Add all files") + "\n" +
			m.fmtKey("s", "Select files manually")
	case StepSelectFiles:
		content = "Select files in the right panel.\n" +
			m.fmtKey("space", "Toggle") + "  " +
			m.fmtKey("enter", "Done") + "  " +
			m.fmtKey("esc", "Cancel")
	case StepMessage:
		content = "Commit message:\n\n" + ui.InputStyle.Render(m.commitInput.View())
	case StepPull:
		content = "Commit successful!\n\nPull changes from remote?\n\n" +
			m.fmtKey("y", "Yes, pull") + "\n" +
			m.fmtKey("n", "No, skip")
	case StepPush:
		content = "Ready to push?\n\n" +
			m.fmtKey("y", "Yes, push") + "\n" +
			m.fmtKey("n", "No, skip")
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		ui.PanelTitleStyle.Render(title),
		"",
		content,
	)
}

func (m *Model) renderHelpMenu() string {
	var sections [][]string

	sections = append(sections, []string{
		ui.PanelTitleStyle.Render(" NAVIGATION "),
		ui.LabelStyle.Render("  jk / arrows:") + "      Move selection",
		ui.LabelStyle.Render("  h / l / arrows:") + "   Switch panels",
		ui.LabelStyle.Render("  1 / 2 / 3:") + "         Jump to panel",
		ui.LabelStyle.Render("  tab:") + "               Cycle focus",
	})

	sections = append(sections, []string{
		ui.PanelTitleStyle.Render(" REPOSITORY "),
		ui.LabelStyle.Render("  f / F:") + "             Fetch (one / all)",
		ui.LabelStyle.Render("  p / P:") + "             Pull (one / all)",
		ui.LabelStyle.Render("  u / U:") + "             Push (one / all)",
		ui.LabelStyle.Render("  c:") + "                 Commit wizard",
		ui.LabelStyle.Render("  b:") + "                 List branches",
		ui.LabelStyle.Render("  z / S:") + "             Stash / Pop stash",
	})

	sections = append(sections, []string{
		ui.PanelTitleStyle.Render(" FILES & DIFF "),
		ui.LabelStyle.Render("  space:") + "            Toggle selection",
		ui.LabelStyle.Render("  x:") + "                Discard changes",
		ui.LabelStyle.Render("  a / n:") + "            Select all / none",
		ui.LabelStyle.Render("  g:") + "                Toggle graph view",
	})

	sections = append(sections, []string{
		ui.PanelTitleStyle.Render(" GENERAL "),
		ui.LabelStyle.Render("  ? / ctrl+p:") + "        Toggle help",
		ui.LabelStyle.Render("  esc:") + "               Back / Cancel",
		ui.LabelStyle.Render("  o:") + "                 Command log",
		ui.LabelStyle.Render("  q:") + "                 Quit",
	})

	var rows []string
	for _, s := range sections {
		rows = append(rows, strings.Join(s, "\n"))
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, rows[0], "", rows[1]),
		"    ",
		lipgloss.JoinVertical(lipgloss.Left, rows[2], "", rows[3]),
	)

	return lipgloss.JoinVertical(lipgloss.Center,
		ui.HeaderStyle.Render(" MONOGIT SHORTCUTS "),
		"",
		content,
		"",
		ui.SubtleStyle.Render("Press ESC or ctrl+p to close"),
	)
}
