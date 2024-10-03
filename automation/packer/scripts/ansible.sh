#!/bin/bash -eux

# Update package lists and upgrade existing packages.
apt-get update -y && apt-get upgrade -y

# Install prerequisites for adding repositories.
apt-get install -y software-properties-common

# Add the Ansible repository.
add-apt-repository --yes --update ppa:ansible/ansible

# Install Ansible.
apt-get install -y ansible
