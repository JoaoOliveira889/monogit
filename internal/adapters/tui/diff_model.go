package tui

import (
	"strconv"
	"strings"
)

// diffLineKind classifies one row of a unified diff.
type diffLineKind int

const (
	diffContext diffLineKind = iota
	diffAdded
	diffRemoved
	diffHunkHeader
	diffFileHeader
	diffMeta
)

// diffLine is one rendered row: the text plus the line numbers it carries on
// each side. A zero number means the row does not exist on that side.
type diffLine struct {
	Kind    diffLineKind
	Text    string
	OldLine int
	NewLine int
	// HunkIndex is the index of the hunk this line belongs to, or -1 outside
	// any hunk. Hunk motions use it to find the next header.
	HunkIndex int
}

// parsedDiff is a unified diff broken into rows with line numbers resolved.
type parsedDiff struct {
	Lines []diffLine
	// HunkStarts indexes into Lines, one entry per hunk header.
	HunkStarts []int
	Added      int
	Removed    int
}

var hunkHeaderPrefix = "@@"

// parseUnifiedDiff turns `git diff` output into numbered rows. Malformed or
// unparseable headers degrade to rows without numbers rather than failing, so a
// diff always renders something.
func parseUnifiedDiff(raw string) parsedDiff {
	var out parsedDiff
	if strings.TrimSpace(raw) == "" {
		return out
	}

	oldLine, newLine := 0, 0
	hunkIndex := -1

	for _, text := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(text, hunkHeaderPrefix):
			hunkIndex++
			out.HunkStarts = append(out.HunkStarts, len(out.Lines))
			oldLine, newLine = parseHunkStarts(text)
			out.Lines = append(out.Lines, diffLine{
				Kind: diffHunkHeader, Text: text, HunkIndex: hunkIndex,
			})

		case strings.HasPrefix(text, "diff --git"),
			strings.HasPrefix(text, "index "),
			strings.HasPrefix(text, "new file"),
			strings.HasPrefix(text, "deleted file"),
			strings.HasPrefix(text, "similarity index"),
			strings.HasPrefix(text, "rename "),
			strings.HasPrefix(text, "old mode"),
			strings.HasPrefix(text, "new mode"):
			out.Lines = append(out.Lines, diffLine{Kind: diffMeta, Text: text, HunkIndex: hunkIndex})

		case strings.HasPrefix(text, "--- "), strings.HasPrefix(text, "+++ "):
			out.Lines = append(out.Lines, diffLine{Kind: diffFileHeader, Text: text, HunkIndex: hunkIndex})

		case strings.HasPrefix(text, "+"):
			out.Added++
			out.Lines = append(out.Lines, diffLine{
				Kind: diffAdded, Text: text[1:], NewLine: newLine, HunkIndex: hunkIndex,
			})
			newLine++

		case strings.HasPrefix(text, "-"):
			out.Removed++
			out.Lines = append(out.Lines, diffLine{
				Kind: diffRemoved, Text: text[1:], OldLine: oldLine, HunkIndex: hunkIndex,
			})
			oldLine++

		case strings.HasPrefix(text, "\\"):
			// "\ No newline at end of file" belongs to neither side.
			out.Lines = append(out.Lines, diffLine{Kind: diffMeta, Text: text, HunkIndex: hunkIndex})

		default:
			// Context rows advance both sides. A leading space is the unified
			// diff marker and is not part of the content.
			content := strings.TrimPrefix(text, " ")
			out.Lines = append(out.Lines, diffLine{
				Kind: diffContext, Text: content, OldLine: oldLine, NewLine: newLine, HunkIndex: hunkIndex,
			})
			oldLine++
			newLine++
		}
	}

	return out
}

// parseHunkStarts reads the starting line numbers out of "@@ -a,b +c,d @@".
func parseHunkStarts(header string) (oldStart, newStart int) {
	fields := strings.Fields(header)
	for _, f := range fields {
		switch {
		case strings.HasPrefix(f, "-"):
			oldStart = leadingNumber(f[1:])
		case strings.HasPrefix(f, "+"):
			newStart = leadingNumber(f[1:])
		}
	}
	return oldStart, newStart
}

func leadingNumber(s string) int {
	if comma := strings.IndexByte(s, ','); comma >= 0 {
		s = s[:comma]
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// hunkAfter returns the index into Lines of the first hunk header strictly
// after the given row, or -1 when there is none.
func (d parsedDiff) hunkAfter(row int) int {
	for _, start := range d.HunkStarts {
		if start > row {
			return start
		}
	}
	return -1
}

// hunkBefore returns the index into Lines of the last hunk header strictly
// before the given row, or -1 when there is none.
func (d parsedDiff) hunkBefore(row int) int {
	found := -1
	for _, start := range d.HunkStarts {
		if start < row {
			found = start
			continue
		}
		break
	}
	return found
}
