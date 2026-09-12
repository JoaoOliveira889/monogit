package tui

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// pendingSequenceTimeout bounds how long a half-typed sequence waits for its
// next key before it is abandoned, mirroring Vim's timeoutlen.
const pendingSequenceTimeout = 1200 * time.Millisecond

// pendingInput holds a motion that is still being typed: the numeric count so
// far and the operator prefix, if any. "3gg" arrives as three separate key
// messages, so the parts have to be remembered between them.
type pendingInput struct {
	count     int
	hasCount  bool
	prefix    string
	startedAt time.Time
}

func (p pendingInput) empty() bool {
	return !p.hasCount && p.prefix == ""
}

// expired reports whether a partially typed sequence has gone stale.
func (p pendingInput) expired(now time.Time) bool {
	if p.empty() {
		return false
	}
	return now.Sub(p.startedAt) > pendingSequenceTimeout
}

func (p pendingInput) describe() string {
	var sb strings.Builder
	if p.hasCount {
		sb.WriteString(strconv.Itoa(p.count))
	}
	sb.WriteString(p.prefix)
	return sb.String()
}

func (m *Model) clearPending() {
	m.pending = pendingInput{}
}

// countOr returns the typed count, or fallback when none was typed. Counts are
// capped so a mistyped "9999999j" cannot stall the event loop.
func (m *Model) countOr(fallback int) int {
	if !m.pending.hasCount {
		return fallback
	}
	if m.pending.count > maxMotionCount {
		return maxMotionCount
	}
	return m.pending.count
}

// prefixKeys maps an operator prefix to the keys that can follow it. It drives
// both sequence resolution and the which-key hint panel.
var prefixKeys = map[string][]struct{ Key, Action string }{
	"g": {
		{"g", "jump to first repository"},
		{"l", "toggle graph / plain log"},
		{"d", "toggle diff view"},
	},
	"ctrl+w": {
		{"1", "focus repositories"},
		{"2", "focus details"},
		{"3", "focus diff"},
		{"w", "cycle panels"},
	},
}

// consumePending folds a key into a partially typed sequence. It reports the
// resolved sequence and whether the key was absorbed rather than acted on.
func (m *Model) consumePending(msg tea.KeyMsg) (resolved string, absorbed bool) {
	key := msg.String()
	now := time.Now()

	if m.pending.expired(now) {
		m.clearPending()
	}

	// Digits build the count, except a leading zero, which is not a count.
	if len(key) == 1 && key[0] >= '0' && key[0] <= '9' && m.pending.prefix == "" {
		if key == "0" && !m.pending.hasCount {
			return "", false
		}
		digit := int(key[0] - '0')
		if !m.pending.hasCount {
			m.pending = pendingInput{hasCount: true, count: digit, startedAt: now}
			return "", true
		}
		if m.pending.count <= maxMotionCount {
			m.pending.count = m.pending.count*10 + digit
		}
		return "", true
	}

	if m.pending.prefix != "" {
		sequence := m.pending.prefix + key
		m.pending.prefix = ""
		return sequence, false
	}

	if _, ok := prefixKeys[key]; ok {
		m.pending.prefix = key
		if m.pending.startedAt.IsZero() {
			m.pending.startedAt = now
		}
		return "", true
	}

	return key, false
}
