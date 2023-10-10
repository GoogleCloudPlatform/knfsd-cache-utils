# Knfsd Agent

By default, each knfsd proxy node will also run the [Knfsd Agent](../../image/knfsd-agent/README.md).

The Knfsd Agent provides:

* Basic information about the knfsd proxy node.
* Metrics to aid with diagnostics and testing.

The Knfsd Agent listens on port `80` and can be disabled by setting `ENABLE_KNFSD_AGENT` to `false` in the Terraform configuration.

## Methods

### GET /api/v1.0/nodeInfo

Deprecated; use `GET /api/v1/nodeInfo` instead.

### GET /api/v1/nodeInfo

Note: This method is also accessible via `/`

This method provides basic information on the knfsd proxy node. It is useful for determining which backend node you are connected to when connecting via the internal load balancer.

```json
{
  "name": "nfsproxy-hr7s",
  "hostname": "nfsproxy-hr7s.europe-west2-a.c.xxx.internal",
  "interfaceConfig": {
    "ipAddress": "10.67.4.79",
    "networkName": "knfsd-demo-vpc",
    "networkURI": "projects/123/networks/knfsd-demo-vpc"
  },
  "zone": "europe-west2-a",
  "machineType": "n1-highmem-16",
  "image": "projects/xxx/global/images/knfsd-image"
}
```

### GET /api/v1/mounts

Lists the NFS mounts on the knfsd proxy node.

```json
{
  "mounts": [
    {
      "device": "10.0.0.2:/files",
      "mount": "/srv/nfs/files",
      "export": "/files",
      "options": {
        "rw": "",
        "noatime": "",
        "vers": "3",
        "rsize": "1048576",
        "wsize": "1048576",
        ...
      }
    },
    ...
  ]
}
```

* `mounts` - List of NFS mounts on the knfsd proxy node.
  * `device` - (`string`) NFS device name (source), in the same format as per the `mount` command (`<source>:<export>`).
  * `mount` - (`string`) Directory where the NFS share is mounted on the proxy.
  * `export` - (`string`) Path that the NFS share is re-exported as.
  * `options` - (`map[string]string`) Mount options for the NFS share.
    These match the mount options, for example `vers=3` would be `"vers": "3"`.
    Not all mount options have a value, for example the options `rw,noatime` are represented as `{"rw": "", "noatime": ""}`.

    **NOTE**: These options *DO NOT* include options that have their default value.
    For example, the option `acregmax=60` is not included in the output as this is the default value.
    To see options with their default values use the `/api/v1/mountstats` instead.

### GET /api/v1/mountstats

Lists NFS per-mount metrics on the knfsd proxy node.

This endpoint is only intended for diagnostics and testing and has been optimised for ease of use. As such the response size can be quite large if you have a lot of mounted volumes. For standard reporting of metrics you should use the [knfsd-metrics-agent](../knfsd-metrics-agent/).

Unlike the knfsd-metrics-agent, the data is not cached or periodically scraped. The latest metrics are fetched every time this endpoint is queried.

```json
{
  "mounts": [
    {
      "device": "10.0.0.2:/files",
      "mount": "/srv/nfs/files",
      "export": "/files",
      "options": {
        "rw": "",
        "noatime": "",
        "vers": "3",
        "rsize": "1048576",
        "wsize": "1048576",
        ...
      },
      "metrics": {
        "age": "1h17m46s",
        "bytes": {
          "normalRead": 645920194560,
          "normalWrite": 548800561152,
          "directRead": 0,
          "directWrite": 0,
          "serverRead": 294205259776,
          "serverWrite": 548336463872,
          "readPages": 71827456,
          "writePages": 133996753
        },
        "events": {
          "inodeRevalidate": 2660,
          "dnodeRevalidate": 108546,
          "dataInvalidate": 11,
          "attributeInvalidate": 23,
          "vfsOpen": 2619,
          "vfsLookup": 3482,
          "vfsAccess": 212467,
          "vfsUpdatePage": 133984512,
          "vfsReadPage": 3,
          "vfsReadPages": 85485,
          "vfsWritePage": 0,
          "vfsWritePages": 523491,
          "vfsGetdents": 534,
          "vfsSetattr": 701,
          "vfsFlush": 2576,
          "vfsFsync": 522837,
          "vfsLock": 0,
          "vfsFileRelease": 2576,
          "congestionWait": 0,
          "truncation": 658,
          "writeExtension": 133984512,
          "sillyRename": 0,
          "shortRead": 0,
          "shortWrite": 125,
          "delay": 0,
          "pnfsRead": 0,
          "pnfsWrite": 0
        },
        "operations": [
          {
            "operation": "NULL",
            "requests": 1,
            "transmissions": 1,
            "retries": 0,
            "majorTimeouts": 0,
            "bytesSent": 44,
            "bytesReceived": 24,
            "queueMilliseconds": 0,
            "rttMilliseconds": 0,
            "executionMilliseconds": 0,
            "errors": 0
          },
          {
            "operation": "GETATTR",
            "requests": 2671,
            "transmissions": 2671,
            "retries": 0,
            "majorTimeouts": 0,
            "bytesSent": 404568,
            "bytesReceived": 299152,
            "queueMilliseconds": 10,
            "rttMilliseconds": 10695,
            "executionMilliseconds": 10749,
            "errors": 0
          },
          ...
        ]
      }
    },
    ...
  ]
}
```

All of these metrics are cumulative counters, to get calculate a rate you need to sample the data periodically and compute the difference.
You should also compute the difference in age between two samples and divide the counters by the difference to get the average rate per second.

**NOTE**: When a share is re-mounted all the counters will be reset. This can be detected by the age also resetting.
If the age of the latest sample is less than the age of the previous sample, assume the counters have been reset.

* `mounts` - List of NFS mounts on the knfsd proxy node.
  * `device` - (`string`) NFS device name (source), in the same format as per the `mount` command (`<source>:<export>`).
  * `mount` - (`string`) Directory where the NFS share is mounted on the proxy.
  * `export` - (`string`) Path that the NFS share is re-exported as.
  * `options` - (`map[string]string`) Mount options for the NFS share.
    These match the mount options, for example `vers=3` would be `"vers": "3"`.
    Not all mount options have a value, for example the options `rw,noatime` are represented as `{"rw": "", "noatime": ""}`.

  * `metrics` - Cumulative NFS counters for the mount.
    * `age` - ([`Duration`](https://pkg.go.dev/time#Duration)) How long the share has been mounted. This can be used to detect counter resets due to the share being re-mounted.

    * `bytes` - NFS Byte counters.

      The total number of bytes read/wrote is obtained by adding normal + direct read/write counters.

      Comparing the number of bytes read (`normalRead`) to the number of bytes read from the server (`serverRead`) gives an indication of how well the cache is performing (both in-memory and on-disk).

      Direct reads will bypass the cache, however there isn't a mechanism for an external client to issue a direct read via the proxy.
      A client opening a file with `O_DIRECT` will only bypass their local cache and issue the read directly to the proxy.

      All writes are passed through the proxy to the source server.

      * `normalRead` - (uint64) Number of bytes read by applications (this includes data served by the cache).
      * `normalWrite` - (uint64) Number of bytes wrote by applications.
      * `directRead` - (uint64) Number of bytes read with the `O_DIRECT` flag.
      * `directWrite` - (uint64) Number of bytes wrote with the `O_DIRECT` flag.
      * `serverRead` - (uint64) Number of bytes read from the source server.
      * `serverWrite` - (uint64) Number of bytes wrote to the source server.
      * `readPages` - (uint64) Number of completed page reads.
      * `writePages` - (uint64) Number of completed page writes.

    * `events` - Event counters.

      The event counters give low-level event counters for various NFS operations to allow monitoring without needing to enable NFS trace debugging.

      The VFS event counters provide metrics on how often various [VFS functions](https://docs.kernel.org/filesystems/vfs.html) are invoked. Due to FS-Cache, not every VFS call results in a call to the source server.

      * `inodeRevalidate` - (uint64) Number of times cached inode attributes have to be re-validate from the source server.

      * `dnodeRevalidate` - (uint64) Number of times cache dnode entries have to be re-validated from the source server.

      * `dataInvalidate` - (uint64) Number of times a cached inode had its cached data thrown out.

      * `attributeInvalidate` - (uint64) Number of times an inode had its cached inode attributes invalidated.

      * `vfsOpen` - (uint64) Number of file or directories that have been opened.

      * `vfsLookup` - (uint64) Number of name lookups in directories.

      * `vfsAccess` - (uint64) Number of times permissions have been read.

      * `vfsUpdatePage` - (uint64) Number of pages updated (and potentially written).

      * `vfsReadPage` - (uint64) Number of single-page reads. These are less frequent as NFS prefers to read multiple pages at once using read-ahead.

      * `vfsReadPages` - (uint64) Number of multi-page reads (`nfs_readahead`). This provides an indication of the number of NFS read-ahead operations.

      * `vfsWritePage` - (uint64) Number of single-page writes.

        These are less frequent as NFS prefers to write multiple pages at once when flushing changes. Most of these calls are in response to automatically flushing dirty pages before reading or when closing a file.

        This can also be invoked via various memory-mapped file operations, though this not relevant for the knfsd proxy. Any memory-mapped files would be on a client, and these would be translated to standard NFS read/write operations by the client before being sent to the proxy.

      * `vfsWritePages` - (uint64) Number of multi-page writes. This indicates a batch write, though for smaller files/changes the batch might only include a single page.

      * `vfsGetdents` - (uint64) Number of times directory entries have been read `getdents(2)`. This is the linux VFS `readdir`, and can be translated by the NFS client to either an NFS `READDIR` or `READDIRPLUS` operation.

      * `vfsSetattr` - (uint64) Number of times the attributes were updated.
        This is invoked by `chmod(2)` and related system calls.

      * `vfsFlush` - (uint64) Number of times pending writes were flushed to the server.

      * `vfsFsync` - (uint64) Number of times `fsync(2)` was called on a file or directory. On NFS flushing a directory has no effect as all directory operations are synchronous. Unfortunately there's no way to tell the two apart.

      * `vfsLock` - (uint64) Number of attempts to lock a file (or parts of a file).

      * `vfsFileRelease` - (uint64) Number of times a file was released.
        A file is released when the last open handle to the file is closed.

      * `truncation` - (uint64) Number of times a file was truncated (e.g. by using `truncate(2)`).

        This is the linux VFS concept of truncating, and as per the docs can also extend the file. Resize would have been a better name, but in all the documentation this metric is referred to as truncation or `SETATTRTRUNC`.

      * `writeExtension` - (uint64) Number of times a file's size was increased due to writing data beyond the end of the file.

      * `sillyRename` - (uint64) Number of times a file was deleted while a process still has the file open.

        This operation is permitted by linux (the file isn't actually deleted until the last file handle is closed) and used by many linux applications. To avoid this if a file is still open the NFS client renames the file to `*.nfsXXXXXX` and then deletes it later.

        Because this is handled on client this event will not occur on the proxy. The proxy will just see a standard file rename by the client.

      * `shortRead` - (uint64) The NFS server returned less data than requested.

      * `shortWrite` - (uint64) The NFS server wrote less data than requested.

      * `delay` - (uint64) The NFS server requires longer than usual to handle the request, the client should re-try the request.

        In the past this was used if the data had to be retrieved from archival storage such as from a tape drive or CD-ROM jukebox. Some file systems still use the `-EJUKEBOX` error code to trigger a client to re-try an operation. For example `cephfs` uses `-EJUKEBOX` in response to the metadata server's state changing.

      * `pnfsRead` - (uint64) Number of parallel NFS reads.

      * `pnfsWrite` - (uint64) Number of parallel NFS writes.

    * `operations` - List of NFS operations.

      ```text
      |<------------------------ Execution Time -------------------------->|
      |<-- Queue Time -->|<------- Round-Trip Time (RTT) ----->|           |
      |------------------|=====================================|-----------|
      |                  \ Request Sent         Reply Received /           |
      \ Request Created                                      Reply Handled /
      ```

      * `operation` - (string) Name of the operation.

      * `requests` - (uint64) Number of RPC requests.

      * `transmissions` - (uint64) Number of times RPC requests were sent.

        An RPC request may be retried if there is an error or timeout with the
        request. Thus the number of transmissions will be equal to or greater than the number of requests.

      * `retries` - (uint64) Number of retries (`transmissions - requests`).

      * `majorTimeouts` - (uint64) Number of major timeouts.

      * `bytesSent` - (uint64) Total number of bytes sent to the source server for this operation. This includes both the RPC header and RPC payload. This closely matches the on-the-wire size.

      * `bytesReceived` - (uint64) Total number of bytes received from the source server for this operation. This includes both the RPC header and RPC payload. This closely matches the on-the-wire size.

      * `queueMilliseconds` - (uint64) Total time requests were queued for before being sent to the source server.

      * `rttMilliseconds` - (uint64) Total time taken to receive a reply after sending a request. This measures the round-trip time (rtt) to the source server and includes network latency, and processing time of the source server.

      * `executionMilliseconds` - (uint64) Total time taken from when the request was created, until the request has been processed. This includes any queuing time and the time taken by the client to process the response.

      * `errors` - (uint64) Total number of requests that returned an error. This includes errors such as `ESTALE`, `ETIMEOUT`, `ENOTSUPP`, and `EAGAIN`.

### GET /api/v1/nfs/client

```json
{
  "io": {
    "read": 294205259776,
    "write": 548336463872
  },
  "net": {
    "totalPackets": 0,
    "udpPackets": 0,
    "tcpPackets": 0,
    "tcpConnections": 0
  },
  "rpc": {
    "count": 406328,
    "authRefreshes": 16,
    "retransmissions": 406492
  },
  "proc3": {
    "NULL": 1,
    "GETATTR": 12118,
    "SETATTR": 5,
    "LOOKUP": 37474,
    "ACCESS": 1334,
    ...
  },
  "proc4": {
    "NULL": 1,
    "READ": 246913,
    "WRITE": 42310,
    "COMMIT": 77,
    "OPEN": 3,
    ...
  }
}
```

* `io`
  * `read` - (uint64) Bytes returned by read requests. This is the total number of bytes read from the source server by all NFS requests.

  * `write` - (uint64) Bytes sent by write requests. This is the total number of bytes wrote to the source server by all NFS requests.

* `net` - Network counters. These only show received (incoming) connections.
For the NFS client this means NFS v4 callbacks. These are normally zero for the proxy.

  * `totalPackets` - (uint64) Total number of packets received.
  * `udpPackets` -  (uint64) Total number of UDP packets received.
  * `tcpPackets` - (uint64) Total number of TCP packets received.
  * `tcpConnections` - (uint64) Total number of TCP connections received.

* `rpc` - RPC counters. These show the number of NFS RPC requests sent by the proxy to the source server.

  * `count` - (uint64) Total number of RPC requests sent.
  * `authRefreshes` - (uint64) Total number of times credentials were bound or refreshed.
  * `retransmissions` - (uint64) Total number of times an RPC request was re-transmitted (retried).

* `proc3` - (map[string]uint64) Total number of each NFS v3 operation sent.

* `proc4` - (map[string]uint64) Total number of each NFS v4 operation sent.

### GET /api/v1/nfs/server

```json
{
  "threads": 512,
  "io": {
    "read": 645920194560,
    "write": 549282906112
  },
  "net": {
    "totalPackets": 2224734,
    "udpPackets": 0,
    "tcpPackets": 2224447,
    "tcpConnections": 281
  },
  "rpc": {
    "count": 2224454,
    "badTotal": 84,
    "badFormat": 84,
    "badAuth": 0
  },
  "proc3": {
    "NULL": 119,
    "GETATTR": 68418,
    "SETATTR": 681,
    "LOOKUP": 54223,
    "ACCESS": 6491,
    ...
  },
  "proc4": {
    "NULL": 1,
    "COMPOUND": 16487
  },
  "proc4ops": {
    "ACCESS": 5673,
    "CLOSE": 2040,
    "COMMIT": 0,
    "CREATE": 1427,
    "DELEGPURGE": 0,
    ...
  }
}
```

* `threads`- (uint64) Number of NFS server threads.
* `io`
  * `read` - (uint64) Bytes sent by read requests. This is the total number of bytes read by clients from the proxy by all NFS requests.

  * `write` - (uint64) Bytes received by write requests. This is the total number of bytes wrote by clients via the proxy by all NFS requests.

* `net` - Network counters. These show the total number of incoming NFS packets from clients to the proxy.
  * `totalPackets` - (uint64) Total number of packets received.
  * `udpPackets` -  (uint64) Total number of UDP packets received.
  * `tcpPackets` - (uint64) Total number of TCP packets received.
  * `tcpConnections` - (uint64) Total number of TCP connections received.

* `rpc` - RPC counters. These show the number of NFS RPC requests received from clients by the proxy.

  * `count` - (uint64) Total number of RPC requests sent.
  * `badTotal` - (uint64) Total number of bad (error) RPC requests sent.
  * `badFormat` - (uint64) Total number of bad RPC requests that were due to a formatting error (invalid program, wrong protocol version, invalid data, etc.)
  * `badAuth` - (uint64) Total number of bad RPC requests due to an authorization error.

* `proc3` - (map[string]uint64) Total number of each NFS v3 procedure received.

* `proc4` - (map[string]uint64) Total number of each NFS v4 procedure received.

  For NFS v4 there are only two procedures, `NULL` and `COMPOUND`. The `COMPOUND` procedure can contain one or more NFS v4 operations. This allows NFS v4 clients to send multiple operations in a single RPC call.

* `proc4ops` - (map[string]uint64) Total number of each NFS v4 operation received.

## References

* [RFC 1813 - NFS Version 3 Protocol Specification](https://www.rfc-editor.org/rfc/rfc1813.html)
* [RFC 7530 - Network File System (NFS) Version 4 Protocol](https://www.rfc-editor.org/rfc/rfc7530.html)
* [Overview of the Linux Virtual File System](https://docs.kernel.org/filesystems/vfs.html)
