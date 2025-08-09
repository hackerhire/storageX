package main

import (
	"fmt"
	"os"

	"github.com/sayuyere/storageX/internal/app"
	"github.com/sayuyere/storageX/internal/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func main() {
	var services *app.ServiceBundle

	log.InitLogger(true)

	rootCmd := &cobra.Command{
		Use:   "storagex",
		Short: "storageX CLI for modular cloud file chunking and storage",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cfgFile != "" {
				viper.SetConfigFile(cfgFile)
			} else {
				viper.SetConfigName("config")
				viper.AddConfigPath("./config")
			}
			viper.SetConfigType("json")
			if err := viper.ReadInConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
				os.Exit(1)
			}
			// Initialize all services for CLI use
			var err error
			services, err = app.NewServiceBundle(viper.ConfigFileUsed())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Service init error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (required)")
	rootCmd.MarkPersistentFlagRequired("config")

	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the version number",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("storageX v0.1.0")
			},
		},
	)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "upload [file]",
		Short: "Upload a file to cloud storage",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filePath := args[0]
			fmt.Println("Uploading:", filePath)
			if services == nil {
				fmt.Fprintln(os.Stderr, "Services not initialized")
				os.Exit(1)
			}
			if err := services.Storage.UploadFile(filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Upload failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Upload successful!")
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "download [file] [output]",
		Short: "Download a file from cloud storage",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fileName := args[0]
			output := args[1]
			fmt.Println("Downloading:", fileName, "to", output)
			if services == nil {
				fmt.Fprintln(os.Stderr, "Services not initialized")
				os.Exit(1)
			}
			file, err := os.Create(output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
				os.Exit(1)
			}
			// TODO: implement services.Storage.GetFile(fileName, output)
			err = services.Storage.GetFile(fileName, file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
				os.Exit(1)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "configfile",
		Short: "Print the path to the config file in use",
		Run: func(cmd *cobra.Command, args []string) {
			if cfgFile != "" {
				fmt.Println(cfgFile)
			} else {
				fmt.Println("./config/config.json")
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
