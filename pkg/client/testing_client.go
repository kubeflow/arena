package client

import "k8s.io/client-go/kubernetes"

func NewClientForTesting(clientset kubernetes.Interface) *Client {
	return &Client{
		clientset: clientset,
	}
}
