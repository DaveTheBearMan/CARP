locals {
    nodes_to_create = {
        for key, value in var.proxy_nodes : key => value
        if value.create
    }
    manager_to_create = {
        for key, value in var.manager_nodes : key => value
        if value.create
    }
}

resource "digitalocean_droplet" "node_instance" {
    for_each = local.nodes_to_create

    image = "${var.node_droplet_image}"
    name = "${var.node_droplet_name}"
    region = "${var.node_droplet_region}"
    size = "${var.node_droplet_size}"
    droplet_agent = false

    ssh_keys = [
        data.digitalocean_ssh_key.terraform.id
    ]
    
    connection {
		host = self.ipv4_address
        user = "root"
        type = "ssh"
        private_key = "${file(var.pvt_key)}"
        timeout = "15m"
    }
}

resource "digitalocean_droplet" "manager_instance" {
    for_each = local.manager_to_create
    image = "${var.manager_droplet_image}"
    name = "${var.manager_droplet_name}"
    region = "${var.manager_droplet_region}"
    size = "${var.manager_droplet_size}"
    droplet_agent = false

    ssh_keys = [
        data.digitalocean_ssh_key.terraform.id
    ]
    
    connection {
		host = self.ipv4_address
        user = "root"
        type = "ssh"
        private_key = "${file(var.pvt_key)}"
        timeout = "15m"
    }
}

output "node_public_ips" {
  value = [for droplet in digitalocean_droplet.node_instance : droplet.ipv4_address]
}

output "manager_public_ips" {
  value = [for droplet in digitalocean_droplet.manager_instance : droplet.ipv4_address]
}