package main

import (
	color "github.com/gookit/color"
)

var (
	styleSuccess   = color.FgGreen.Render
	styleDanger    = color.FgRed.Render
	styleWarning   = color.FgYellow.Render
	styleHighlight = color.New(color.FgMagenta, color.OpBold).Render
	styleWhite     = color.FgWhite.Render
	styleWhiteBold = color.New(color.FgWhite, color.OpBold).Render
)
