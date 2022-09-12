package embed

import (
    goembed "embed"
)

//go:embed templates
var embededFiles goembed.FS

// ReadTemplate read the template file embed.
func ReadTemplate(path string) ([]byte, error) {
    return embededFiles.ReadFile(path)
}
