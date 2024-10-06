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

# Snapshot ID for node
variable "node_droplet_image" {
  description = "The Digital Ocean Snapshot ID that was returned from Packer"
}

# Snapshot ID for manager
variable "manager_droplet_image" {
  description = "The Digital Ocean Snapshot ID that was returned from Packer"
}

# Node table
# The way that this works is we have a table of proxy nodes, and initially we create all 8.
# As a node needs to be torn down, create is changed to false, which leads to terraform plan 
# suggesting the deleting of that node, and then removing it when terraform apply is ran.
# Then, we set create back to true, and rerun terraform plan and then apply it, which will 
# artifically "recreate" nodes that needed to be torn down.
#
# The reason we are tracking private IP addresses for the nodes statically is because
# it will allow for us to be able to pass clients off really easily- for example, we can send 
# the public ip of proxy-2 from proxy-1 as the next IP to a client by calling 192.168.1.102, an
# ip that can be accessed by simply adding 1 to the current private IP and looping around to 1 on 
# 9.
#
# This does leave the door open for a client to cycle fully through all proxy ip addresses, which is a problem,
# but ideally we tear down from 1 to 8 and then cycle again, so if a client is down 
variable "proxy_nodes" {
    type = map(object({
        create  = bool
    }))
    default = {
        "proxy-1" = { create = true }
        "proxy-2" = { create = true }
        "proxy-3" = { create = true }
        "proxy-4" = { create = true }
        "proxy-5" = { create = true }
        "proxy-6" = { create = true }
        "proxy-7" = { create = true }
    }
}

# Same as above but for manager 
variable "manager_nodes" {
    type = map(object({
        create = bool
    }))
    default = {
        "manager-1" = { create = true }
    }
}