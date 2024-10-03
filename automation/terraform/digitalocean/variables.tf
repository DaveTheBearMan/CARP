# Manager variables
variable "manager_droplet_name" {}
variable "manager_droplet_region" {}
variable "manager_droplet_size" {}

# Node variables
variable "node_droplet_name" {}
variable "node_droplet_region" {}
variable "node_droplet_size" {}

# Digital Ocean Variables
variable "do_token" {}
variable "pvt_key" {}
variable "pub_key" {}
variable "project_id" {}

variable "node_droplet_image" {
  description = "The Digital Ocean Snapshot ID that was returned from Packer"
}

variable "manager_droplet_image" {
  description = "The Digital Ocean Snapshot ID that was returned from Packer"
}