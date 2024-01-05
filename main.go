package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	metadata "github.com/linode/go-metadata"
)

var version string

func init() {
	_ = flag.Set("logtostderr", "true")
}

func GetCurrentNode(clientset *kubernetes.Clientset) (*corev1.Node, error) {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return nil, errors.New("Environment variable NODE_NAME is not set")
	}

	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return node, nil
}

func UpdateNodeLabels(
	clientset *kubernetes.Clientset,
	instanceData *metadata.InstanceData,
) error {
	if instanceData == nil {
		return fmt.Errorf("instance data received from Linode metadata service is nil")
	}

	node, err := GetCurrentNode(clientset)
	if err != nil {
		return fmt.Errorf("failed to get the node: %w", err)
	}

	klog.Infof("Updating node labels with Linode instance data: %v", instanceData)

	node.Labels["linode_label"] = instanceData.Label
	node.Labels["linode_id"] = strconv.Itoa(instanceData.ID)
	node.Labels["linode_region"] = instanceData.Region
	node.Labels["linode_type"] = instanceData.Type
	node.Labels["linode_host"] = instanceData.HostUUID

	_, err = clientset.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})

	if err != nil {
		klog.Errorf("Failed to update labels: %s", err.Error())
		return err
	}

	return nil
}

func GetClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func StartDecorator(client metadata.Client, clientset *kubernetes.Clientset, interval time.Duration) {
	instanceData, err := client.GetInstance(context.TODO())
	if err != nil {
		klog.Errorf("Failed to get the initial instance data: %s", err.Error())
	} else {
		err = UpdateNodeLabels(clientset, instanceData)
	}

	if err != nil {
		klog.Error(err)
	}

	instanceWatcher := client.NewInstanceWatcher(
		metadata.WatcherWithInterval(interval),
	)

	go instanceWatcher.Start(context.Background())

	for {
		select {
		case data := <-instanceWatcher.Updates:
			err = UpdateNodeLabels(clientset, data)
			if err != nil {
				klog.Error(err)
			}
		case err := <-instanceWatcher.Errors:
			klog.Errorf("Got error from instance watcher: %s", err)
		}
	}
}

func main() {
	pollingIntervalSeconds := flag.Int(
		"poll-interval", 60,
		"The interval (in seconds) to poll and update node information",
	)
	flag.Parse()

	interval := time.Duration(*pollingIntervalSeconds) * time.Second

	klog.Infof("Starting Linode Kubernetes Node Decorator: version %s", version)
	klog.Infof("The poll interval is set to %v.", interval)

	clientset, err := GetClientset()
	if err != nil {
		panic(err.Error())
	}

	_, err = GetCurrentNode(clientset)
	if err != nil {
		panic(err.Error())
	}

	client, err := metadata.NewClient(
		context.Background(),
		metadata.ClientWithManagedToken(),
	)
	if err != nil {
		panic(err)
	}

	StartDecorator(*client, clientset, interval)
}
