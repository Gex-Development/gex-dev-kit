package main

import (
	"io"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func writePreview(w io.Writer, t theme.Theme, groups []components.ChangeGroup) {}

func writeProgress(errOut io.Writer, quiet bool) func(app.Step) {
	return func(app.Step) {}
}
