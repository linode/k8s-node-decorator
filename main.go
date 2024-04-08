// Copyright 2024 Akamai Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	metadata "github.com/linode/go-metadata"
	decorator "github.com/linode/k8s-node-decorator/k8snodedecorator"
)

var version string

func init() {
	_ = flag.Set("logtostderr", "true")
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

func StartDecorator(ctx context.Context, client metadata.Client, clientset *kubernetes.Clientset, interval time.Duration) {
	instanceData, err := client.GetInstance(ctx)
	if err != nil {
		klog.Fatalf("Failed to get the initial instance data: %s", err.Error())
	}

	err = decorator.UpdateNodeLabels(ctx, clientset, instanceData)
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
			err = decorator.UpdateNodeLabels(ctx, clientset, data)
			if err != nil {
				klog.Fatal(err)
			}
		case err := <-instanceWatcher.Errors:
			klog.Errorf("Got error from instance watcher: %s", err)
		}
	}
}

func main() {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		klog.Fatal("Environment variable NODE_NAME is not set")
	}
	decorator.SetNodeName(nodeName)

	var interval time.Duration
	flag.DurationVar(
		&interval, "poll-interval", 5*time.Minute,
		"The time interval to poll and update node information",
	)
	var timeout time.Duration
	flag.DurationVar(
		&timeout, "timeout", 30*time.Second,
		"The timeout for metadata and k8s client operations",
	)

	flag.Parse()

	klog.Infof("Starting Linode Kubernetes Node Decorator: version %s", version)
	klog.Infof("The poll interval is set to %v.", interval)
	klog.Infof("The timeout is set to %v.", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	clientset, err := GetClientset()
	if err != nil {
		klog.Fatal(err)
	}

	client, err := metadata.NewClient(
		ctx,
		metadata.ClientWithManagedToken(),
	)
	if err != nil {
		klog.Fatal(err)
	}

	StartDecorator(ctx, *client, clientset, interval)
}
