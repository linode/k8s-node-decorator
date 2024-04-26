/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	// If testing with a cloud vendor managed cluster uncomment one of the below dependencies to properly get authorised.
	//_ "k8s.io/client-go/plugin/pkg/client/auth/azure" // auth for AKS clusters
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"   // auth for GKE clusters
	//_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"  // auth for OIDC
	"context"
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"

	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	TagEnvVar   = "DECORATOR_TEST_TAG"
	RepoEnvVar  = "DECORATOR_TEST_REPO"
	NodeNameVar = "NODE_NAME"
	LabelPrefix = "decorator.linode.com"
)

var (
	testenv        env.Environment
	curDir         string
	namespace      string
	tag            string
	repo           string
	expectedLabels []string
	rbacName       string
)

func configureRBACName() {
	rbacName = envconf.RandomName("test-decorator-rbac", 32)
}

func configureExpectedNodeLabels() {
	expectedLabels = []string{
		LabelPrefix + "/label",
		LabelPrefix + "/instance-id",
		LabelPrefix + "/region",
		LabelPrefix + "/instance-type",
		LabelPrefix + "/host",
	}
}

func configureImage() {
	tag = os.Getenv(TagEnvVar)
	if tag == "" {
		panic(fmt.Sprintf(
			"you have to configure the environment variable %q "+
				"for the test container image.",
			TagEnvVar,
		))
	}

	repo = os.Getenv(RepoEnvVar)
	if repo == "" {
		repo = "docker.io/linode/k8s-node-decorator"
	}
}

func configureCurrentDirectory() {
	c, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	curDir = c
}

func cleanupDecoratorClusterRoles(ctx context.Context, config *envconf.Config) (context.Context, error) {
	clusterRole := rbacv1.ClusterRole{}
	clusterRoleBinding := rbacv1.ClusterRoleBinding{}

	config.Client().Resources().Get(ctx, rbacName, namespace, &clusterRole)
	config.Client().Resources().Get(ctx, rbacName, namespace, &clusterRoleBinding)

	config.Client().Resources().Delete(ctx, &clusterRole)
	config.Client().Resources().Delete(ctx, &clusterRoleBinding)

	return ctx, nil
}

func TestMain(m *testing.M) {
	configureRBACName()
	configureCurrentDirectory()
	configureImage()
	configureExpectedNodeLabels()

	testenv = env.New()
	namespace = envconf.RandomName("test-decorator", 32)
	path := conf.ResolveKubeConfigFile()
	if path == "" {
		panic("a kubeconfig file is required for e2e testing.")
	}
	cfg := envconf.NewWithKubeConfig(path)
	testenv = env.NewWithConfig(cfg)

	testenv.Setup(
		envfuncs.CreateNamespace(namespace),
	)
	testenv.Finish(
		envfuncs.DeleteNamespace(namespace),
		cleanupDecoratorClusterRoles,
	)

	os.Exit(testenv.Run(m))
}
