package cmd

import (
	"github.com/fatih/color"
)

type Command interface {
	Execute()
}

type HelpCommand struct{}

type StatusCommand struct{}

type LoginCommad struct{}

var (
	red      = color.RGB(255, 0, 0)
	green    = color.RGB(0, 255, 0)
	fpYellow = color.RGB(255, 249, 87)
	yellow   = color.RGB(255, 255, 0)
)
