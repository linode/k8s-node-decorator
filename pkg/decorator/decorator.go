package decorator

import (
	"context"
	"time"

	metadata "github.com/linode/go-metadata"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func StartDecorator(ctx context.Context, client metadata.Client, clientset *kubernetes.Clientset, interval time.Duration) {
	instanceData, err := client.GetInstance(ctx)
	if err != nil {
		klog.Fatalf("Failed to get the initial instance data: %s", err.Error())
	}

	err = UpdateNodeLabels(ctx, clientset, instanceData)
	if err != nil {
		klog.Error(err)
	}

	instanceWatcher := client.NewInstanceWatcher(
		metadata.WatcherWithInterval(interval),
	)

	go instanceWatcher.Start(ctx)

	for {
		select {
		case data := <-instanceWatcher.Updates:
			err = UpdateNodeLabels(ctx, clientset, data)
			if err != nil {
				klog.Fatal(err)
			}
		case err := <-instanceWatcher.Errors:
			klog.Errorf("Got error from instance watcher: %s", err)
		}
	}
}
