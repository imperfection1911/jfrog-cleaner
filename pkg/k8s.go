package jfrog

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func CheckInCluster() bool {
	info, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func OutsideClusterClient() (client *kubernetes.Clientset, err error) {
	var kubeconfig string
	home := homedir.HomeDir()
	kubeconfig = filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return
	}
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return
	}
	return
}

func InClusterClient() (client *kubernetes.Clientset, err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return
	}
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return
	}
	return
}

func GetNamespaces(client *kubernetes.Clientset, env string) (namespaces []string, err error) {
	label := map[string]string{"env": env}
	filter := metav1.ListOptions{LabelSelector: labels.SelectorFromSet(label).String()}
	ns, err := client.CoreV1().Namespaces().List(filter)
	if err != nil {
		return
	}
	for _, v := range ns.Items {
		namespaces = append(namespaces, v.Name)
	}
	return
}

func GetPodImages(client *kubernetes.Clientset, namespace string) (images []string, err error) {
	pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, v := range pods.Items {
		var podImages []string
		for _, c := range v.Spec.Containers {
			podImages = append(podImages, c.Image)
		}
		for _, ic := range v.Spec.InitContainers {
			podImages = append(podImages, ic.Image)
		}
		for _, i := range podImages {
			images = append(images, i)
		}
	}
	return
}

func GetBanList(client *kubernetes.Clientset, env string) (banList []string, err error) {
	namespaces, err := GetNamespaces(client, env)
	for _, n := range namespaces {
		images, err := GetPodImages(client, n)
		if err != nil {
			return banList, err
		}
		for _, i := range images {
			banList = append(banList, i)
		}
	}
	return
}

func CheckInBanlist(banList []string, image string) bool {
	for _, b := range banList {
		if b == image {
			return true
		}
	}
	return false
}
