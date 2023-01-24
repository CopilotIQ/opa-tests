package internals

import (
	"fmt"
	"github.com/CopilotIQ/opa-tests/gentests"
)

func CreateBundle(manifestPath string, srcDir string) (string, error) {
	var manifest = gentests.ReadManifest(manifestPath)
	if manifest == nil {
		return "", fmt.Errorf("cannot load manifest %s", manifestPath)
	}

	return "bundle", nil
}
