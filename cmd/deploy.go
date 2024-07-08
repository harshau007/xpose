/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/harshau007/xpose/services"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectName, _ := cmd.Flags().GetString("name")
		dockerImage, _ := cmd.Flags().GetString("docker")
		githubRepo, _ := cmd.Flags().GetString("github")
		port, _ := cmd.Flags().GetString("port")

		currPath, _ := os.Getwd()
		pathName := filepath.Join(currPath)
		var err error
		if dockerImage != "" {
			err = services.DeployDocker(dockerImage, port, projectName, pathName)
		}

		if githubRepo != "" {
			err = services.DeployGitHub(githubRepo, port, pathName)
		}
		if err != nil {
			fmt.Println("Error: ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("docker", "d", "", "Docker image to deploy")
	deployCmd.Flags().StringP("github", "g", "", "GitHub repository URL to deploy")
	deployCmd.Flags().StringP("port", "p", "8080", "Port to expose for the deployment")
	deployCmd.Flags().StringP("name", "n", "", "Project name (required)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
