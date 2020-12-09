package main

import (
	color "github.com/gookit/color"
)

var (
	txtSuccess   = color.FgGreen.Render
	txtDanger    = color.FgRed.Render
	txtWarning   = color.FgYellow.Render
	txtHighlight = color.New(color.FgMagenta, color.OpBold).Render
	txtWhite     = color.FgWhite.Render
	txtWhiteBold = color.New(color.FgWhite, color.OpBold).Render
)
