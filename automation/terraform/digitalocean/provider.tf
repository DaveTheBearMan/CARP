# Terraform provider
terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# Image configuration variables
variable "droplet_name" {}
variable "droplet_region" {}
variable "droplet_size" {}

# Token variables
variable "do_token" {}
variable "pvt_key" {}
variable "project_id" {}


# Get digital ocean provider
provider "digitalocean" {
  token = var.do_token
}

# Create ssh key for terraform from api key
data "digitalocean_ssh_key" "terraform" {
  name = "terraform"
}
