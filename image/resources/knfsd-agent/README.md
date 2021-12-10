# knfsd-agent

By default, each Knfsd node will also run the [Knfsd Agent](../../image/knfsd-agent/README.md).

This is a small Golang application that exposes a web server with some API Methods. Currently the Knfsd Agent only supports a basic informational API method (`/api/v1.0/nodeinfo`). This method provides basic information on the Knfsd node. It is useful for determining which backend node you are connected to when connecting to the Knfsd Cluster via the Internal Load Balancer.

Over time this will API will be expanded with additional capabilities.

This agent listens on port `80` and can be disabled by setting `ENABLE_KNFSD_AGENT` to `false` in the Terraform.

## Methods

### GET /api/v1.0/nodeinfo

Note: This method is also accessible via `/`

This method provides basic information on the Knfsd node. It is useful for determining which backend node you are connected to when connecting via the internal load balancer.

```
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
