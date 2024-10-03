# Terraform provider
terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# Get digital ocean provider
provider "digitalocean" {
  token = var.do_token
}

# Create ssh key for terraform from api key
data "digitalocean_ssh_key" "terraform" {
  name = "terraform"
}