package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

// validateStruct is the validation structure for the configuration.
// This is used to validate the full structure of the configuration. This
// requires duplication between this struct and the other config structs
// since we don't do any lazy loading here.
type validateStruct struct {
	Project string            `hcl:"project,attr"`
	Runner  *Runner           `hcl:"runner,block" default:"{}"`
	Labels  map[string]string `hcl:"labels,optional"`
	Plugin  []*Plugin         `hcl:"plugin,block"`
	Apps    []*validateApp    `hcl:"app,block"`
}

type validateApp struct {
	Name    string            `hcl:",label"`
	Path    string            `hcl:"path,optional"`
	Labels  map[string]string `hcl:"labels,optional"`
	URL     *AppURL           `hcl:"url,block" default:"{}"`
	Build   *Build            `hcl:"build,block"`
	Deploy  *Deploy           `hcl:"deploy,block"`
	Release *Release          `hcl:"release,block"`
}

// Validate the structure of the configuration.
//
// This will validate required fields are specified and the types of some fields.
// Plugin-specific fields won't be validated until later. Fields that use functions
// and variables will not be validated until those values can be realized.
//
// Users of this package should call Validate on each subsequent configuration
// that is loaded (Apps, Builds, Deploys, etc.) for further rich validation.
func (c *Config) Validate() error {
	// Validate root
	schema, _ := gohcl.ImpliedBodySchema(&validateStruct{})
	content, diag := c.hclConfig.Body.Content(schema)
	if diag.HasErrors() {
		return diag
	}

	// Validate apps
	var result error
	for _, block := range content.Blocks.OfType("app") {
		err := c.validateApp(block)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Validate labels
	if errs := ValidateLabels(c.Labels); len(errs) > 0 {
		result = multierror.Append(result, errs...)
	}

	return result
}

func (c *Config) validateApp(b *hcl.Block) error {
	// Validate root
	schema, _ := gohcl.ImpliedBodySchema(&validateApp{})
	content, diag := b.Body.Content(schema)
	if diag.HasErrors() {
		return diag
	}

	// Build required
	if len(content.Blocks.OfType("build")) != 1 {
		return &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'build' stanza required",
			Subject:  &b.DefRange,
			Context:  &b.TypeRange,
		}
	}

	// Deploy required
	if len(content.Blocks.OfType("deploy")) != 1 {
		return &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'deploy' stanza required",
			Subject:  &b.DefRange,
			Context:  &b.TypeRange,
		}
	}

	return nil
}

// ValidateLabels validates a set of labels. This ensures that labels are
// set according to our requirements:
//
//   * key and value length can't be greater than 255 characters each
//   * keys must be in hostname format (RFC 952)
//   * keys can't be prefixed with "waypoint/" which is reserved for system use
//
func ValidateLabels(labels map[string]string) []error {
	var errs []error
	for k, v := range labels {
		name := fmt.Sprintf("label[%s]", k)

		if strings.HasPrefix(k, "waypoint/") {
			errs = append(errs, fmt.Errorf("%s: prefix 'waypoint/' is reserved for system use", name))
		}

		if len(k) > 255 {
			errs = append(errs, fmt.Errorf("%s: key must be less than or equal to 255 characters", name))
		}

		if !hostnameRegexRFC952.MatchString(strings.SplitN(k, "/", 2)[0]) {
			errs = append(errs, fmt.Errorf("%s: key before '/' must be a valid hostname (RFC 952)", name))
		}

		if len(v) > 255 {
			errs = append(errs, fmt.Errorf("%s: value must be less than or equal to 255 characters", name))
		}
	}

	return errs
}

var hostnameRegexRFC952 = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9]$`)
