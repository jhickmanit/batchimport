package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ory/jsonschema/v3"
)

// Schema represents the Ory Network project schema
type Schema struct {
	ID     string         `json:"$id"`
	Schema string         `json:"$schema"`
	Title  string         `json:"title"`
	Type   string         `json:"type"`
	Properties Properties `json:"properties"`
}

type Properties struct {
	Traits TraitProperties `json:"traits"`
}

type TraitProperties struct {
	Type                 string                 `json:"type"`
	Properties          map[string]TraitField  `json:"properties"`
	Required            []string               `json:"required"`
	AdditionalProperties bool                  `json:"additionalProperties"`
}

type TraitField struct {
	Type      string                 `json:"type"`
	Format    string                 `json:"format,omitempty"`
	Title     string                 `json:"title,omitempty"`
	MaxLength int                    `json:"maxLength,omitempty"`
	Properties map[string]TraitField `json:"properties,omitempty"`
	Credentials map[string]any       `json:"ory.sh/kratos,omitempty"`
}

// SchemaValidator provides methods for schema validation
type SchemaValidator struct {
	schema *Schema
	compiler *jsonschema.Compiler
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(schema *Schema) *SchemaValidator {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true

	return &SchemaValidator{
		schema: schema,
		compiler: compiler,
	}
}

// ValidateHeaders checks if the provided headers match the schema properties
func (v *SchemaValidator) ValidateHeaders(headers []string) error {
	// Get the traits properties
	traits := v.schema.Properties.Traits.Properties
	if traits == nil {
		return fmt.Errorf("invalid schema format: traits properties not found")
	}

	// Create a map of schema properties for quick lookup
	schemaProps := make(map[string]bool)
	for prop := range traits {
		schemaProps[prop] = true
	}

	// Check if all headers exist in the schema
	for _, header := range headers {
		if !schemaProps[header] {
			return fmt.Errorf("header '%s' not found in schema", header)
		}
	}

	// Check if all required fields are present
	for _, required := range v.schema.Properties.Traits.Required {
		found := false
		for _, header := range headers {
			if header == required {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required field '%s' not found in headers", required)
		}
	}

	return nil
}

// ValidateTrait validates a trait value against its schema definition
func (v *SchemaValidator) ValidateTrait(name string, value any) error {
	trait, exists := v.schema.Properties.Traits.Properties[name]
	if !exists {
		return fmt.Errorf("trait '%s' not found in schema", name)
	}

	// Create a schema for the specific trait
	traitSchema := map[string]any{
		"$schema": v.schema.Schema,
		"type":    trait.Type,
		"format":  trait.Format,
		"title":   trait.Title,
	}

	if trait.MaxLength > 0 {
		traitSchema["maxLength"] = trait.MaxLength
	}

	// Add nested properties if they exist
	if len(trait.Properties) > 0 {
		traitSchema["properties"] = trait.Properties
	}

	// Convert the trait schema to JSON
	schemaBytes, err := json.Marshal(traitSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal trait schema: %w", err)
	}

	// Add the schema to the compiler
	if err := v.compiler.AddResource(name, strings.NewReader(string(schemaBytes))); err != nil {
		return fmt.Errorf("failed to add schema to compiler: %w", err)
	}

	// Get the compiled schema
	compiledSchema, err := v.compiler.Compile(context.Background(), name)
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}

	// Convert the value to JSON for validation
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Validate the value against the schema
	if err := compiledSchema.Validate(strings.NewReader(string(valueBytes))); err != nil {
		return fmt.Errorf("validation failed for trait '%s': %w", name, err)
	}

	return nil
} 