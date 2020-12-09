package main

import (
	color "github.com/gookit/color"
)

var (
	TxtSuccess   = color.FgGreen.Render
	TxtDanger    = color.FgRed.Render
	TxtWarning   = color.FgYellow.Render
	TxtHighlight = color.New(color.FgMagenta, color.OpBold).Render
	TxtWhite     = color.FgWhite.Render
	TxtWhiteBold = color.New(color.FgWhite, color.OpBold).Render
)
