# Ory Batch Import

A Go-based tool for batch importing identities into Ory Network projects. This tool helps you validate and import large sets of identity data from CSV or JSON files into your Ory Network project.

## Features

- Support for CSV and JSON input formats
- Schema validation against Ory Network project configuration
- Duplicate detection
- Batch processing with configurable batch sizes
- Error handling and reporting
- Progress tracking

## Prerequisites

- Go 1.21 or later
- Ory Network project with API access
- Valid Ory API credentials

## Installation

```bash
go install github.com/ory/batchimport@latest
```

## Usage

```bash
# Import from CSV file
batchimport import --file users.csv --format csv --project-id your-project-id

# Import from JSON file
batchimport import --file users.json --format json --project-id your-project-id
```

## Configuration

The tool can be configured using environment variables or command-line flags:

- `ORY_API_KEY`: Your Ory API key
- `ORY_PROJECT_ID`: Your Ory project ID
- `BATCH_SIZE`: Number of identities to process in each batch (default: 800)

## Development

1. Clone the repository:

```bash
git clone https://github.com/ory/batchimport.git
cd batchimport
```

2. Install dependencies:

```bash
go mod download
```

3. Build the project:

```bash
go build
```

## License

Apache 2.0
