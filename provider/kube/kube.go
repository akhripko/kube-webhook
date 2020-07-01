package kube

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Provider struct {
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) GetPod(namespace string, name string) (*core.Pod, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	pods, err := c.CoreV1().Pods(namespace).List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, item := range pods.Items {
		if item.Name == name {
			return &item, nil
		}
	}
	return nil, ErrNotFound(namespace + "/" + name)
}
