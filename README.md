<h1><p align="center">
  C.A.R.P
  <hr>
  <img src="https://github.com/user-attachments/assets/df84d888-b9b6-45ec-898e-874aa5facd0b" border=10>
  <hr>
</p></h1>

Cloud Agnostic Rotating Proxy, or CARP, utilizes packer and terraform to generate images that can be deployed to any cloud provider. Currently, while still being produced the tool is designed to work on Digital Ocean, however, as time goes on more providers will be added. 

---
### Set up guide
To set up the tool, create a symbolic link between the executable and /usr/bin/proxy-manager so that if you pull an update it automatically updates the service.
After that, you can review any helpful commands with `proxy-manager -h` or refer to this guide.

### Help Flag
```
Usage: /usr/bin/proxy-manager [-v] [-o output_file] [-b] [-d] [-c] [-p target_json_file] [-h] argument
Options:
  -v                                  Enable verbose mode
  -o FILE                             Specify output file
  -b                                  Rebuild Go package
  -d                                  Destroy terraform environment
  -c                                  Clean packer snapshots
  -C                                  Clean entire workspace including terraform environment
  -p FILE                             Run packer on a target JSON file
  -h                                  Display this help message
Arguments:
   deploy                             Deploys current terraform configuration
   rebuild <proxy/manager> <index>    Rebuilds a specified node
```
