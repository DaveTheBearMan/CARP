#!/bin/bash -eux
export DEBIAN_FRONTEND=noninteractive

# Clean and remove the apt lists to ensure proper install
# sudo apt-get clean
# sudo rm -rf /var/lib/apt/lists/*
# sleep 4

# Update package lists and upgrade existing packages.
apt-get -o DPkg::Lock::Timeout=900 update -y
sleep 4

# Install prerequisites for adding repositories.
apt-get -o DPkg::Lock::Timeout=900 install -y software-properties-common
sleep 4

# Add the Ansible repository.
add-apt-repository --yes --update ppa:ansible/ansible
sleep 4

# Install Ansible.
apt-get -o DPkg::Lock::Timeout=300 install -y ansible
sleep 4