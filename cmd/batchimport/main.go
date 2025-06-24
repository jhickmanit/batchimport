package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ory/batchimport/pkg/importer"
	"github.com/ory/batchimport/pkg/reader"
	"github.com/ory/batchimport/pkg/schema"
	"github.com/spf13/cobra"
)

var (
	validateOnly     bool
	errorOutputFile  string
	maxRetries       int
	batchSize        int
	apiKey           string
)

var rootCmd = &cobra.Command{
	Use:   "batchimport",
	Short: "Batch import identities into Ory Network",
	Long: `A tool for batch importing identities into Ory Network projects.
It supports CSV and JSON input formats, validates data against your project's schema,
and handles the import process in configurable batch sizes.`,
}

func init() {
	// Add import command
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import identities from a file",
		Long:  `Import identities from a CSV or JSON file into your Ory Network project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			format, _ := cmd.Flags().GetString("format")
			projectID, _ := cmd.Flags().GetString("project-id")

			// Create schema fetcher
			fetcher := schema.NewFetcher(apiKey, projectID)
			schema, err := fetcher.FetchSchema()
			if err != nil {
				return fmt.Errorf("failed to fetch schema: %w", err)
			}

			// Create reader
			fileReader := reader.NewReader(filePath, reader.Format(format), schema)

			// Validate headers
			if err := fileReader.ValidateHeaders(); err != nil {
				return fmt.Errorf("header validation failed: %w", err)
			}

			if validateOnly {
				fmt.Println("Validation successful!")
				return nil
			}

			// Create importer
			imp := importer.NewImporter(apiKey, projectID, fileReader, batchSize, maxRetries, errorOutputFile)

			// Start import process
			ctx := context.Background()
			if err := imp.Import(ctx); err != nil {
				return fmt.Errorf("import failed: %w", err)
			}

			fmt.Println("Import completed successfully!")
			return nil
		},
	}

	// Add flags to import command
	importCmd.Flags().StringP("file", "f", "", "Path to the input file (required)")
	importCmd.Flags().StringP("format", "t", "csv", "Input format (csv or json)")
	importCmd.Flags().StringP("project-id", "p", "", "Ory project ID (required)")
	importCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "Ory API key (required)")
	importCmd.Flags().BoolVar(&validateOnly, "validate-only", false, "Only validate the input file without importing")
	importCmd.Flags().StringVar(&errorOutputFile, "error-file", "", "File to write failed identities to")
	importCmd.Flags().IntVar(&maxRetries, "max-retries", 3, "Maximum number of retries for failed batches")
	importCmd.Flags().IntVar(&batchSize, "batch-size", 1000, "Number of identities to import in each batch")

	importCmd.MarkFlagRequired("file")
	importCmd.MarkFlagRequired("project-id")
	importCmd.MarkFlagRequired("api-key")

	rootCmd.AddCommand(importCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 