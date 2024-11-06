#!/bin/bash

# Kill old session
tmux kill-sesssion -t manager

# Start the tmux session if it doesn't exist
tmux new-session -d -s manager

# Send the command to the tmux session
tmux send-keys -t manager:0 '/usr/local/bin/manager' C-m

# Keep the script running to prevent immediate exit
# You can customize this as needed, for example, waiting indefinitely
while true; do sleep 60; done