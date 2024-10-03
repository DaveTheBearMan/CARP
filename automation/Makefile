include .env
export $(grep -v '^#' .env | xargs)

node-packer-build:
	cd packer && \
	packer build node.json && \
	cd -
