package decorator

import (
	"context"
	"time"

	metadata "github.com/linode/go-metadata"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type Decorator struct {
	client    *metadata.Client
	clientset *kubernetes.Clientset
	interval  time.Duration
	nodeName  string
}

func NewDecorator(options ...func(*Decorator)) *Decorator {
	d := &Decorator{}
	for _, o := range options {
		o(d)
	}

	return d
}

func WithClient(c *metadata.Client) func(*Decorator) {
	return func(d *Decorator) {
		d.client = c
	}
}

func WithClientSet(k *kubernetes.Clientset) func(*Decorator) {
	return func(d *Decorator) {
		d.clientset = k
	}
}

func WithInterval(i time.Duration) func(*Decorator) {
	return func(d *Decorator) {
		d.interval = i
	}
}

func WithNodeName(n string) func(*Decorator) {
	return func(d *Decorator) {
		d.nodeName = n
	}
}

func (d *Decorator) Start(ctx context.Context) {
	instanceData, err := d.client.GetInstance(ctx)
	if err != nil {
		klog.Fatalf("Failed to get the initial instance data: %s", err.Error())
	}

	err = d.updateNodeLabels(ctx, instanceData)
	if err != nil {
		klog.Error(err)
	}

	instanceWatcher := d.client.NewInstanceWatcher(
		metadata.WatcherWithInterval(d.interval),
	)

	go instanceWatcher.Start(ctx)

	for {
		select {
		case data := <-instanceWatcher.Updates:
			err = d.updateNodeLabels(ctx, data)
			if err != nil {
				klog.Fatal(err)
			}
		case err := <-instanceWatcher.Errors:
			klog.Errorf("Got error from instance watcher: %s", err)
		}
	}
}
