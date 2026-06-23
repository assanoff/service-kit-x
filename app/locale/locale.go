// Package locale embeds the example's message catalogs and builds the
// servicekit i18n.Translator used to localize error responses.
package locale

import (
	"embed"

	"github.com/assanoff/servicekit/i18n"
)

//go:embed locales/en.json locales/ru.json
var catalogs embed.FS

// New builds the application translator (default English, plus Russian).
func New() (*i18n.Translator, error) {
	return i18n.New("en", catalogs, "locales/en.json", "locales/ru.json")
}
