package k8snodedecorator

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var nodeName string

func SetNodeName(newNodeName string) {
	nodeName = newNodeName
}

func GetCurrentNode(clientset *kubernetes.Clientset) (*corev1.Node, error) {
	return clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
}
