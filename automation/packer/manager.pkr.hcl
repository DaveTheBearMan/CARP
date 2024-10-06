packer {
  required_plugins {
    ansible = {
      source  = "github.com/hashicorp/ansible"
      version = "~> 1"
    }
  }
}

variable "digitalocean_api_token" {
  type    = string
  default = "${env("DIGITAL_OCEAN_API_TOKEN")}"
}

variable "droplet_image" {
  type    = string
  default = "${env("DROPLET_IMAGE")}"
}

variable "droplet_name" {
  type    = string
  default = "${env("MANAGER_DROPLET_NAME")}"
}

variable "droplet_region" {
  type    = string
  default = "${env("DROPLET_REGION")}"
}

variable "droplet_size" {
  type    = string
  default = "${env("MANAGER_DROPLET_SIZE")}"
}

source "digitalocean" "manager" {
  api_token    = "${var.digitalocean_api_token}"
  droplet_name = "${var.droplet_name}"
  image        = "${var.droplet_image}"
  region       = "${var.droplet_region}"
  size         = "${var.droplet_size}"
  ssh_username = "root"
}

build {
  sources = ["source.digitalocean.manager"]

  provisioner "shell" {
    script = "scripts/ansible.sh"
  }

  provisioner "file" {
    destination = "/"
    source      = "keys/private_key.asc"
  }

  provisioner "ansible-local" {
    group_vars    = "../ansible/group_vars"
    playbook_file = "../ansible/manager.yml"
    role_paths    = ["../ansible/roles/golang", "../ansible/roles/manager"]
  }

  provisioner "shell" {
    script = "scripts/cleanup.sh"
  }
}
