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

disable_unattended_upgrades() {
    # Stop unattended-upgrades while building the image, otherwise apt-get might
    # fail because the apt cache is locked by the unattended-upgrade service.
    #
    # Bug in current (6.11.0) HWE kernel can cause a kernel panic when the
    # nfs-server.service is restarted. Unattended upgrades can restart this
    # service if upgrading the NFS packages (or libraries NFS depends on).
    #
    # To avoid this issue, disabling unattended-upgrades.service.
    begin_command "Disabling unattended-upgrades.service"
    systemctl disable unattended-upgrades.service
    complete_command
}

# install_nfs_packages() installs NFS Packages
install_nfs_packages() {

    begin_command "Installing rpcbind and nfs-kernel-server"
    apt-get update
    apt-get install -y rpcbind nfs-kernel-server
    systemctl disable nfs-kernel-server
    systemctl disable nfs-idmapd.service
    complete_command

}

# install_build_dependencies() installs the dependencies to required to build
# the kernel and nfs-utils
install_build_dependencies() {
    begin_command "Installing build dependencies"
    apt-get update
    apt-get install -y ubuntu-dev-tools cdbs debhelper
    complete_command
}

install_cachefilesd() (
    begin_command "Building and installing cachefilesd"
    echo -e "------${SHELL_DEFAULT}"

    pull-lp-source cachefilesd 0.10.10-0.2ubuntu1
    cd cachefilesd-0.10.10/

    quilt import "$patches"/cachefilesd/*.patch
    quilt push -a

    debchange --local +knfsd "Applying custom patches"
    debuild -i -uc -us -b

    cd ..
    apt-get install -y \
        ./cachefilesd_0.10.10-0.2ubuntu1+knfsd1_amd64.deb \
        ./cachefilesd-dbgsym_0.10.10-0.2ubuntu1+knfsd1_amd64.ddeb

    systemctl disable cachefilesd
    echo "RUN=yes" >> /etc/default/cachefilesd
)

# install_stackdriver_agent() installs the Cloud Ops Agent for metrics
install_stackdriver_agent() {

    begin_command "Installing Cloud Ops Agent dependencies"
    cd ops-agent
    curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
    bash add-google-cloud-ops-agent-repo.sh --also-install --version=2.49.0
    systemctl disable google-cloud-ops-agent
    cp google-cloud-ops-agent.conf /etc/logrotate.d/
    cd ..
    complete_command

}

# install_golang() installs golang
install_golang() {

    begin_command "Installing golang"
    curl -o go1.20.1.linux-amd64.tar.gz https://dl.google.com/go/go1.20.1.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.1.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    complete_command

}

install_fsidd_service() {
    begin_command "Installing knfsd-fsidd service"
    cd knfsd-fsidd
    go build -o /usr/local/sbin/knfsd-fsidd
    cd ..
    complete_command
}

# install_knfsd_agent() installs the knfsd-agent (see https://github.com/GoogleCloudPlatform/knfsd-cache-utils/tree/main/image/knfsd-agent)
install_knfsd_agent() (

    begin_command "Installing Knfsd agent"
    cd knfsd-agent
    go build -o /usr/local/bin/knfsd-agent
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
    complete_command
)

# copy_config() copies the NFS Server configuration
copy_config() {
    chown --recursive root:root etc
    chmod --recursive 0644 etc
    cp --recursive ./etc /
    mkdir -p /srv/nfs
}

# Run Build
disable_unattended_upgrades
install_nfs_packages
install_build_dependencies
install_cachefilesd
install_stackdriver_agent
install_golang
install_fsidd_service
install_knfsd_agent
install_knfsd_metrics_agent
install_filter_exports
install_netapp_exports
copy_config

echo
echo
echo -e -n "${SHELL_YELLOW}"
echo "SUCCESS: Please reboot for new kernel to take effect"
echo -e "${SHELL_DEFAULT}"
