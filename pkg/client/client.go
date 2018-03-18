package client

import (
	"k8s.io/client-go/kubernetes"
        "k8s.io/client-go/rest"
)


func GetClient() (kubernetes.Interface, err) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}
