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
	"sync"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete images",
	Long: `delete images by date and/or remain count`,
	Run: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runDelete(cmd *cobra.Command, args []string) {
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
	var allImages []string
	if !Oneshot {
		folders, err := j.GetFolders(Repo, Folder)
		if err != nil {
			log.Fatal(err)
		}
		for _, folder := range folders.Results {
			subFolders, err := j.GetFolders(Repo, path.Join(folder.Path, folder.Name))
			if err != nil {
				log.Fatal(err)
			}
			for _, subfolder := range subFolders.Results {
				images, err := j.GetImages(Repo, path.Join(subfolder.Path, subfolder.Name), Created, Num)
				if err != nil {
					log.Debug(err)
				} else {
					for _, image := range images.Results {
						allImages = append(allImages, image.Path)
					}
				}
				}
			}
		} else {
		images, err := j.GetImages(Repo, Folder, Created, Num)
		if err != nil {
			log.Warn(err)
		} else {
			for _, image := range images.Results {
				allImages = append(allImages, image.Path)
			}
		}
	}
	if len(allImages) > 0 {
		ch := make(chan string)
		wg := sync.WaitGroup{}
		// starting workers
		for worker := 0; worker < Workers; worker++ {
			wg.Add(1)
			go func(ch chan string) {
				for imagePath := range ch {
					imageString, tag := j.ParseImage(imagePath, Registry)
					if !jfrog.CheckInBanlist(banList, imageString) && !jfrog.CheckInBanlist(FilterTags, tag) {
						log.WithFields(log.Fields{
							"image":     imageString,
							"image_tag": tag,
							"status":    "removing",
						}).Info("removing image")
						err = j.DeleteImage(Repo, imagePath)
						if err != nil {
							log.WithFields(log.Fields{
								"image":     imageString,
								"error":     err,
								"image_tag": tag,
								"status":    "failed",
							}).Warn("failed to remove image")
						} else {
							log.WithFields(log.Fields{
								"image":     imageString,
								"image_tag": tag,
								"status":    "success",
							}).Info("image removed")
						}
					} else {
						log.WithFields(log.Fields{
							"image":     imageString,
							"image_tag": tag,
						}).Info("image banned from deletion")
					}
				}
				wg.Done()
			}(ch)
		}
		for _, image := range allImages {
			ch <- image
		}
		close(ch)
		wg.Wait()
	}
}
