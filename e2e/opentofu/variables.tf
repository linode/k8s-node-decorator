variable "linode_region" {
  type    = string
  default = "us-mia"
}

variable "lke_k8s_version" {
  type    = string
  default = "1.29"
}

variable "lke_cluster_label" {
  type = string
  default = "test-decorator"
}

variable "lke_kubeconfig_path" {
  type = string
  default = "./kubeconfig.yaml"
}
