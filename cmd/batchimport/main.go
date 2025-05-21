package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement import logic
			fmt.Println("Import command not yet implemented")
		},
	}

	// Add flags to import command
	importCmd.Flags().StringP("file", "f", "", "Path to the input file (required)")
	importCmd.Flags().StringP("format", "t", "csv", "Input format (csv or json)")
	importCmd.Flags().StringP("project-id", "p", "", "Ory project ID (required)")
	importCmd.MarkFlagRequired("file")
	importCmd.MarkFlagRequired("project-id")

	rootCmd.AddCommand(importCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 