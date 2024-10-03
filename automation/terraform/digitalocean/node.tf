resource "digitalocean_droplet" "node_instance" {
    image = "${var.node_droplet_image}"
    name = "${var.node_droplet_name}"
    region = "${var.node_droplet_region}"
    size = "${var.node_droplet_size}"

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