package reader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ory/batchimport/pkg/schema"
)

// Format represents the supported file formats
type Format string

const (
	FormatCSV  Format = "csv"
	FormatJSON Format = "json"
)

// Reader handles reading and validating input files
type Reader struct {
	filePath string
	format   Format
	schema   *schema.Schema
}

// NewReader creates a new file reader
func NewReader(filePath string, format Format, schema *schema.Schema) *Reader {
	return &Reader{
		filePath: filePath,
		format:   format,
		schema:   schema,
	}
}

// ValidateHeaders validates the headers/keys against the schema
func (r *Reader) ValidateHeaders() error {
	headers, err := r.getHeaders()
	if err != nil {
		return fmt.Errorf("failed to get headers: %w", err)
	}

	validator := schema.NewSchemaValidator(r.schema)
	return validator.ValidateHeaders(headers)
}

// getHeaders returns the headers from the file based on its format
func (r *Reader) getHeaders() ([]string, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	switch r.format {
	case FormatCSV:
		return r.getCSVHeaders(file)
	case FormatJSON:
		return r.getJSONHeaders(file)
	default:
		return nil, fmt.Errorf("unsupported format: %s", r.format)
	}
}

// getCSVHeaders reads the headers from a CSV file
func (r *Reader) getCSVHeaders(file *os.File) ([]string, error) {
	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}
	return headers, nil
}

// getJSONHeaders reads the keys from the first object in a JSON file
func (r *Reader) getJSONHeaders(file *os.File) ([]string, error) {
	decoder := json.NewDecoder(file)
	
	// Read the first token to determine if it's an array or object
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON token: %w", err)
	}

	var headers []string
	switch token {
	case json.Delim('['):
		// It's an array, read the first object
		var firstObject map[string]interface{}
		if err := decoder.Decode(&firstObject); err != nil {
			return nil, fmt.Errorf("failed to decode first JSON object: %w", err)
		}
		headers = make([]string, 0, len(firstObject))
		for key := range firstObject {
			headers = append(headers, key)
		}
	case json.Delim('{'):
		// It's a single object
		var object map[string]interface{}
		if err := decoder.Decode(&object); err != nil {
			return nil, fmt.Errorf("failed to decode JSON object: %w", err)
		}
		headers = make([]string, 0, len(object))
		for key := range object {
			headers = append(headers, key)
		}
	default:
		return nil, fmt.Errorf("invalid JSON format: expected array or object")
	}

	return headers, nil
}

// DetectFormat attempts to detect the file format based on extension
func DetectFormat(filePath string) (Format, error) {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".csv":
		return FormatCSV, nil
	case ".json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
} 