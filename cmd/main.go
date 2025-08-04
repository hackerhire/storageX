package main

import (
	"fmt"
	"os"

	"github.com/sayuyere/storageX/internal/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func main() {
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
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config/config.json)")

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
			// TODO: wire up storage service and call UploadFile
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
			// TODO: wire up storage service and call GetFile
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
