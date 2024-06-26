package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/defenseunicorns/uds-security-hub/internal/log"
	"github.com/defenseunicorns/uds-security-hub/pkg/scan"
)

var errFlagRetrieval = errors.New("error getting flag")
var errRequiredFlagEmpty = errors.New("is required and cannot be empty")

func main() {
	var rootCmd = &cobra.Command{
		Use:   "scan",
		Short: "Scan is a tool for scanning packages",
		Run:   runScanner,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			requiredFlags := []string{"ghcr-token", "org", "package-name", "tag"}
			for _, flag := range requiredFlags {
				value, err := cmd.Flags().GetString(flag)
				if err != nil {
					return fmt.Errorf("%w: %s: %w", errFlagRetrieval, flag, err)
				}
				if value == "" {
					return fmt.Errorf("%s %w", flag, errRequiredFlagEmpty)
				}
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringP("trivy-username", "u", "", "Username for Trivy")
	rootCmd.PersistentFlags().StringP("trivy-password", "p", "", "Password for Trivy")
	rootCmd.PersistentFlags().StringP("ghcr-token", "t", "", "Token for GHCR")
	rootCmd.PersistentFlags().StringP("org", "o", "", "Organization")
	rootCmd.PersistentFlags().StringP("package-name", "n", "", "Package Name")
	rootCmd.PersistentFlags().StringP("tag", "g", "", "Tag")
	rootCmd.PersistentFlags().StringP("output-file", "f", "", "Output file for CSV results")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runScanner(cmd *cobra.Command, args []string) {
	logger := log.NewLogger(context.Background())
	trivyUsername, _ := cmd.Flags().GetString("trivy-username") //nolint:errcheck
	trivyPassword, _ := cmd.Flags().GetString("trivy-password") //nolint:errcheck
	ghcrToken, _ := cmd.Flags().GetString("ghcr-token")         //nolint:errcheck
	org, _ := cmd.Flags().GetString("org")                      //nolint:errcheck
	packageName, _ := cmd.Flags().GetString("package-name")     //nolint:errcheck
	tag, _ := cmd.Flags().GetString("tag")                      //nolint:errcheck
	outputFile, _ := cmd.Flags().GetString("output-file")       //nolint:errcheck

	scanner, err := scan.New(context.Background(), logger, trivyUsername, trivyPassword, ghcrToken)
	if err != nil {
		logger.Fatalf("Error creating scanner: %v", err)
	}
	results, err := scanner.ScanZarfPackage(org, packageName, tag)
	if err != nil {
		logger.Fatalf("Error scanning package: %v", err)
	}

	var combinedCSV string
	for _, v := range results {
		r, err := scanner.ScanResultReader(v)
		if err != nil {
			logger.Fatalf("Error scanning: %v", err)
		}

		csv := r.GetResultsAsCSV()
		combinedCSV += csv
	}

	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(combinedCSV), 0o600)
		if err != nil {
			logger.Fatalf("Error writing to file: %v", err)
		}
		logger.Info("Results written to %s", outputFile)
	} else {
		logger.Info(combinedCSV)
	}
}
