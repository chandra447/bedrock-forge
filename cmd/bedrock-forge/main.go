package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"bedrock-forge/internal/commands"
	"bedrock-forge/pkg/config"
)

var logger *logrus.Logger

var rootCmd = &cobra.Command{
	Use:   "bedrock-forge",
	Short: "Transform YAML configurations into AWS Bedrock agent deployments",
	Long:  `Bedrock Forge is a CLI tool that transforms YAML configurations into AWS Bedrock agent deployments using Terraform modules.`,
}

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Discover and list all resources in the current directory",
	Long:  `Scan the current directory for YAML files and discover all Bedrock resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		var scanPath string
		if len(args) > 0 {
			scanPath = args[0]
		}

		scanCommand := commands.NewScanCommand(logger)
		if err := scanCommand.Execute(scanPath); err != nil {
			logger.WithError(err).Fatal("Failed to execute scan command")
		}
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate YAML syntax and dependencies",
	Long:  `Validate all discovered YAML files for syntax errors and dependency issues.`,
	Run: func(cmd *cobra.Command, args []string) {
		var validatePath string
		if len(args) > 0 {
			validatePath = args[0]
		}

		validateCommand := commands.NewValidateCommand(logger)
		if err := validateCommand.Execute(validatePath); err != nil {
			logger.WithError(err).Fatal("Failed to execute validate command")
		}
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate [path] [output-dir]",
	Short: "Generate Terraform configuration from YAML resources",
	Long:  `Generate Terraform configuration files from discovered YAML resources.

Arguments:
  path        Path to directory containing YAML files (default: current directory)
  output-dir  Output directory for generated Terraform files (default: outputs_tf)

The generated Terraform files will be placed in the outputs_tf directory by default,
so you can immediately inspect the generated .tf files without any additional setup.`,
	Run: func(cmd *cobra.Command, args []string) {
		var scanPath, outputDir string
		if len(args) > 0 {
			scanPath = args[0]
		}
		if len(args) > 1 {
			outputDir = args[1]
		}

		generateCommand := commands.NewGenerateCommand(logger)
		if err := generateCommand.Execute(scanPath, outputDir); err != nil {
			logger.WithError(err).Fatal("Failed to execute generate command")
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version and build info",
	Long:  `Display the version number and build information for bedrock-forge.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("bedrock-forge version 0.1.0")
	},
}

func init() {
	logger = config.SetupSimpleLogger()

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Fatal("Command execution failed")
		os.Exit(1)
	}
}
