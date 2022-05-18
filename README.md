# knfsd-cache-utils

**Note:** This is not an officially supported Google product. NFS re-exporting is only suited for very limited use cases. You should be extremely familiar with NFS, FS-Cache and the other components used in this repository before deploying and/or using these utilities.

## Overview

This repository contains a set of utilities for building, deploying and operating a high performance NFS cache in Google Cloud. It is designed to be used for certain HPC and burst compute use-cases where there is a requirement for a high performance NFS cache between a NFS server and its downstream NFS clients.

This solution is based on existing Linux kernel modules, including `nfs-kernel-server` which supports NFS re-exporting and `cachefilesd` (FS-Cache) which provides a persistent cache of network filesystems on disk.

The solution works by mounting NFS exports from a source NFS filer (typically located on-premise) and re-exporting the mount points to downstream NFS clients (typically in Google Cloud).

Performing this re-export provides two layers of caching:

* **Level 1:** The standard block cache of the operating system, residing in RAM.
* **Level 2:** FS-Cache. A Linux kernel module which caches data from network filesystems locally on disk. When the volume of data exceeds available RAM (L1), the data is cached on the disk by FS-Cache. In this deployment, we use Local SSD's for the L2 cache, although this could be configured in a number of ways.

Using the deployment scripts in this repository, we further extend this architecture by creating multiple NFS proxies in a [Managed Instance Group](https://cloud.google.com/compute/docs/instance-groups) and utilising a [Internal TCP Load Balancer](https://cloud.google.com/load-balancing/docs/internal) to manage connections between the NFS clients and NFS cache.

The NFS caching solution is collectively referred to as `Knfsd` in this repository.

## Building and Deploying

This repository is broken down into two key sections:

1. [Build of Knfsd Image](image/)
2. [Deployment and Operation of Knfsd Cluster on GCP](deployment/)


You should start with the [Build of Knfsd Image](image/). Once built, you can use this image and the [deployment scripts](deployment/) to deploy and operate a Knfsd Cluster on GCP.

## Testing the Image

There is a basic [suite of smoke tests](image/smoke-tests/) that can be run after building a new image. These tests check for common configuration issues such as the correct kernel version, cachefilesd is enabled and active, etc.
