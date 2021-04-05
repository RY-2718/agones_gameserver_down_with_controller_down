package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getKubeConfig() (*rest.Config, error) {
	useInClusterConfig := os.Getenv("USE_INCLUSTERCONFIG")
	if useInClusterConfig == "" {
		kubeconfig := filepath.Join(homeDir(), ".kube", "config")
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {
	namespace := "default"

	fmt.Println("getting k8s config...")
	config, err := getKubeConfig()
	if err != nil {
		log.Fatal(err)
	}
	agonescs, err := versioned.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("getting gameservers...")
	ctx := context.Background()
	gss, err := agonescs.AgonesV1().GameServers(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("deleting a gameserver...")
	gs := gss.Items[0]
	gs.Status.State = agonesv1.GameServerStateShutdown
	if _, err := agonescs.AgonesV1().GameServers(namespace).Update(ctx, &gs, metav1.UpdateOptions{}); err != nil {
		log.Fatal(err)
	}
}
