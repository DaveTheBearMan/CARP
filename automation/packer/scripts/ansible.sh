#!/bin/bash -eux
export DEBIAN_FRONTEND=noninteractive

# Clean and remove the apt lists to ensure proper install
sudo apt-get clean
sudo rm -rf /var/lib/apt/lists/*

# Sleep for 2 seconds for Dpkg lock
sleep 2

# Update package lists and upgrade existing packages.
apt-get update -y

# Wait for dpkg lock before moving on
sleep 2

# Install prerequisites for adding repositories.
apt-get install -y software-properties-common

# Add the Ansible repository.
add-apt-repository --yes --update ppa:ansible/ansible

# Install Ansible.
apt-get install -y ansible
