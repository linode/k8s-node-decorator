package decorator

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"strings"

	metadata "github.com/linode/go-metadata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (d *Decorator) updateNodeLabels(ctx context.Context, instanceData *metadata.InstanceData) error {
	if instanceData == nil {
		return fmt.Errorf("instance data received from Linode metadata service is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	node, err := d.getCurrentNode(ctx)
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

	handleUpdated(setLabel(node, d.prefix+"/label", instanceData.Label))
	handleUpdated(setLabel(node, d.prefix+"/instance-id", strconv.Itoa(instanceData.ID)))
	handleUpdated(setLabel(node, d.prefix+"/region", instanceData.Region))
	handleUpdated(setLabel(node, d.prefix+"/instance-type", instanceData.Type))
	handleUpdated(setLabel(node, d.prefix+"/host", instanceData.HostUUID))

	oldTags := make(map[string]string)
	maps.Copy(oldTags, node.Labels)

	tagLabelPrefix := d.tagsPrefix + "." + d.prefix + "/"
	newTags := ParseTags(instanceData.Tags, tagLabelPrefix)

	for key := range oldTags {
		if !strings.HasPrefix(key, tagLabelPrefix) {
			continue
		}
		if _, ok := newTags[key]; !ok {
			delete(node.Labels, key)
			labelsUpdated = true
			continue
		}
	}

	for key, value := range newTags {
		handleUpdated(setLabel(node, key, value))
	}

	if !labelsUpdated {
		return nil
	}

	_, err = d.clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update labels: %s", err.Error())
		return err
	}

	klog.Infof("Successfully updated the labels of the node")

	return nil
}

func setLabel(node *corev1.Node, key, newValue string) (changed bool) {
	changed = false
	oldValue, ok := node.Labels[key]

	if !ok || oldValue != newValue {
		changed = true
		node.Labels[key] = newValue
	}

	return changed
}
