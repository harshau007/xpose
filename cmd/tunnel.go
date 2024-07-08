/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"

	"github.com/harshau007/xpose/services"
	"github.com/spf13/cobra"
)

// tunnelCmd represents the tunnel command
var tunnelCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		_, _ = cmd.Flags().GetString("ssh")

		_, err := strconv.Atoi(port)
		if err != nil {
			fmt.Println(err)
		}

		// err = services.UsingBore(true, services.BoreClient{
		// 	RemoteServer: ssh,
		// 	RemotePort:   2200,
		// 	LocalServer:  "localhost",
		// 	LocalPort:    portNo,
		// 	BindPort:     0,
		// 	ID:           "",
		// 	KeepAlive:    true,
		// })
		err = services.Tunnel(port)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tunnelCmd)
	tunnelCmd.Flags().StringP("port", "p", "8080", "Port of running service to expose")
	tunnelCmd.Flags().StringP("ssh", "s", "bore.digital", "SSH server remote host")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tunnelCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tunnelCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
