terraform {
  required_version = ">= 1.5.0"
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_namespace" "rate_limiter" {
  metadata {
    name = "rate-limiter"
  }
}
