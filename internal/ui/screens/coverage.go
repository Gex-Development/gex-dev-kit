package screens

import "github.com/Gex-Development/gex-dev-kit/internal/ui/theme"

type Coverage struct {
	Lib        string
	Feature    string
	ScreenFile string
	Call       string
	Page       string
	Demo       func(t theme.Theme, width int) string
	Marker     string
	TestOnly   bool
}

var CoverageMap []Coverage
