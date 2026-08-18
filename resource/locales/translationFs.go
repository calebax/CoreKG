package locales

import (
	"embed"
)

//go:embed *.yaml
var TranslationFs embed.FS
