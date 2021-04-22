package jfrog

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type SearchResponse struct {
	Results []struct {
		Repo       string    `json:"repo"`
		Path       string    `json:"path"`
		Name       string    `json:"name"`
		Type       string    `json:"type"`
		Size       int       `json:"size"`
		Created    time.Time `json:"created"`
		CreatedBy  string    `json:"created_by"`
		Modified   time.Time `json:"modified"`
		ModifiedBy string    `json:"modified_by"`
		Updated    time.Time `json:"updated"`
	} `json:"results"`
	Range struct {
		StartPos int `json:"start_pos"`
		EndPos   int `json:"end_pos"`
		Total    int `json:"total"`
	} `json:"range"`
}

type Jfrog struct {
	BaseUrl  string
	AqlUrl   *url.URL
	Login    string
	Password string
	Client   *http.Client
}

func (j *Jfrog) GetClient(artifactoryUrl string) (err error) {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	j.Client = &http.Client{Transport: transport}
	j.BaseUrl = artifactoryUrl
	j.AqlUrl, err = url.Parse(artifactoryUrl)
	j.AqlUrl.Path = path.Join(j.AqlUrl.Path, "api/search/aql")
	return
}

func (j *Jfrog) GetFolders(repo, rootFolder string) (searchResponse SearchResponse, err error) {
	body := fmt.Sprintf("items.find({\"repo\":{\"$eq\": \"%s\"}, \"type\": "+
		"{\"$eq\": \"folder\"}, \"path\":{\"$eq\": \"%s\"}})", repo, rootFolder)
	request, err := http.NewRequest(http.MethodPost, j.AqlUrl.String(), bytes.NewBuffer([]byte(body)))
	if err != nil {
		return
	}
	request.SetBasicAuth(j.Login, j.Password)
	response, err := j.Client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(responseData, &searchResponse)
	if err != nil {
		return
	}
	return
}

func (j *Jfrog) GetImages(repo, path, created string, num int) (searchResponse SearchResponse, err error) {
	var body string
	if created != "" {
		body = fmt.Sprintf("items.find({\"repo\":{\"$eq\": \"%s\"}, \"path\":{\"$match\": "+
			"\"%s*\"}, \"modified\": {\"$before\": \"%s\"}, "+
			"\"name\":{\"$eq\":\"manifest.json\"}}).sort({\"$desc\": [\"modified\"]})", repo, path, created)
	} else {
		body = fmt.Sprintf("items.find({\"repo\":{\"$eq\": \"%s\"}, \"path\":{\"$match\": "+
			"\"%s*\"}, \"name\":{\"$eq\":\"manifest.json\"}}).sort({\"$desc\": [\"modified\"]})", repo, path)
	}
	request, err := http.NewRequest(http.MethodPost, j.AqlUrl.String(), bytes.NewBuffer([]byte(body)))
	if err != nil {
		return
	}
	request.SetBasicAuth(j.Login, j.Password)
	response, err := j.Client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(responseData, &searchResponse)
	if err != nil {
		return
	}
	if len(searchResponse.Results) >= num {
		searchResponse.Results = searchResponse.Results[num-1:]
	} else {
		err = errors.New(fmt.Sprintf("too few images in repo %s", path))
	}
	return
}

func (j *Jfrog) DeleteImage(repo, imagePath string) (err error) {
	deleteUrl, err := url.Parse(j.BaseUrl)
	if err != nil {
		return
	}
	deleteUrl.Path = path.Join(deleteUrl.Path, repo, imagePath)
	request, err := http.NewRequest(http.MethodDelete, deleteUrl.String(), nil)
	if err != nil {
		return
	}
	request.SetBasicAuth(j.Login, j.Password)
	response, err := j.Client.Do(request)
	if err != nil {
		return
	}
	if response.StatusCode != 204 {
		err = errors.New(fmt.Sprintf("troubles while deleting artifact %s. Status code: %s", imagePath, strconv.Itoa(response.StatusCode)))
	}
	return
}

func (j *Jfrog) ParseImage(imagePath, registry string) (image, tag string){
	splited := strings.Split(imagePath, "/")
	image = registry
	tag = splited[len(splited) - 1]
	splited = splited[:len(splited) - 1]
	for _, part := range splited {
		image = path.Join(image, part)
	}
	image = fmt.Sprintf("%s:%s", image, tag)
	return
}
