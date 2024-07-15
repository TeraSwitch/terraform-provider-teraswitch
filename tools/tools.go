//go:build tools

package tools

import (
	// API client generation
	_ "github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen"
	// Documentation generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
