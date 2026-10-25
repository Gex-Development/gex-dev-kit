package components

import (
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

type ChatRole uint8

const (
	ChatAssistant ChatRole = iota
	ChatUser
	ChatNote
	ChatError
)

func ChatMessage(t theme.Theme, width int, role ChatRole, text string) string {
	if width <= 0 {
		return ""
	}
	prefix := "  "
	if role == ChatAssistant {
		prefix = "◆ "
		if t.Plain {
			prefix = "* "
		}
	}
	if role == ChatUser {
		prefix = "│ "
		if t.Plain {
			prefix = "| "
		}
	}
	if role == ChatError {
		prefix = "× "
		if t.Plain {
			prefix = "x "
		}
	}
	wrapped := ansi.Wrap(text, max(1, width-2), "")
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		p := prefix
		if i > 0 && role != ChatUser {
			p = "  "
		}
		if width < 2 {
			p = ansi.Truncate(p, width, "")
			line = ""
		} else {
			line = ansi.Truncate(line, width-2, "")
		}
		switch role {
		case ChatAssistant:
			if i == 0 {
				p = t.Accent.Render(p)
			}
			line = t.Text.Render(line)
		case ChatUser:
			p = t.UserBar.Render(p)
			line = t.Strong.Render(line)
		case ChatNote:
			line = t.Secondary.Render(line)
		case ChatError:
			if i == 0 {
				p = t.Error.Render(p)
				line = t.Error.Render(line)
			} else {
				line = t.Text.Render(line)
			}
		}
		lines[i] = p + line
	}
	return strings.Join(lines, "\n")
}
