package decorator

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *Decorator) getCurrentNode(ctx context.Context) (*corev1.Node, error) {
	return d.clientset.CoreV1().Nodes().Get(ctx, d.nodeName, metav1.GetOptions{})
}
