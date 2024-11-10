#!/bin/bash -eux
export DEBIAN_FRONTEND=noninteractive

# Clean and remove the apt lists to ensure proper install
# sudo apt-get clean
# sudo rm -rf /var/lib/apt/lists/*
# sleep 4

# Update package lists and upgrade existing packages.
apt-get -o DPkg::Lock::Timeout=-1 update -y

# Install prerequisites for adding repositories.
apt-get -o DPkg::Lock::Timeout=-1 install -y software-properties-common

# Add the Ansible repository.
add-apt-repository -o DPkg::Lock::Timeout=-1 --yes --update ppa:ansible/ansible

# Install Ansible.
apt-get -o DPkg::Lock::Timeout=-1 install -y ansible