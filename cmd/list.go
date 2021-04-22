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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"jfrog-cleaner/pkg"
	"k8s.io/client-go/kubernetes"
	"path"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list images supposed to cleanup",
	Long: `dry run for delete command. returns list images that will be deleted`,
	Run: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runList(cmd *cobra.Command, args []string) {
	log.SetFormatter(&log.JSONFormatter{})
	if Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	var k8sClient *kubernetes.Clientset
	var err error
	// k8s client creation
	if !jfrog.CheckInCluster() {
		k8sClient, err = jfrog.OutsideClusterClient()
	} else {
		k8sClient, err = jfrog.InClusterClient()
	}
	if err !=nil {
		log.Fatal(err)
	}
	banList, err := jfrog.GetBanList(k8sClient, Env)

	j := jfrog.Jfrog{Login: User, Password: Password}
	err = j.GetClient(Host)
	if err != nil {
		log.Fatal(err)
	}
	if !Oneshot {
		folders, err := j.GetFolders(Repo, Folder)
		if err != nil {
			log.Fatal(err)
		}
		log.Debug(folders.Results)
		for _, folder := range folders.Results {
			subFolders, err := j.GetFolders(Repo, path.Join(folder.Path, folder.Name))
			if err != nil {
				log.Fatal(err)
			}
			log.Debug(subFolders.Results)
			for _, subfolder := range subFolders.Results {
				images, err := j.GetImages(Repo, path.Join(subfolder.Path, subfolder.Name), Created, Num)
				log.Debug(images.Results)
				if err != nil {
					log.Warn(err)
				} else {
				for _, image := range images.Results {
					imageString, tag := j.ParseImage(image.Path, Registry)
					if !jfrog.CheckInBanlist(banList, imageString) && !jfrog.CheckInBanlist(FilterTags, tag) {
						log.WithFields(log.Fields{
							"modified":  image.Modified,
							"image":     imageString,
							"image_tag": tag,
						}).Info("image found")
					} else {
						log.WithFields(log.Fields{
							"modified":  image.Modified,
							"image":     imageString,
							"image_tag": tag,
						}).Info("image banned from deletion")
					}
				}
				}
			}
		}
	} else {
				images, err := j.GetImages(Repo, Folder, Created, Num)
				log.Debug(images.Results)
				if err != nil {
					log.Warn(err)
				} else {
					for _, image := range images.Results {
						imageString, tag := j.ParseImage(image.Path, Registry)
						if !jfrog.CheckInBanlist(banList, imageString) && !jfrog.CheckInBanlist(FilterTags, tag) {
							log.WithFields(log.Fields{
								"modified":  image.Modified,
								"image":     imageString,
								"image_tag": tag,
							}).Info("image found")
						} else {
							log.WithFields(log.Fields{
								"modified":  image.Modified,
								"image":     imageString,
								"image_tag": tag,
							}).Info("image banned from deletion")
						}
					}
				}
			}
}