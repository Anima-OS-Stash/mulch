package topics

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/Xfennec/mulch/cmd/mulch/client"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var globalHome string
var globalCfgFile string

var globalAPI *client.API
var globalConfig *RootConfig

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mulch",
	Short: "Mulch CLI client",
	Long: `Mulch is a light and practical virtual machine manager, using
libvirt API. This is the client.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n\n", cmd.Short)
		fmt.Printf("%s\n\n", cmd.Long)
		fmt.Printf("Use --help to list commands and options.\n\n")
		if globalConfig.ConfigFile != "" {
			fmt.Printf("configuration file: '%s'\n", globalConfig.ConfigFile)
		} else {
			fmt.Printf(`No configuration file found (%s).

This config file can provide 'key' setting (with your API
key) and mulchd API URL with the 'url' setting.

Example:
url = "http://192.168.10.104:8585"
key = "gein2xah7keel5Ohpe9ahvaeg8suurae3Chue4riokooJ5Wu"

Others available settings: trace, time
Note: you can also use environment variables (URL, KEY, …).
------
`, path.Clean(globalHome+"/.mulch.toml"))
		}
		fmt.Printf("current URL to mulchd: %s\n", globalConfig.URL)
		if globalConfig.Key == "" {
			fmt.Printf("\nWARNING: no API key defined! Add 'key' line to\nconfig file ordefine KEY environment variable.\n")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	var err error
	globalHome, err = homedir.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&globalCfgFile, "config", "c", "", "config file (default is $HOME/.mulch.yaml)")

	rootCmd.PersistentFlags().StringP("url", "u", "http://localhost:8585", "mulchd URL")
	rootCmd.PersistentFlags().BoolP("trace", "t", false, "also show server TRACE messages (debug)")
	rootCmd.PersistentFlags().BoolP("time", "d", false, "show server timestamps on messages")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	cfgFile := globalCfgFile
	if cfgFile == "" {
		cfgFile = path.Clean(globalHome + "/.mulch.toml")
	}

	var err error
	globalConfig, err = NewRootConfig(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	globalAPI = client.NewAPI(
		globalConfig.URL,
		globalConfig.Key,
		globalConfig.Trace,
		globalConfig.Time,
	)
}
