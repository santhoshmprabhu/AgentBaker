package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/Azure/azure-kusto-go/kusto/kql"
)

func main() {

	// Kusto variables
	kustoTable := os.Getenv("BUILD_PERFORMANCE_TABLE_NAME")
	kustoEndpoint := os.Getenv("BUILD_PERFORMANCE_KUSTO_ENDPOINT")
	kustoDatabase := os.Getenv("BUILD_PERFORMANCE_DATABASE_NAME")
	kustoClientID := os.Getenv("BUILD_PERFORMANCE_CLIENT_ID")
	// Build data variables
	sigImageName := os.Getenv("SIG_IMAGE_NAME")
	buildPerformanceDataFile := sigImageName + "-build-performance.json"
	sourceBranch := os.Getenv("GIT_BRANCH")

	var err error

	fmt.Printf("\nRunning build performance program for %s...\n\n", sigImageName)

	// Create Connection String
	kcsb := kusto.NewConnectionStringBuilder(kustoEndpoint).WithUserManagedIdentity(kustoClientID)

	// Create  Client
	client, err := kusto.New(kcsb)
	if err != nil {
		log.Fatalf("Kusto ingestion client could not be created.")
	} else {
		fmt.Printf("Created ingestion client...\n\n")
	}
	defer client.Close()

	// Create Context
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	if sourceBranch == "refs/heads/zb/ingestBuildPerfData" {

		fmt.Printf("Branch is %s, ingesting data.\n", sourceBranch)

		// Create Ingestor
		ingestor, err := ingest.New(client, kustoDatabase, kustoTable)
		if err != nil {
			client.Close()
			log.Fatalf("Kusto ingestor could not be created.")
		} else {
			fmt.Printf("Created ingestor...\n\n")
		}
		defer ingestor.Close()

		// Ingest Data
		_, err = ingestor.FromFile(ctx, buildPerformanceDataFile, ingest.IngestionMappingRef("buildPerfMap", ingest.MultiJSON))
		if err != nil {
			fmt.Printf("Ingestion failed: %v\n\n", err)
			ingestor.Close()
			client.Close()
			cancel()
			log.Fatalf("Igestion command failed to be sent.\n")
		} else {
			fmt.Printf("Ingestion started successfully.\n\n")
		}
		defer ingestor.Close()

		fmt.Printf("Successfully ingested build performance data.\n")
	}

	// Create query regardless of the branch
	query := kql.New("get_perf_data | project SIG_IMAGE_NAME, DATA | where SIG_IMAGE_NAME == SKU")

	params := kql.NewParameters().AddString("SKU", sigImageName).AddString("DATA", "BUILD_PERFORMANCE")

	result, err := client.Query(ctx, kustoDatabase, query, kusto.QueryParameters(params))
	if err != nil {
		fmt.Printf("Failed to query build performance data for %s.\n\n", sigImageName)
	}

}
