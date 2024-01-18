package main

import (
	"context"
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
	decorator "github.com/linode/k8s-node-decorator/k8snodedecorator"
)

var (
	version  string
	nodeName string
)

func init() {
	_ = flag.Set("logtostderr", "true")
}

func GetCurrentNode(clientset *kubernetes.Clientset) (*corev1.Node, error) {
	return clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
}

func SetLabel(node *corev1.Node, key, newValue string) (changed bool) {
	changed = false
	oldValue, ok := node.Labels[key]

	if !ok || oldValue != newValue {
		changed = true
		node.Labels[key] = newValue
	}

	return changed
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
	labelsUpdated := false

	handleUpdated := func(updated bool) {
		if updated {
			labelsUpdated = updated
		}
	}

	handleUpdated(SetLabel(node, "decorator.linode.com/label", instanceData.Label))
	handleUpdated(SetLabel(node, "decorator.linode.com/instance-id", strconv.Itoa(instanceData.ID)))
	handleUpdated(SetLabel(node, "decorator.linode.com/region", instanceData.Region))
	handleUpdated(SetLabel(node, "decorator.linode.com/instance-type", instanceData.Type))
	handleUpdated(SetLabel(node, "decorator.linode.com/host", instanceData.HostUUID))

	for key, value := range decorator.ParseTags(instanceData.Tags) {
		handleUpdated(SetLabel(node, key, value))
	}

	if !labelsUpdated {
		return nil
	}

	_, err = clientset.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})

	if err != nil {
		klog.Errorf("Failed to update labels: %s", err.Error())
		return err
	}

	klog.Infof("Successfully updated the labels of the node")

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
		klog.Fatalf("Failed to get the initial instance data: %s", err.Error())
	}

	err = UpdateNodeLabels(clientset, instanceData)
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
				klog.Fatal(err)
			}
		case err := <-instanceWatcher.Errors:
			klog.Errorf("Got error from instance watcher: %s", err)
		}
	}
}

func main() {
	nodeName = os.Getenv("NODE_NAME")
	if nodeName == "" {
		klog.Fatal("Environment variable NODE_NAME is not set")
	}

	var interval time.Duration
	flag.DurationVar(
		&interval, "poll-interval", 5*time.Minute,
		"The time interval to poll and update node information",
	)
	flag.Parse()

	klog.Infof("Starting Linode Kubernetes Node Decorator: version %s", version)
	klog.Infof("The poll interval is set to %v.", interval)

	clientset, err := GetClientset()
	if err != nil {
		klog.Fatal(err)
	}

	_, err = GetCurrentNode(clientset)
	if err != nil {
		klog.Fatal(err)
	}

	client, err := metadata.NewClient(
		context.Background(),
		metadata.ClientWithManagedToken(),
	)
	if err != nil {
		klog.Fatal(err)
	}

	StartDecorator(*client, clientset, interval)
}
