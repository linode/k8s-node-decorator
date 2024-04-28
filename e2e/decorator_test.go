package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

func assertLabelsSet(t *testing.T, nodes corev1.NodeList) bool {
	t.Helper()

	for _, node := range nodes.Items {
		for _, label := range expectedLabels {
			if _, ok := node.Labels[label]; !ok {
				return ok
			}
		}
	}
	return true
}

func TestSetLabels(t *testing.T) {
	feature := features.New("Set Labels").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			manager := helm.New(config.KubeconfigFile())
			err := manager.RunInstall(
				helm.WithName("test-decorator"),
				helm.WithNamespace(namespace),
				helm.WithChart(filepath.Join(curDir, "..", "helm", "k8s-node-decorator")),
				helm.WithArgs("--set", fmt.Sprintf("decorator.image.tag=%s", tag)),
				helm.WithArgs("--set", fmt.Sprintf("decorator.image.repository=%s", repo)),
				helm.WithArgs("--set", fmt.Sprintf("decorator.prefix=%s", LabelPrefix)),
				helm.WithArgs("--set", fmt.Sprintf("rbac.name=%s", rbacName)),
				helm.WithWait(),
				helm.WithTimeout("10m"),
			)
			if err != nil {
				t.Fatal("failed to invoke helm install operation due to an error", err)
			}
			return ctx
		}).
		Assess("Check Label", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			nodes := corev1.NodeList{}

			err := config.Client().Resources().List(ctx, &nodes)
			if err != nil {
				t.Fatal(err)
			}

			// retry when labels are not set in a timely manner
			for i := 1; i <= 3; i++ {
				ok := assertLabelsSet(t, nodes)
				if ok {
					t.Log("all expected labels are set")
					break
				} else {
					t.Log("failed to assert decorator labels, retrying after 5s...")
					time.Sleep(5 * time.Second)
				}
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, feature)
}
