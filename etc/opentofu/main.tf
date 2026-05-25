terraform {
  required_version = ">= 1.6.0"

  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
  }
}

provider "local" {}

variable "example" {
  type = string
}

variable "evkey" {
  type = string
}

resource "local_file" "hello" {
  filename = "${path.module}/opentofu-output.txt"
  content  = "Hello from OpenTofu! example=${var.example}, evkey=${var.evkey}\n"
}
