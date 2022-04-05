# Knfsd Agent

By default, each Knfsd node will also run the [Knfsd Agent](../../image/resources/knfsd-agent/README.md). This is a small Golang application that exposes a web server with some API Methods. Currently the Knfsd Agent only supports a basic informational API method (`/api/v1.0/nodeinfo`). This method provides basic information on the Knfsd node. It is useful for determining which backend node you are connected to when connecting to the Knfsd Cluster via the Internal Load Balancer.

Over time this will API will be expanded with additional capabilities.

This agent listens on port `80` and can be disabled by setting `ENABLE_KNFSD_AGENT` to `false` in the Terraform.

For information on the API Methods, see the [Knfsd Agent README.md](../../image/resources/knfsd-agent/README.md).
