package cmd

import (
	"fmt"
	"os"

	"caching-proxy/cache"
	"caching-proxy/proxy"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	port       int
	origin     string
	clearCache bool
)

var rootCmd = &cobra.Command{
	Use:   "caching-proxy",
	Short: "A caching proxy server CLI",
	Run: func(cmd *cobra.Command, args []string) {
		if clearCache {
			cache.ClearCache()
			fmt.Println("Cache cleared!")
			return
		}
		if origin == "" {
			fmt.Println("Error: origin must be specified when not clearing cache")
			os.Exit(1)
		}
		proxy.StartServer(origin, port)
	},
}

func init() {
	rootCmd.Flags().IntVarP(&port, "port", "p", 3000, "Port on which the proxy will listen")
	rootCmd.Flags().StringVarP(&origin, "origin", "o", "", "Origin server URL to forward requests to")
	rootCmd.Flags().BoolVar(&clearCache, "clear-cache", false, "Clear the cache")

	if err := viper.BindPFlag("port", rootCmd.Flags().Lookup("port")); err != nil {
		fmt.Println("Error binding port flag:", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("origin", rootCmd.Flags().Lookup("origin")); err != nil {
		fmt.Println("Error binding origin flag:", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("clearCache", rootCmd.Flags().Lookup("clear-cache")); err != nil {
		fmt.Println("Error binding clearCache flag:", err)
		os.Exit(1)
	}
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
