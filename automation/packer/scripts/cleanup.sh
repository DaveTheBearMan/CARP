#!/bin/bash -eux

# Uninstall Ansible and remove PPA.
apt-get remove --purge ansible -y
apt-add-repository --remove ppa:ansible/ansible

# Apt cleanup.
apt-get autoremove -y
apt-get update -y
