package theme

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/table"
)

func (t Theme) Help() help.Styles {
	p := t.Palette
	s := help.DefaultStyles(t.Dark)
	s.ShortKey = t.Keycap
	s.FullKey = t.Keycap
	s.ShortDesc = s.ShortDesc.Foreground(p.TextSecondary)
	s.FullDesc = s.FullDesc.Foreground(p.TextSecondary)
	s.ShortSeparator = s.ShortSeparator.Foreground(p.TextFaint)
	s.FullSeparator = s.FullSeparator.Foreground(p.TextFaint)
	s.Ellipsis = s.Ellipsis.Foreground(p.TextFaint)
	return s
}

func (t Theme) Table() table.Styles {
	p := t.Palette
	s := table.DefaultStyles()
	s.Header = s.Header.Foreground(p.Accent)
	s.Cell = s.Cell.Foreground(p.Text)
	s.Selected = s.Selected.Foreground(p.Accent).Background(p.SurfaceActive)
	return s
}

func (t Theme) ListItem() list.DefaultItemStyles {
	p := t.Palette
	s := list.NewDefaultDelegate().Styles
	s.NormalTitle = s.NormalTitle.Foreground(p.Text)
	s.NormalDesc = s.NormalDesc.Foreground(p.TextSecondary)
	s.SelectedTitle = s.SelectedTitle.Foreground(p.Accent).BorderForeground(p.Accent).Background(p.SurfaceActive)
	s.SelectedDesc = s.SelectedDesc.Foreground(p.TextSecondary).BorderForeground(p.Accent).Background(p.SurfaceActive)
	s.DimmedTitle = s.DimmedTitle.Foreground(p.TextSecondary)
	s.DimmedDesc = s.DimmedDesc.Foreground(p.TextFaint)
	return s
}
