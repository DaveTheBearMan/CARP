#!/bin/bash -eux

# Uninstall Ansible and remove PPA.
apt-get -o DPkg::Lock::Timeout=-1 remove --purge ansible -y
apt-add-repository --remove ppa:ansible/ansible

# Apt cleanup.
apt-get -o DPkg::Lock::Timeout=-1 autoremove -y
apt-get -o DPkg::Lock::Timeout=-1 update -y
