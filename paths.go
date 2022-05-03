package defaults

import (
	"os"
	"path/filepath"
)

var (
	// ResourcesDir is the base path of where all static resources like templates and assets are stored
	// This is dynamically set to the current working directory on startup
	ResourcesDir = ""

	// TemplatePrefix is a prefix that is appended to the ResourcesDir to get the base path for all templates
	// This is used by the default renderer implementation to figure out where to look for partials.
	TemplatePrefix = "/templates/"
)

func init() {
	ResourcesDir, _ = os.Getwd()
	ResourcesDir += "/resources/"
}

// TemplateDir will return the absolute path of the configured templates directory
func TemplateDir() string {
	path, err := filepath.Abs(ResourcesDir + TemplatePrefix)
	if err != nil {
		return ResourcesDir + "/templates"
	}
	return path
}
