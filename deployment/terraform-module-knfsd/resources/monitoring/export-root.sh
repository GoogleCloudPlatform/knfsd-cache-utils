# !/bin/bash 
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

# Make a directory for stats export
mkdir -p /statsexport

# Create a file for nfs_inode_cache_active_objects and populate an initial value
touch /statsexport/nfs_inode_cache_active_objects
chown "nobody" /statsexport/nfs_inode_cache_active_objects
nfs_inode_cache_active_objects=$(cat /proc/slabinfo | grep nfs_inode | awk '{print $2}')
echo "$nfs_inode_cache_active_objects" > /statsexport/nfs_inode_cache_active_objects

# Create a file for nfs_inode_cache_objsize and populate an initial value
touch /statsexport/nfs_inode_cache_objsize
chown "nobody" /statsexport/nfs_inode_cache_objsize
nfs_inode_cache_objsize=$(cat /proc/slabinfo | grep nfs_inode | awk '{print $4}')
echo "$nfs_inode_cache_objsize" > /statsexport/nfs_inode_cache_objsize

# Create a file for nfs_inode_cache_active_objects and populate an initial value
touch /statsexport/dentry_cache_active_objects
chown "nobody" /statsexport/dentry_cache_active_objects
dentry_cache_active_objects=$(cat /proc/slabinfo | grep dentry | awk '{print $2}')
echo "$dentry_cache_active_objects" > /statsexport/dentry_cache_active_objects

# Create a file for dentry_cache_objsize and populate an initial value
touch /statsexport/dentry_cache_objsize
chown "nobody" /statsexport/dentry_cache_objsize
dentry_cache_objsize=$(cat /proc/slabinfo | grep dentry | awk '{print $4}')
echo "$dentry_cache_objsize" > /statsexport/dentry_cache_objsize

# Loop that runs and updates the files with updated values every 60 seconds
while sleep 60
do
    # Export nfs_inode_cache_active_objects to a file that CollectD can read
    nfs_inode_cache_active_objects=$(cat /proc/slabinfo | grep nfs_inode | awk '{print $2}')
    echo "$nfs_inode_cache_active_objects" > /statsexport/nfs_inode_cache_active_objects

    # Export nfs_inode_cache_objsize to a file that CollectD can read
    nfs_inode_cache_objsize=$(cat /proc/slabinfo | grep nfs_inode | awk '{print $4}')
    echo "$nfs_inode_cache_objsize" > /statsexport/nfs_inode_cache_objsize

    # Export dentry_cache_active_objects to a file that CollectD can read
    dentry_cache_active_objects=$(cat /proc/slabinfo | grep dentry | grep -v ext4_fc_dentry_update | awk '{print $2}')
    echo "$dentry_cache_active_objects" > /statsexport/dentry_cache_active_objects

    # Export dentry_cache_objsize to a file that CollectD can read
    dentry_cache_objsize=$(cat /proc/slabinfo | grep dentry | grep -v ext4_fc_dentry_update | awk '{print $4}')
    echo "$dentry_cache_objsize" > /statsexport/dentry_cache_objsize

done