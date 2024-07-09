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

var (
	Path, _  = os.Getwd()
	PathName = filepath.Join(Path)
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		projectName, _ := cmd.Flags().GetString("name")
		dockerImage, _ := cmd.Flags().GetString("docker")
		githubRepo, _ := cmd.Flags().GetString("github")
		port, _ := cmd.Flags().GetString("port")

		var err error
		if dockerImage != "" {
			err = services.DeployDocker(dockerImage, port, projectName, PathName)
		}

		if githubRepo != "" {
			err = services.DeployGitHub(githubRepo, port, PathName)
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
