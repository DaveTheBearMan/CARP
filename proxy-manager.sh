#!/bin/bash
# This file is to build all of the go images before utilizing ansible scripts to make sure the most up to date build is provided.
# First, we need to make sure we are in the correct directory
SCRIPT_DIR="$(find ~ -type d -name "CloudFlareway")"

# Define all of the directories for accessing Go builds
SERVICE_DIR="${SCRIPT_DIR}/service/"
MANAGER_DIR="${SERVICE_DIR}/manager/"
CLIENT_DIR="${SERVICE_DIR}/client/"
NODE_DIR="${SERVICE_DIR}/node/"

# Directory for terraform
TERRAFORM_DIR="${SCRIPT_DIR}/automation/terraform/digitalocean"

# Directory for snapshots
SNAPSHOT_DIR="${SCRIPT_DIR}/snapshot_ids"

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
    echo "  -v                                  Enable verbose mode"
    echo "  -o FILE                             Specify output file"
    echo "  -b                                  Rebuild Go package"
    echo "  -d                                  Destroy terraform environment"
    echo "  -c                                  Clean packer snapshots"
    echo "  -C                                  Clean entire workspace including terraform environment"
    echo "  -p FILE                             Run packer on a target JSON file"
    echo "  -h                                  Display this help message"
    echo "Arguments:"
    echo "   deploy                             Deploys current terraform configuration"
    echo "   rebuild <proxy/manager> <index>    Rebuilds a specified node"
    echo "   build                              Builds the entire proxy manager, destructively."
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
cd "${SCRIPT_DIR}"
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

    # Remove snapshot ids and make new directory for storing them
    cd "${SCRIPT_DIR}"
    rm -rf "${SNAPSHOT_DIR}"
    mkdir -p "${SNAPSHOT_DIR}"

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
    # Build the packer image
    cd "${PACKER_DIR}"
    print_status "Beginning packer build!" 2
    packer build "${packer_image}.pkr.hcl" | tee build_temp.txt
    print_status "Packer build successful." 1

    # First get the line with the snapshot ID. #9 is the position of the snapshot id currently (ID: ######)
    snapshot_id=$(cat build_temp.txt | grep "A snapshot was created" | awk '{print $9}' | sed 's/.$//')
    # Get the first word from file path (ex: node.pkr.hcl -> node)
    snapshot_type=$(echo $packer_image | sed 's/\..*//')
    #rm build_temp.txt

    # Put most recent image onto the top of the snapshot file
    print_status "Writing snapshot id to file." 2
    cd "${SNAPSHOT_DIR}"
    touch "${snapshot_type}.txt" # Ensure file exists
    echo -e "${snapshot_id}\n$(cat "${snapshot_type}.txt")" > "${snapshot_type}.txt"
fi

if [ "$1" == "deploy" ]; then
    # Run the terraform required to create node, using the snapshot image provided by preimaging
    cd "${TERRAFORM_DIR}"
    terraform init

    print_status "Running terraform" 2
    terraform apply -auto-approve -var="node_droplet_image=$(cat ${SNAPSHOT_DIR}/node.txt | head -n 1)" -var="manager_droplet_image=$(cat ${SNAPSHOT_DIR}/manager.txt | head -n 1)"

    terraform output | sed -e 's/  "/      /g' -e 's/",/:/g' -e 's/]//g' -e 's/ = \[/:/' -e 's/[*]*/    &/' -e '1 s/[*]*/all:\n  children:\n/' -e 's/public_ips:/&\n      hosts:/' | grep "\S" > ../../ansible/inventory.yml

elif [ "$1" == "rebuild" ]; then
    # Run the terraform required to create node, using the snapshot image provided by preimaging
    cd "${TERRAFORM_DIR}"

    print_status "Tearing image down" 2
    if [ "$2" == "proxy" ]; then
        if [[ $3 =~ ^-?[0-9]+$ ]]; then # Check if its an integer
            # Match to get proxy node we're removing then building
            pattern="proxy-${3}"
            print_status "Removing ${pattern}" 2
            sleep 2

            # Stop creating the instance
            sed -i s/"\"${pattern}\" = { create = true }"/"\"${pattern}\" = { create = false }"/ variables.tf

            print_status "Running terraform to bring node down" 2
            terraform apply -auto-approve -var="node_droplet_image=$(cat ${SNAPSHOT_DIR}/node.txt | head -n 1)" -var="manager_droplet_image=$(cat ${SNAPSHOT_DIR}/manager.txt | head -n 1)"

            # Create the instance
            sed -i s/"\"${pattern}\" = { create = false }"/"\"${pattern}\" = { create = true }"/ variables.tf

            print_status "Running terraform to bring node up" 2
            terraform apply -auto-approve -var="node_droplet_image=$(cat ${SNAPSHOT_DIR}/node.txt | head -n 1)" -var="manager_droplet_image=$(cat ${SNAPSHOT_DIR}/manager.txt | head -n 1)"
        fi
    elif [ "$2" == "manager" ]; then
        echo "Create manager"
    else
        echo "Invalid argument"
    fi
elif [ "$1" == "build" ]; then
    clear
    if [ "$2" == "in-place" ]; then
        echo "Not implemented"
    else
        while true; do
            read -p "Are you sure? This will rebuild the entire proxy-manager service. [y/n] " yn
            case $yn in
                [Yy]* ) print_status "Building proxy-manager" 2; break;;
                [Nn]* ) exit;;
                * ) echo "Please answer yes or no.";;
            esac
        done
        
        proxy-manager -v -C
        proxy-manager -v -b
        proxy-manager -v -p node
        proxy-manager -v -p manager
        proxy-manager -v deploy  
    fi
fi
