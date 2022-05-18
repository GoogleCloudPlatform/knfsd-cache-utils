# Ports

Knfsd uses static ports for `nfs-kernel-server`. These ports are listed below.

You should allow inbound traffic to Knfsd from your NFS clients on the below ports.

For outbound connectivity from Knfsd to your source NFS server different ports may be used. Please consult the docs for your source NFS server for more information.

## General

* 80    - HTTP (knfsd-agent)

## NFS v3

* 111   - RPC portmapper
* 2049  - NFS
* 20048 - mountd
* 20050 - nlm
* 20051 - statd
* 20052 - lockd

## NFS v4

* 2049  - NFS
* 20053 - NFS v4 callback
