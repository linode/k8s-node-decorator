terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
    }
  }
}

provider "linode" {
}

resource "linode_lke_cluster" "e2e-test-cluster" {
    label       = var.lke_cluster_label
    k8s_version = var.lke_k8s_version
    region      = var.linode_region

    pool {
        type  = "g6-standard-1"
        count = 1
    }
}

resource "local_sensitive_file" "kubeconfig" {
  content_base64  = linode_lke_cluster.e2e-test-cluster.kubeconfig
  filename = var.lke_kubeconfig_path
  file_permission = "0600"
}
