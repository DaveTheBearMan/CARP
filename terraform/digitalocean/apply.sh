#!/bin/bash

# Set terraform apply to log instead of print
export TF_LOG=DEBUG

# Create log directory in case this is on a new system
mkdir -p ./logs
yes 'yes' | terraform apply > ./logs/"Apply - $(date)" 2>&1
