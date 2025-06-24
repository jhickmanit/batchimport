package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/ory/batchimport/pkg/reader"
	ory "github.com/ory/client-go"
)

// Importer handles the batch import process
type Importer struct {
	client        *ory.APIClient
	projectID     string
	reader        *reader.Reader
	batchSize     int
	maxRetries    int
	errorFile     string
	errorChan     chan ImportError
	errorWg       sync.WaitGroup
}

// ImportError represents an error during the import process
type ImportError struct {
	Identity map[string]interface{}
	Error    error
}

// NewImporter creates a new importer instance
func NewImporter(apiKey, projectID string, reader *reader.Reader, batchSize, maxRetries int, errorFile string) *Importer {
	config := ory.NewConfiguration()
	config.Servers = ory.ServerConfigurations{
		{
			URL: "https://api.ory.sh",
		},
	}
	config.AddDefaultHeader("Authorization", "Bearer "+apiKey)

	return &Importer{
		client:     ory.NewAPIClient(config),
		projectID:  projectID,
		reader:     reader,
		batchSize:  batchSize,
		maxRetries: maxRetries,
		errorFile:  errorFile,
		errorChan:  make(chan ImportError, 1000),
	}
}

// StartErrorWriter starts the error writer goroutine
func (i *Importer) StartErrorWriter() {
	if i.errorFile == "" {
		return
	}

	i.errorWg.Add(1)
	go func() {
		defer i.errorWg.Done()
		file, err := os.Create(i.errorFile)
		if err != nil {
			fmt.Printf("Failed to create error file: %v\n", err)
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		for err := range i.errorChan {
			if err := encoder.Encode(err); err != nil {
				fmt.Printf("Failed to write error to file: %v\n", err)
			}
		}
	}()
}

// StopErrorWriter stops the error writer and waits for it to finish
func (i *Importer) StopErrorWriter() {
	if i.errorFile != "" {
		close(i.errorChan)
		i.errorWg.Wait()
	}
}

// ImportBatch imports a batch of identities
func (i *Importer) ImportBatch(ctx context.Context, identities []map[string]interface{}) error {
	var retries int
	for retries < i.maxRetries {
		// Create identities one by one since batch creation is not supported
		for _, identity := range identities {
			// Create the identity request
			req := ory.NewCreateIdentityBody("default", identity)

			// Execute the import
			_, resp, err := i.client.IdentityAPI.CreateIdentity(ctx).CreateIdentityBody(*req).Execute()
			if err != nil {
				// Handle errors
				if resp != nil && resp.StatusCode >= 500 {
					retries++
					continue
				}

				// For non-retryable errors, write to error file
				i.errorChan <- ImportError{
					Identity: identity,
					Error:    err,
				}
				return err
			}
		}

		return nil
	}

	return fmt.Errorf("failed to import batch after %d retries", i.maxRetries)
}

// Import processes the input file and imports identities in batches
func (i *Importer) Import(ctx context.Context) error {
	i.StartErrorWriter()
	defer i.StopErrorWriter()

	for {
		// Read a batch of identities
		batch, err := i.reader.ReadBatch(i.batchSize)
		if err != nil {
			return fmt.Errorf("failed to read batch: %w", err)
		}

		// If no identities were read, we're done
		if len(batch) == 0 {
			break
		}

		// Import the batch
		if err := i.ImportBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to import batch: %w", err)
		}

		fmt.Printf("Successfully imported batch of %d identities\n", len(batch))
	}

	return nil
} 