#!/bin/bash
# This file is to build all of the go images before utilizing ansible scripts to make sure the most up to date build is provided.
# First, we need to make sure we are in the correct directory
SCRIPT_DIR="$(dirname "$(realpath "$0")")"

# Define all of the directories for accessing Go builds
SERVICE_DIR="${SCRIPT_DIR}/service/"
MANAGER_DIR="${SERVICE_DIR}/manager/"
CLIENT_DIR="${SERVICE_DIR}/client/"
NODE_DIR="${SERVICE_DIR}/node/"

# Directory for terraform
TERRAFORM_DIR="${SCRIPT_DIR}/automation/terraform/digitalocean"

# Directory for packer builds
PACKER_DIR="${SCRIPT_DIR}/automation/packer"

# Directory for ansible builds
ANSIBLE_DIR="${SCRIPT_DIR}/automation/ansible/"
MANAGER_TEMPLATE_DIR="${ANSIBLE_DIR}/roles/manager/templates/"
NODE_TEMPLATE_DIR="${ANSIBLE_DIR}/roles/node/templates/"

# Default values
verbose=false
output_file=""
rebuild_go=false
destroy_terraform=false
destroy_snapshots=false
packer_image=""

# Options for what the script can provide on cli
usage() {
    echo "Usage: $0 [-v] [-o output_file] [-b] [-d] [-c] [-p target_json_file] [-h] argument"
    echo "Options:"
    echo "  -v            Enable verbose mode"
    echo "  -o FILE       Specify output file"
    echo "  -b            Rebuild Go package"
    echo "  -d            Destroy terraform environment"
    echo "  -c            Clean packer snapshots"
    echo "  -C            Clean entire workspace including terraform environment"
    echo "  -p FILE       Run packer on a target JSON file"
    echo "  -h            Display this help message"
    echo "Arguments:"
    echo "   deploy-manager  Deploys a manager first by building the packer image, and then deploying that image with terraform."
}

# Parse options
while getopts "vo:bcCdp:h" option; do
    case $option in
        v)
            verbose=true
            ;;
        o)
            output_file=$OPTARG
            ;;
        b)
            rebuild_go=true
            ;;
        d)
            destroy_terraform=true
            ;;
        c)
            clean_snapshots=true
            ;;
        C)
            destroy_terraform=true
            clean_snapshots=true
            ;;
        p)
            packer_image=$OPTARG
            ;;
        h | *)
            usage
            exit 0
            ;;
    esac
done

# Shift to remove processed options
shift $((OPTIND - 1))

# Update source so we can access our environment variables
source .env

# Function to print status messages in green or red
print_status() {
    if [ "$verbose" = false ] && [ "$output_file" != "" ]; then
        return
    fi

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

	# Print the message with padding. 
    echo -e "${color}${prefix} ${message} ${padding}${reset}"
    if [ "$output_file" != "" ]; then
        # Append to a log file
        echo -e "${color}${prefix} ${message} ${padding}${reset}" >> "$output_file"
    fi
}

# Destroy terraform which is a blocking action
if [ "$destroy_terraform" == true ]; then
    cd "${TERRAFORM_DIR}"
    print_status "Running terraform destroy" 2
    # Node and manager droplet image actually aren't all together too helpful in this scenario, as we are merely tearing down.
    terraform destroy -auto-approve -var="node_droplet_image=0" -var="manager_droplet_image=0"
    print_status "Terraform destroy complete" 1
fi

if [ "$clean_snapshots" == true ]; then
	red="\e[31m"    # Red
	reset="\e[0m"   # Reset color

    print_status "Cleaning snapshots" 2
    snapshots=$(curl -s -X GET "https://api.digitalocean.com/v2/snapshots" -H "Authorization: Bearer $DIGITAL_OCEAN_API_TOKEN" | jq -c '.snapshots[]')

    for snapshot in $snapshots; do
        #echo " * $snapshot"
        snapshot_id=$(echo "$snapshot" | jq -r '.id')
        snapshot_name=$(echo "$snapshot" | jq -r '.name')

        if [[ "$snapshot_name" == *"packer"* ]]; then
            echo -e " * Begin Deleting snapshot: ${red}${snapshot_name} ${reset}(${red}${snapshot_id}${reset})"
            curl -s -X DELETE "https://api.digitalocean.com/v2/snapshots/$snapshot_id" -H "Authorization: Bearer $DIGITAL_OCEAN_API_TOKEN"
            echo -e " * Snapshot ${red}removed${reset}"
        fi
    done

    print_status "Snapshots cleaned" 1
fi

# Cleans snapshots and terraform, and exits if we did.
if [ "$clean_snapshots" == true ] || [ "destroy_terraform" == true ]; then
    exit 0
fi

# Whether or not we have to rebuild and repackage our go binaries
if [ "$rebuild_go" == true ]; then
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
fi

# Options for this script
# Preimage: Create the packer image required for running terraform for the node
if [ "$packer_image" != "" ]; then
    cd "${PACKER_DIR}"
    packer build "$packer_image"
fi

if [ "$1" == "deploy" ]; then
    # Run the terraform required to create node, using the snapshot image provided by preimaging
    cd "${TERRAFORM_DIR}"
    terraform init

    print_status "Running terraform" 2
    terraform apply -auto-approve
fi