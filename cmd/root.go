/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/cwxstat/go-pod-launch-run/pkg"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gopl",
	Short: "Pod launch and run command",
	Long: `Command launches a aws-cli pod and runs a command in it.

`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		err := pkg.Run(podName, namespace, container, serviceaccount, args, outputFile)
		if err != nil {
			panic(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var podName string
var namespace string
var container string
var serviceaccount string

var outputFile string

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mylaunch.yaml)")

	rootCmd.PersistentFlags().StringVar(&podName, "podName", "aws-cli-pod", "Pod name")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "Namespace")
	rootCmd.PersistentFlags().StringVar(&container, "container", "aws-cli", "Container name")
	rootCmd.PersistentFlags().StringVar(&serviceaccount, "serviceaccount", "default", "Service account name")
	rootCmd.PersistentFlags().StringVar(&outputFile, "output", "result.pod", "Output file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
