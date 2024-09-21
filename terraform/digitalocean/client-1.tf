resource "digitalocean_droplet" "http-proxies" {
	count=2
	image = "ubuntu-24-04-x64"
	name = "proxy-${count.index}"
	region = "nyc3"
	size = "s-1vcpu-1gb"
	ssh_keys = [
		data.digitalocean_ssh_key.terraform.id
	]

	connection {
		host = self.ipv4_address
		user = "root"
		type = "ssh"
		private_key = file(var.pvt_key)
		timeout = "2m"
	}
}

resource "digitalocean_droplet" "http-manager" {
	image = "ubuntu-24-04-x64"
	name = "manager"
	region = "nyc3"
	size = "s-2vcpu-2gb"
	ssh_keys = [
		data.digitalocean_ssh_key.terraform.id
	]

	connection {
		host = self.ipv4_address
		user = "root"
		type = "ssh"
		private_key = file(var.pvt_key)
		timeout = "2m"
	}
}
