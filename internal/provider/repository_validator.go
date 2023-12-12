package provider

import (
	"context"
	"fmt"
	"strings"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// repositoryTypes see https://github.com/DependencyTrack/dependency-track/blob/master/src/main/java/org/dependencytrack/model/RepositoryType.java
var repositoryTypes = []string{
	dtrack.RepositoryTypeCargo,
	dtrack.RepositoryTypeComposer,
	dtrack.RepositoryTypeCpan,
	dtrack.RepositoryTypeGem,
	dtrack.RepositoryTypeGithub,
	dtrack.RepositoryTypeGoModules,
	dtrack.RepositoryTypeHex,
	dtrack.RepositoryTypeMaven,
	dtrack.RepositoryTypeNpm,
	dtrack.RepositoryTypeNuget,
	dtrack.RepositoryTypePypi,
}

type repositoryTypeValidator struct {
}

func (r repositoryTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Available Type Values: %s", strings.Join(repositoryTypes, ", "))
}

func (r repositoryTypeValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("# Available Type Values: %s\n\n- ", strings.Join(repositoryTypes, "\n- "))
}

func (r repositoryTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	tp := req.ConfigValue.ValueString()
	if tp == "" {
		return
	}
	for _, t := range repositoryTypes {
		if t == tp {
			return
		}
	}
	resp.Diagnostics.AddError(
		fmt.Sprintf("Unknown Repository Type: %q", tp),
		r.Description(ctx),
	)
}
