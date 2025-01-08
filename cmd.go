package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goneypot",
	Short: "SSH honeypot",
	Long:  `Low interaction SSH honeypot.`,
	Run: func(cmd *cobra.Command, args []string) {
		startHoneypot()
	},
}

func ExecuteCLI() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&Addr, "addr", "a", "0.0.0.0", "honeypot listen address")
	rootCmd.Flags().StringVarP(&Port, "port", "p", "2222", "honeypot listen port")
	rootCmd.Flags().StringVarP(&PrivateKeyFile, "key", "k", "id_rsa", "private SSH key file")
	rootCmd.Flags().StringVar(&Prompt, "prompt", "[user@db01:~]$ ", "shell prompt")
}
