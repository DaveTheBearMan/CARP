#!/bin/bash
# This file is to build all of the go images before utilizing ansible scripts to make sure the most up to date build is provided.
# First, we need to make sure we are in the correct directory
SCRIPT_DIR="$(dirname "$(realpath "$0")")"

# Define all of the directories for accessing Go builds
SERVICE_DIR="${SCRIPT_DIR}/service/"
MANAGER_DIR="${SERVICE_DIR}/manager/"
CLIENT_DIR="${SERVICE_DIR}/client/"
NODE_DIR="${SERVICE_DIR}/node/"

# Directory for ansible builds
ANSIBLE_DIR="${SCRIPT_DIR}/automation/ansible/"
MANAGER_TEMPLATE_DIR="${ANSIBLE_DIR}/roles/manager/templates/"
NODE_TEMPLATE_DIR="${ANSIBLE_DIR}/roles/node/templates/"

# Function to print status messages in green or red
print_status() {
	local message="[$1]"
	local status="$2"  # 1 for success (green), 0 for failure (red)

	# Color codes
	local green="\e[32m"  # Green
	local red="\e[31m"    # Red
	local yellow="\e[33m" # Yellow
	local reset="\e[0m"   # Reset color

	# Determine color based on status
	if [ "$status" -eq 1 ]; then
		color="$green"
		prefix="SUCCESS"
	elif [ "$status" -eq 2 ]; then
		color="$yellow"
		prefix="BEGIN"
	else
		color="$red"
		prefix="FAILURE"
	fi

	# Calculate terminal width and padding
	local term_width
	term_width=$(tput cols) # Get the width of the terminal
	local total_length=$(( ${#prefix} + ${#message} + 2)) # Corrected to use arithmetic expansion
	local padding_length=$(( term_width - total_length )) # Calculate padding length

	# Ensure non-negative padding length
	if [ "$padding_length" -lt 0 ]; then
		padding_length=0
	fi

	# Generate right padding with asterisks
	local padding=""
	for ((i=0; i<padding_length; i++)); do
		padding+="*"
	done

	# Print the message with padding
	echo -e "${color}${prefix} ${message} ${padding}${reset}"
}

# Move into manager directory and attempt to build
cd "${MANAGER_DIR}"
go build -o manager
if [ $? -eq 0 ]; then
        print_status "Manager build successful!" 1
        chmod +x manager
        mv manager "${MANAGER_TEMPLATE_DIR}"
else
        print_status "Manager build failed!" 0
fi

# Move into node directory and attempt to build
cd "${NODE_DIR}"
go build -o node
if [ $? -eq 0 ]; then
        print_status "Node build successful!" 1
        chmod +x node
        mv node "${NODE_TEMPLATE_DIR}"
else
        print_status "Node build failed!" 0
fi

# Move to ansible directory and run ansible
cd "${ANSIBLE_DIR}"
print_status "Run ansible onto manager node" 2
ansible-playbook manager.yml -i inventory.yml -u root
print_status "Completed ansible manager playbook" 1

print_status "Run ansible playbook onto the nodes" 2
ansible-playbook node.yml -i inventory.yml -u root
print_status "Completed ansible node playbook" 1
