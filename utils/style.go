package utils

import (
	color "github.com/gookit/color"
)

var (
	StyleSuccess   = color.FgGreen.Render
	StyleDanger    = color.FgRed.Render
	StyleWarning   = color.FgYellow.Render
	StyleHighlight = color.New(color.FgMagenta, color.OpBold).Render
	StyleWhite     = color.FgWhite.Render
	StyleWhiteBold = color.New(color.FgWhite, color.OpBold).Render
)
