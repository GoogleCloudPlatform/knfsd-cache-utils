# knfsd-metrics-agent

This collects various system metrics from the NFS proxy and writes them in a format suitable for collectd.

## Environment Variables

* `COLLECTD_HOSTNAME`   - Overrides the default system hostname.
* `COLLECTD_INTERVAL`   - Metric collection interval in seconds, defaults to 60.
* `METRICS_SOCKET_PATH` - See `-socket` option.
* `METRICS_MODE`        - See `-mode` option.
* `METRICS_ENABLE_META` - See `-enable-meta` option.

When using the systemd service, these environment variables can be provided by the `/etc/default/knfsd-metrics-agent` file.

## Options

* `-socket path` \
  Sets the path to the unix socket created by the collectd `unixsock` plugin.
  This socket will be used to write metrics instead of stdout.

* `-proc path` \
  Sets the path to the procfs filesystem (default `/proc`).

* `-mode mode` \
  Sets which metrics are collected, valid options are proxy or client (default `proxy`).

* `-enable-meta` \
  Enables including metadata with the metrics. This requires collectd 5.11.0 or greater. To disable this use `-enable-meta=false`.

## Mode

### Proxy

* Proxy to Source, per-mount:
  * Operations per second
  * RPC backlog
  * Read RTT and exe time
  * Write RTT and exe time
* Inode cache
  * Active objects
  * Object size
* Dentry cache
  * Active objects
  * Object size
* Number of client connections

### Client

* Client to Proxy, per-mount:
  * Operations per second
  * RPC backlog
  * Read RTT and exe time
  * Write RTT and exe time
* Inode cache
  * Active objects
  * Object size
* Dentry cache
  * Active objects
  * Object size
