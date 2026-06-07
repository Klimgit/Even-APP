package swagger

import (
	_ "embed"
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Shared component names repeated across backend specs.
var dedupPaths = []string{
	"/api/v1/health",
}

var dedupSchemas = []string{
	"ErrorResponse",
	"HealthResponse",
}

var dedupResponses = []string{
	"DefaultError",
}

var dedupSecuritySchemes = []string{
	"bearerAuth",
}

//go:embed bootstrap.yaml
var bootstrapSpec []byte

type specBuilder struct {
	pathItems *orderedmap.Map[string, *v3.PathItem]

	schemas       *orderedmap.Map[string, *base.SchemaProxy]
	responses     *orderedmap.Map[string, *v3.Response]
	parameters    *orderedmap.Map[string, *v3.Parameter]
	requestBodies *orderedmap.Map[string, *v3.RequestBody]

	securitySchemes *orderedmap.Map[string, *v3.SecurityScheme]
}

func (sb *specBuilder) init() error {
	sb.pathItems = orderedmap.New[string, *v3.PathItem]()
	sb.schemas = orderedmap.New[string, *base.SchemaProxy]()
	sb.responses = orderedmap.New[string, *v3.Response]()
	sb.parameters = orderedmap.New[string, *v3.Parameter]()
	sb.requestBodies = orderedmap.New[string, *v3.RequestBody]()
	sb.securitySchemes = orderedmap.New[string, *v3.SecurityScheme]()

	bootstrapDoc, err := unmarshalSpec(bootstrapSpec)
	if err != nil {
		return fmt.Errorf("bootstrap spec: %w", err)
	}
	if err := sb.mergeComponents(bootstrapDoc.Components); err != nil {
		return err
	}
	return sb.mergePaths(bootstrapDoc.Paths)
}

func (sb *specBuilder) add(doc *v3.Document) error {
	if doc == nil {
		return nil
	}
	if err := sb.mergePaths(doc.Paths); err != nil {
		return err
	}
	return sb.mergeComponents(doc.Components)
}

func (sb *specBuilder) mergePaths(paths *v3.Paths) error {
	if paths == nil {
		return nil
	}
	for path, item := range paths.PathItems.FromOldest() {
		if _, exists := sb.pathItems.Get(path); exists {
			if contains(dedupPaths, path) {
				continue
			}
			return fmt.Errorf("path already exists: %s", path)
		}
		sb.pathItems.Set(path, item)
	}
	return nil
}

func (sb *specBuilder) mergeComponents(components *v3.Components) error {
	if components == nil {
		return nil
	}
	for name, schema := range components.Schemas.FromOldest() {
		if _, exists := sb.schemas.Get(name); exists {
			if contains(dedupSchemas, name) {
				continue
			}
			return fmt.Errorf("schema already exists: %s", name)
		}
		sb.schemas.Set(name, schema)
	}
	for name, response := range components.Responses.FromOldest() {
		if _, exists := sb.responses.Get(name); exists {
			if contains(dedupResponses, name) {
				continue
			}
			return fmt.Errorf("response already exists: %s", name)
		}
		sb.responses.Set(name, response)
	}
	for name, parameter := range components.Parameters.FromOldest() {
		if _, exists := sb.parameters.Get(name); exists {
			return fmt.Errorf("parameter already exists: %s", name)
		}
		sb.parameters.Set(name, parameter)
	}
	for name, body := range components.RequestBodies.FromOldest() {
		if _, exists := sb.requestBodies.Get(name); exists {
			return fmt.Errorf("request body already exists: %s", name)
		}
		sb.requestBodies.Set(name, body)
	}
	for name, scheme := range components.SecuritySchemes.FromOldest() {
		if _, exists := sb.securitySchemes.Get(name); exists {
			if contains(dedupSecuritySchemes, name) {
				continue
			}
			return fmt.Errorf("security scheme already exists: %s", name)
		}
		sb.securitySchemes.Set(name, scheme)
	}
	return nil
}

func (sb *specBuilder) build() *v3.Document {
	return &v3.Document{
		Version: "3.0.3",
		Info: &base.Info{
			Title:       "even-app",
			Version:     "0.2.0",
			Description: "Merged OpenAPI specification for the Even language learning platform",
		},
		Paths: &v3.Paths{PathItems: sb.pathItems},
		Components: &v3.Components{
			Schemas:         sb.schemas,
			Responses:       sb.responses,
			Parameters:      sb.parameters,
			RequestBodies:   sb.requestBodies,
			SecuritySchemes: sb.securitySchemes,
		},
	}
}

func contains(items []string, item string) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}
