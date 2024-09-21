#!/bin/bash

# Set terraform to log to a file instead of stdout
export TF_LOG=DEBUG

# Create directory in case it doesnt exist
mkdir -p ./logs
yes 'yes' | terraform destroy > ./logs/"Destroy - $(date)" 2>&1
