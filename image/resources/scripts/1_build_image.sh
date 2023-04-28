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

# Set variables
SHELL_YELLOW='\033[0;33m'
SHELL_DEFAULT='\033[0m'

export NEEDRESTART_MODE=a
export NEEDRESTART_SUSPEND=1
export DEBIAN_FRONTEND=noninteractive
export DEBIAN_PRIORITY=critical

export QUILT_PATCHES=debian/patches
export NAME=build EMAIL=build

patches="$(pwd)/patches/"
nproc="$(nproc)"

# begin_command() formats the terminal for a command output
begin_command() {
    echo -e "${SHELL_YELLOW}"
    echo "RUNNING STEP: $1..."
    echo -e "------${SHELL_DEFAULT}"
}

# begin_command() formats the terminal after command completion
complete_command() {
    echo -e -n "${SHELL_YELLOW}------ "
    echo "DONE"
    echo -e "${SHELL_DEFAULT}"

}

# install_nfs_packages() installs NFS Packages
install_nfs_packages() {

    begin_command "Installing rpcbind and nfs-kernel-server"
    apt-get update
    apt-get install -y rpcbind=1.2.5-8 nfs-kernel-server=1:1.3.4-2.5ubuntu3.4
    systemctl disable nfs-kernel-server
    systemctl disable nfs-idmapd.service
    complete_command

}

# install_build_dependencies() installs the dependencies to required to build the kernel
install_build_dependencies() {

    begin_command "Installing build dependencies"
    apt-get update
    apt-get install -y \
        libtirpc-dev libncurses-dev flex bison openssl libssl-dev dkms \
        libelf-dev libudev-dev libpci-dev libiberty-dev autoconf dwarves \
        build-essential libevent-dev libsqlite3-dev libblkid-dev \
        libkeyutils-dev libdevmapper-dev cdbs debhelper ubuntu-dev-tools
    complete_command

}

# download_nfs-utils() downloads version 2.5.3 of nfs-utils
download_nfs-utils() (

    begin_command "Downloading nfs-utils"
    echo -e "------${SHELL_DEFAULT}"
    curl -o nfs-utils-2.5.3.tar.gz https://mirrors.edge.kernel.org/pub/linux/utils/nfs-utils/2.5.3/nfs-utils-2.5.3.tar.gz
    tar xvf nfs-utils-2.5.3.tar.gz
    complete_command

)

# build_install_nfs-utils() builds and installs nfs-utils
build_install_nfs-utils() (

    begin_command "Building and installing nfs-utils"
    cd nfs-utils-2.5.3
    ./configure --prefix=/usr --sysconfdir=/etc --sbindir=/sbin --disable-gss
    make -j$nproc
    make install -j$nproc
    chmod u+w,go+r /sbin/mount.nfs
    chown nobody:nogroup /var/lib/nfs
    complete_command

)

# install_cachefilesd() builds and installs cachefilesd 0.10.10 from source
install_cachefilesd() (
    begin_command "Downloading cachefilesd"
    pull-lp-source cachefilesd 0.10.10-0.2ubuntu1

    begin_command "Building cachefilesd"
    cd cachefilesd-0.10.10
    quilt import "${patches}"/cachefilesd/*.patch
    quilt push -a
    debchange --local +knfsd "Apply knfsd patches"
    dpkg-buildpackage -b -us -uc

    begin_command "Installing cachefilesd"
    cd ..
    apt-get install -y ./cachefilesd_0.10.10-0.2ubuntu1+knfsd1_amd64.deb
    echo "RUN=yes" >> /etc/default/cachefilesd
    systemctl disable cachefilesd

    complete_command

)

# install_stackdriver_agent() installs the Cloud Ops Agent for metrics
install_stackdriver_agent() {

    begin_command "Installing Cloud Ops Agent dependencies"
    cd ops-agent
    curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
    bash add-google-cloud-ops-agent-repo.sh --also-install --version=2.22.0
    systemctl disable google-cloud-ops-agent
    cp google-cloud-ops-agent.conf /etc/logrotate.d/
    cd ..
    complete_command

}

# install_golang() installs golang
install_golang() {

    begin_command "Installing golang"
    curl -o go1.17.3.linux-amd64.tar.gz https://dl.google.com/go/go1.17.3.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    complete_command

}

# install_knfsd_agent() installs the knfsd-agent (see https://github.com/GoogleCloudPlatform/knfsd-cache-utils/tree/main/image/knfsd-agent)
install_knfsd_agent() (

    begin_command "Installing Knfsd agent"
    cd knfsd-agent/src
    go build -o /usr/local/bin/knfsd-agent *.go
    cd ..
    cp knfsd-logrotate.conf /etc/logrotate.d/
    cp knfsd-agent.service /etc/systemd/system/
    complete_command

)

# Install_knfsd_metrics_agent() installs the custom Knfsd Metrics Agent
install_knfsd_metrics_agent() (

    begin_command "Installing knfsd-metrics-agent"

    cd knfsd-metrics-agent
    go build -o /usr/local/bin/knfsd-metrics-agent

    mkdir /etc/knfsd-metrics-agent
    cp config/*.yaml /etc/knfsd-metrics-agent/
    cp systemd/proxy.service /etc/systemd/system/knfsd-metrics-agent.service

    complete_command
)

install_knfsd_cull() (
    begin_command "Installing knfsd-cull"

    cd knfsd-cull
    go test ./...
    go build -o /usr/local/bin/knfsd-cull
    cp knfsd-cull.conf /etc/
    cp knfsd-cull.service /etc/systemd/system/

    complete_command
)

# install_filter_exports installs the agent that filters NFS Exports
install_filter_exports() (
    begin_command "Installing filter-exports"
    cd filter-exports
    go test ./...
    go build -o /usr/local/bin/filter-exports
    complete_command
)

# install_netapp_exports() installs the NetApp export detection service
install_netapp_exports() (
    begin_command "Installing netapp-exports"
    cd netapp-exports
    go test ./...
    go build -o /usr/local/bin/netapp-exports
    echo -e -n "${SHELL_YELLOW}------ "
    complete_command
)

# install_kernel() installs the latest 5.13 kernel
install_kernel() {

    # Install the new kernel using dpkg
    begin_command "Installing kernel"
    apt-get install -y 'linux-generic-hwe-20.04=5.13.*' \
        'linux-image-generic-hwe-20.04=5.13.*' \
        'linux-headers-generic-hwe-20.04=5.13.*'

    complete_command

}

# copy_config() copies the NFS Server configuration
copy_config() {
    chown --recursive root:root etc
    chmod --recursive 0644 etc
    cp --recursive ./etc /
    mkdir -p /srv/nfs
}

# Run Build
install_nfs_packages
install_build_dependencies
download_nfs-utils
build_install_nfs-utils
install_cachefilesd
install_stackdriver_agent
install_golang
install_knfsd_agent
install_knfsd_metrics_agent
install_knfsd_cull
install_filter_exports
install_netapp_exports

# 5.13 HWE packages have been removed from the 20.04 distro.
# Currently using a GCP image that comes with 5.13.
# install_kernel

copy_config

echo
echo
echo "SUCCESS: Please reboot for new kernel to take effect"
echo -e "${SHELL_DEFAULT}"
