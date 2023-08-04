# Firewall configuration

## Health checks

Health checks are required when:

* `ENABLE_AUTOHEALING_HEALTHCHECKS = true`
* or `TRAFFIC_DISTRIBUTION_MODE = "loadbalancer"`

If `AUTO_CREATE_FIREWALL_RULES = true` then the firewall rule to allow health checks will be created for you.

Setting `AUTO_CREATE_FIREWALL_RULES = false` is advised for production, or when deploying multiple knfsd proxy clusters to the same GCP project as the rules only need to be defined once per VPC network.

**NOTE:** Trying to deploy multiple knfsd proxy clusters in the same project with `AUTO_CREATE_FIREWALL_RULES = true` will fail due to a duplicate health check name.

### Using the GCP Console

1. In the Google Cloud console go to the [Firewalls](https://console.cloud.google.com/networking/firewalls/list) page.

2. Click **Create firewall rule**.

3. On the *Create firewall rule* page, supply the following information:

   * **Name**: Provide a name for the health check (e.g.`allow-nfs-tcp-healthcheck`).
   * **Network**: Choose a VPC network.
   * **Priority**: Enter a number for the priority. Lower numbers have higher priorities. Be sure that the firewall rule has a higher priority than other rules that might deny ingress traffic.
   * **Targets**: Choose **Specified target tags**.
   * **Target tags**: `knfsd-cache-server`.
   * **Source filter**: Choose **IPv4 ranges**.
   * **Source IPv4 ranges**: `130.211.0.0/22, 35.191.0.0/16, 209.85.152.0/22, 209.85.204.0/22`.
   * **Protocols and ports**: Choose **Specified protocols and ports**.
     * **TCP**: `2049`.

### Using Terraform

```terraform
# Firewall rule to allow healthchecks from the GCP Healthcheck ranges
resource "google_compute_firewall" "allow-tcp-healthcheck" {
  project  = "my-gcp-project"
  name     = "allow-nfs-tcp-healthcheck"
  network  = "my-vpc"
  priority = 1000

  allow {
    protocol = "tcp"
    ports    = ["2049"]
  }
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "209.85.152.0/22", "209.85.204.0/22"]
  target_tags   = ["knfsd-cache-server"]
}
```

## Knfsd proxy instances to source servers

By default, no firewall rules are required in GCP to allow egress traffic. However, if you have created a firewall rule restricting egress traffic then you will need to create a firewall rule (with a higher priority) to allow egress traffic from the knfsd proxy instances to the the source servers.

Many of the ports used by NFS v3 are dynamic, so you will need to check with the source NFS server to see which ports are required.

## Clients to knfsd proxy instances

To allow traffic from clients to knfsd proxy instances we need to add ingress rules in GCP.

Filtering by network tag is preferred. To support this assign all the clients that need to access the NFS proxy cluster a common network tag such as `nfs-client`.

### Using the GCP Console

Create a firewall rule with the following information:

* **Name**: Provide a name for firewall rule.
* **Network**: Choose a VPC network.

* **Priority**: Enter a number for the priority. Lower numbers have higher priorities. Be sure that the firewall rule has a higher priority than other rules that might deny ingress traffic.

* Specify the **Source filter**:

  * To filter incoming traffic by network tag, choose Source tags, and then type the network tags into the Source tags field.

  * To filter incoming traffic by source IPv4 ranges, select IPv4 ranges, and then enter the CIDR blocks into the Source IPv4 ranges field. Use 0.0.0.0/0 for any IPv4 source.

* **Targets**: Choose **Specified target tags**.
* **Target tags**: `knfsd-cache-server`.
* **Protocols and ports**: Choose **Specified protocols and ports** and
  * **TCP**: `111, 2049, 20048, 20050, 20051, 20052, 20053`.

### Using Terraform

```terraform
# Firewall rule to allow client to knfsd proxy
resource "google_compute_firewall" "allow-nfs" {
  project  = "my-gcp-project"
  name     = "allow-nfs"
  network  = "my-vpc"
  priority = 1000

  allow {
    protocol = "tcp"
    ports    = ["111", "2049", "20048", "20050", "20051", "20052", "20053"]
  }

  # # Choose either source_tags or source_ranges
  # source_tags   = ["nfs-client"]
  # source_ranges = ["10.0.0.0/8"]

  # The knfsd proxy always has the network tag knfsd-cache-server
  target_tags   = ["knfsd-cache-server"]
}
```
