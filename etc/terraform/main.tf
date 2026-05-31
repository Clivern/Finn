terraform {
  required_version = ">= 1.3.0"

  backend "http" {}

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
  filename = "${path.module}/terraform-output.txt"
  content  = "Hello from Terraform! example=${var.example}, evkey=${var.evkey}\n"
}
