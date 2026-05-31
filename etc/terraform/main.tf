terraform {
  required_version = ">= 1.3.0"

  backend "http" {}

  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.13"
    }
  }
}

provider "local" {}
provider "time" {}

variable "example" {
  type = string
}

variable "evkey" {
  type = string
}

resource "local_file" "hello" {
  depends_on = [time_sleep.backend_lock_window]

  filename = pathexpand("~/Desktop/terraform-output.txt")
  content  = "Hello from Terraform! example=${var.example}, evkey=${var.evkey}\n"
}

resource "time_sleep" "backend_lock_window" {
  create_duration = "10s"

  triggers = {
    always = timestamp()
  }
}
