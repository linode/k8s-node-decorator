package main

import (
	"flag"

	"k8s.io/klog/v2"
)

var version string

func init() {
	_ = flag.Set("logtostderr", "true")
}

func main() {
	klog.Infof("Starting Linode Kubernetes Node Decorator: version %s", version)
}
