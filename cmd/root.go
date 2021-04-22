/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var (
	Host     string
	User     string
	Password string
	Repo     string
	Folder   string
	Created  string
	Num      int
	Workers  int
	Oneshot bool
	Registry string
	Debug bool
	Env string
	FilterTags []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jfrog-cleaner",
	Short: "cli tool for docker registry cleanup",
	Long: `Search and delete old docker images.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jfrog-cleaner.yaml)")
	rootCmd.PersistentFlags().StringVarP(&User, "user", "u", "", "Username for artifactory login")
	rootCmd.PersistentFlags().StringVarP(&Password, "password", "p", "", "Password for artifactory login")
	rootCmd.PersistentFlags().StringVarP(&Repo, "repo", "r", "", "Repository to clean")
	rootCmd.PersistentFlags().StringVarP(&Folder, "folder", "f", "", "Directory to clean")
	rootCmd.PersistentFlags().StringVarP(&Created, "created", "c", "", "Determines how old artifact need to be deleted. 1mo - 1 month old."+
		" 2w - two weeks old")
	rootCmd.PersistentFlags().IntVarP(&Num, "num", "n", 5, "Determines how many artifacts to keep after cleanup")
	rootCmd.PersistentFlags().StringVarP(&Host, "Host", "H", "", "artifactory url")
	rootCmd.PersistentFlags().IntVarP(&Workers, "workers", "w", 5, "Determines how many concurent workers app will start")
	rootCmd.PersistentFlags().BoolVarP(&Oneshot, "oneshot", "O", false, "Search images recursively" )
	rootCmd.PersistentFlags().StringVarP(&Registry, "docker-registry", "d", "", "docker registry")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "b", false, "enables debug logging")
	rootCmd.PersistentFlags().StringVarP(&Env, "env", "e", "stage", "namespace filtering env label")
	rootCmd.PersistentFlags().StringSliceVarP(&FilterTags, "filter-tags", "j", []string{}, "slice of docker tags to filter")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".jfrog-cleaner" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".jfrog-cleaner")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
