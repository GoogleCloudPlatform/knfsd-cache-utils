#!/bin/bash
#
# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Exit on command failure
set -e

# Set shell color variables
SHELL_YELLOW='\033[0;33m'
SHELL_DEFAULT='\033[0m'

# install_nfs_packages() installs NFS Packages
install_nfs_packages() {

    # Install Cachefilesd
    echo "Installing cachefilesd and rpcbind..."
    echo -e "------${SHELL_DEFAULT}"
    apt-get update
    apt-get install -y cachefilesd=0.10.10-0.2ubuntu1 rpcbind=1.2.5-8 nfs-kernel-server=1:1.3.4-2.5ubuntu3.4 tree
    echo "RUN=yes" >> /etc/default/cachefilesd
    systemctl disable cachefilesd
    systemctl disable nfs-kernel-server
    systemctl disable nfs-idmapd.service
    echo -e -n "${SHELL_YELLOW}------"
    echo "DONE"

}

# install_build_dependencies() installs the dependencies to required to build the kernel
install_build_dependencies() {

    echo -e "${SHELL_YELLOW}"
    echo "Installing build dependencies..."
    echo -e "------${SHELL_DEFAULT}"
    apt-get update
    apt-get upgrade -y
    apt-get install libtirpc-dev libncurses-dev flex bison openssl libssl-dev dkms libelf-dev libudev-dev libpci-dev libiberty-dev autoconf dwarves build-essential libevent-dev libsqlite3-dev libblkid-dev libkeyutils-dev libdevmapper-dev -y
    echo -e -n "${SHELL_YELLOW}------ "
    echo "DONE"

}

# download_nfs-utils() downloads version 2.5.3 of nfs-utils
download_nfs-utils() (

    echo -e "${SHELL_YELLOW}"
    echo "Downloading nfs-utils..."
    echo -e "------${SHELL_DEFAULT}"
    curl -o nfs-utils-2.5.3.tar.gz https://mirrors.edge.kernel.org/pub/linux/utils/nfs-utils/2.5.3/nfs-utils-2.5.3.tar.gz
    tar xvf nfs-utils-2.5.3.tar.gz
    echo -e -n "${SHELL_YELLOW}------"
    echo "DONE"

)

# build_install_nfs-utils() builds and installs nfs-utils
build_install_nfs-utils() (

    echo -e "${SHELL_YELLOW}"
    echo "Installing nfs-utils..."
    echo -e "------${SHELL_DEFAULT}"
    cd nfs-utils-2.5.3
    ./configure --prefix=/usr --sysconfdir=/etc --sbindir=/sbin --disable-gss
    make -j20
    make install -j20
    chmod u+w,go+r /sbin/mount.nfs
    chown nobody.nogroup /var/lib/nfs
    echo -e -n "${SHELL_YELLOW}------"
    echo "DONE"

)

# install_stackdriver_agent() installs the Stackdriver Agent for metrics
install_stackdriver_agent() {

    echo -e "${SHELL_YELLOW}"
    echo "Installing Stackdriver Agent dependencies..."
    echo -e "------${SHELL_DEFAULT}"
    curl -sSO https://dl.google.com/cloudagents/add-monitoring-agent-repo.sh
    bash add-monitoring-agent-repo.sh
    apt-get update
    sudo apt-get install -y stackdriver-agent=6.1.4-1.focal
    systemctl disable stackdriver-agent
    echo -e -n "${SHELL_YELLOW}------ "
    echo "DONE"

}

# install_golang() installs golang
install_golang() {

    echo "Installing golang...."
    echo -e "------${SHELL_DEFAULT}"
    curl -o go1.17.3.linux-amd64.tar.gz https://dl.google.com/go/go1.17.3.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo -e -n "${SHELL_YELLOW}------ "
    echo "DONE"

}

# install_knfsd_agent() installs the knfsd-agent (see https://github.com/GoogleCloudPlatform/knfsd-cache-utils/tree/main/image/knfsd-agent)
install_knfsd_agent() (

    echo "Installing knfsd-agent...."
    echo -e "------${SHELL_DEFAULT}"
    cd knfsd-agent/src
    go build -o /usr/local/bin/knfsd-agent *.go
    cd ..
    cp knfsd-logrotate.conf /etc/logrotate.d/
    cp knfsd-agent.service /etc/systemd/system/
    echo -e -n "${SHELL_YELLOW}------ "
    echo "DONE"

)

install_netapp_exports() (
    echo "Installing netapp-exports...."
    echo -e "------${SHELL_DEFAULT}"
    cd netapp-exports
    go test ./...
    go build -o /usr/local/bin/netapp-exports
    echo -e -n "${SHELL_YELLOW}------ "
    echo "DONE"
)

# download_kernel() downloads the 5.11.8 Kernel
download_kernel() {

    # Make directory for kernel Images
    echo -n "Creating directory for kernel Images... "
    mkdir -p kernel-images
    echo "DONE"

    # Download Kernel .deb packages from kernel.ubuntu.com
    echo "Downloading kernel .deb files..."
    echo -e "------${SHELL_DEFAULT}"
    curl -o kernel-images/linux-headers-5.11.8-051108-generic_5.11.8-051108.202103200636_amd64.deb https://kernel.ubuntu.com/~kernel-ppa/mainline/v5.11.8/amd64/linux-headers-5.11.8-051108-generic_5.11.8-051108.202103200636_amd64.deb
    curl -o kernel-images/linux-headers-5.11.8-051108_5.11.8-051108.202103200636_all.deb https://kernel.ubuntu.com/~kernel-ppa/mainline/v5.11.8/amd64/linux-headers-5.11.8-051108_5.11.8-051108.202103200636_all.deb
    curl -o kernel-images/linux-image-unsigned-5.11.8-051108-generic_5.11.8-051108.202103200636_amd64.deb https://kernel.ubuntu.com/~kernel-ppa/mainline/v5.11.8/amd64/linux-image-unsigned-5.11.8-051108-generic_5.11.8-051108.202103200636_amd64.deb
    curl -o kernel-images/linux-modules-5.11.8-051108-generic_5.11.8-051108.202103200636_amd64.deb https://kernel.ubuntu.com/~kernel-ppa/mainline/v5.11.8/amd64/linux-modules-5.11.8-051108-generic_5.11.8-051108.202103200636_amd64.deb
    echo -e -n "${SHELL_YELLOW}------"
    echo "DONE"

}

# install_kernel() installs the 5.11.8 kernel
install_kernel() {

    # Install the new kernel using dpkg
    echo "Installing kernel...."
    echo -e "------${SHELL_DEFAULT}"
    dpkg -i kernel-images/*
    echo -e -n "${SHELL_YELLOW}------"
    echo "DONE"

}

copy_config() {
    chown --recursive root:root etc
    chmod --recursive 0644 etc
    cp --recursive ./etc /
}

# Prep Server
install_nfs_packages
install_build_dependencies
download_nfs-utils
build_install_nfs-utils
install_stackdriver_agent
install_golang
install_knfsd_agent
install_netapp_exports
download_kernel
install_kernel
copy_config

echo
echo
echo "SUCCESS: Please reboot for new kernel to take effect"
echo -e "${SHELL_DEFAULT}"
