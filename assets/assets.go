package assets

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var rawVersion string

//go:embed banner.txt
var rawBanner string

func Version() string {
	return strings.TrimSpace(rawVersion)
}

func Banner() string {
	b := strings.ReplaceAll(rawBanner, "\\033", "\x1b")
	b += " \x1b[38;2;200;200;200mHUNT CLI | v" + Version() + "\x1b[0m\n"
	return b
}
