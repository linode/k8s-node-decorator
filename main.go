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
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	metadata "github.com/linode/go-metadata"
	"github.com/linode/k8s-node-decorator/pkg/decorator"
)

var version string

func init() {
	_ = flag.Set("logtostderr", "true")
}

func getClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func main() {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		klog.Fatal("Environment variable NODE_NAME is not set")
	}

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

	var prefix string
	flag.StringVar(
		&prefix, "prefix", "decorator.linode.com",
		"Node label prefix",
	)

	var tagsPrefix string
	flag.StringVar(
		&tagsPrefix, "tags-prefix", "tags",
		"Node label tags prefix",
	)

	flag.Parse()

	if !decorator.IsValidObjectName(prefix) {
		klog.Fatal(fmt.Errorf("invalid prefix"))
	}

	if !decorator.IsValidObjectName(tagsPrefix) {
		klog.Fatal(fmt.Errorf("invalid tags prefix"))
	}

	klog.Infof("Starting Linode Kubernetes Node Decorator: version %s", version)
	klog.Infof("The poll interval is set to %v.", interval)
	klog.Infof("The timeout is set to %v.", timeout)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	clientset, err := getClientset()
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

	decorator.NewDecorator(
		decorator.WithClient(client),
		decorator.WithClientSet(clientset),
		decorator.WithInterval(interval),
		decorator.WithTimeout(timeout),
		decorator.WithNodeName(nodeName),
		decorator.WithPrefix(prefix),
		decorator.WithTagsPrefix(tagsPrefix),
	).Start(ctx)
}
