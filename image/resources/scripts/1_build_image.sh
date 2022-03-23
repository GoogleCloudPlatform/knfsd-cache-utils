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

export NEEDRESTART_MODE=a
export NEEDRESTART_SUSPEND=1
export DEBIAN_FRONTEND=noninteractive
export DEBIAN_PRIORITY=critical

# install_nfs_packages() installs NFS Packages
install_nfs_packages() {

    # Install Cachefilesd
    echo "Installing cachefilesd and rpcbind..."
    echo -e "------${SHELL_DEFAULT}"
    apt-get update
    apt-get install -y cachefilesd=0.10.10-0.2ubuntu1 rpcbind=1.2.5-9build1 nfs-kernel-server=1:1.3.4-6ubuntu1
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
    apt-get install -y libtirpc-dev libncurses-dev flex bison openssl libssl-dev dkms libelf-dev libudev-dev libpci-dev libiberty-dev autoconf dwarves build-essential libevent-dev libsqlite3-dev libblkid-dev libkeyutils-dev libdevmapper-dev
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

# install_stackdriver_agent() installs the Cloud Ops Agent for metrics
install_stackdriver_agent() {

    echo -e "${SHELL_YELLOW}"
    echo "Installing Cloud Ops Agent dependencies..."
    echo -e "------${SHELL_DEFAULT}"
    curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
    bash add-google-cloud-ops-agent-repo.sh --also-install --version=2.11.0
    systemctl disable google-cloud-ops-agent
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

install_knfsd_metrics_agent() (
    echo "Installing knfsd-metrics-agent...."
    echo -e "------${SHELL_DEFAULT}"

    cd knfsd-metrics-agent
    go build -o /usr/local/bin/knfsd-metrics-agent

    mkdir /etc/knfsd-metrics-agent
    cp config/*.yaml /etc/knfsd-metrics-agent/
    cp systemd/proxy.service /etc/systemd/system/knfsd-metrics-agent.service

    echo -e -n "${SHELL_YELLOW}------ "
    echo "DONE"
)

install_filter_exports() (
    echo "Installing netapp-exports...."
    echo -e "------${SHELL_DEFAULT}"
    cd filter-exports
    go test ./...
    go build -o /usr/local/bin/filter-exports
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
download_kernel() (

    # Make directory for kernel Images
    echo -n "Creating directory for kernel Images... "
    mkdir -p kernel-images
    cd kernel-images
    echo "DONE"

    echo "Downloading kernel .deb files..."
    echo -e "------${SHELL_DEFAULT}"

    # libssl3 is not available in the 21.10 repositories, it will be included
    # in 22.04. For now download the package directly and install it.
    curl -O http://mirrors.edge.kernel.org/ubuntu/pool/main/o/openssl/libssl3_3.0.1-0ubuntu1_amd64.deb

    # Download 5.17 from mainline as is not available through Ubuntu's hardware enablement yet.
    curl -O https://kernel.ubuntu.com/~kernel-ppa/mainline/v5.17/amd64/linux-headers-5.17.0-051700-generic_5.17.0-051700.202203202130_amd64.deb
    curl -O https://kernel.ubuntu.com/~kernel-ppa/mainline/v5.17/amd64/linux-headers-5.17.0-051700_5.17.0-051700.202203202130_all.deb
    curl -O https://kernel.ubuntu.com/~kernel-ppa/mainline/v5.17/amd64/linux-image-unsigned-5.17.0-051700-generic_5.17.0-051700.202203202130_amd64.deb
    curl -O https://kernel.ubuntu.com/~kernel-ppa/mainline/v5.17/amd64/linux-modules-5.17.0-051700-generic_5.17.0-051700.202203202130_amd64.deb

    echo -e -n "${SHELL_YELLOW}------"
    echo "DONE"

)

# install_kernel() installs the 5.17-rc5 kernel
install_kernel() {

    # Install the new kernel using dpkg
    echo "Installing kernel...."
    echo -e "------${SHELL_DEFAULT}"
    dpkg -i kernel-images/*.deb
    echo -e -n "${SHELL_YELLOW}------"
    echo "DONE"

}

copy_config() {
    chown --recursive root:root etc
    chmod --recursive 0644 etc
    cp --recursive ./etc /
    mkdir -p /srv/nfs
}

# Prep Server
install_nfs_packages
install_build_dependencies
download_nfs-utils
build_install_nfs-utils
install_stackdriver_agent
install_golang
install_knfsd_agent
install_knfsd_metrics_agent
install_filter_exports
install_netapp_exports
download_kernel
install_kernel
copy_config

echo
echo
echo "SUCCESS: Please reboot for new kernel to take effect"
echo -e "${SHELL_DEFAULT}"
