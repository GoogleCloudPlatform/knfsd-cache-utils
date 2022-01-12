# terraform-module-knfsd

This directory contains a [Terraform Module](https://www.terraform.io/docs/modules/index.html) for deploying a Knfsd cluster on Google Cloud.

**Note:** The `main` branch may be updated at any time with the latest changes which could be breaking. You should always configure your module to use a release. This can be configured in the modules Terraform Configuration block.

```
source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.4.0"
```

## Prerequisites

Before deploying a Knfsd Cluster in Google Cloud, you should make sure the following prerequisites are met.

| Prerequisite                                                                   | Details                                                                                                                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Knfsd Disk Image](https://cloud.google.com/compute/docs/images#custom_images) | You must have followed the steps in the [/image](/image) directory to build a Knfsd Image.                                                                                                                                                                                                                                                  |
| [Firewall Rules](https://cloud.google.com/vpc/docs/firewalls)                  | This module only creates the firewall rules for [healthchecks](https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges) (this can be disabled). You must make sure you implement the firewall rules required for NFS Communication between:<br /><br /> - Knfsd --> Source NFS Server<br /> - Render Clients --> Knfsd |
| [gcloud](https://cloud.google.com/sdk/install)                                 | You should have [gcloud](https://cloud.google.com/sdk/install) installed and configured                                                                                                                                                                                                                                                     |
| [Terraform](https://www.terraform.io/downloads.html)                           | You should have Terraform 0.12 or above installed.                                                                                                                                                                                                                                                                                          |

If you **do not** currently use Terraform, follow [this guide](https://learn.hashicorp.com/tutorials/terraform/install-cli) to download and install it.

## Metrics

These deployment scripts can optionally configure the exporting a range of metrics from each Knfsd Node into [Google Cloud Operations](https://cloud.google.com/products/operations) (formerly Stackdriver). These are exported via the [Stackdriver Monitoring Agent](https://cloud.google.com/monitoring/agent) which is installed as part of the [build scripts](/image).

These metrics can be enabled via the `ENABLE_STACKDRIVER_METRICS` Terraform Variable as detailed below. **If you wish to use auto-scaling then metrics must be enabled**.

The following additional prerequisites must be met if you wish to enable metrics:

| Prerequisite                                                                                                             | Details                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Metric Descriptors and Dashboard Import](metrics)                                                                       | If this is the first time you are deploying Knfsd in a Google Cloud Project you need to setup the Metric Descriptors and import the Knfsd Monitoring Dashboard. This is achieved via a standalone Terraform configuration and the process is described in the [metrics](metrics) directory.                                                                                                                                                                                                                                                                                                                                                                      |
| [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access)                               | You must have [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access) enabled on the subnet that you will be using for the Knfsd Nodes. This is required to allow connectivity to the Monitoring API for VM's without a Public IP.                                                                                                                                                                                                                                                                                                                                                                                            |
| [Service Account Permissions](https://cloud.google.com/compute/docs/access/service-accounts#service_account_permissions) | A Service Account needs to be configured for the Knfsd Nodes with the `logging-write` and `monitoring-write` scopes. This is performed automatically by the Terraform Module when you have metrics enabled. By default, the [Compute Engine Default Service Account](https://cloud.google.com/compute/docs/access/service-accounts#default_service_account) will be used. <br><br>You need to make sure the Service Account you create for Terraform has the `roles/iam.serviceAccountUser` role on the Compute Engine Default Service Account so that it can assign it to the Knfsd Nodes. This is covered in the "Generating a Service Account" section below. |

The following custom metrics are exported currently:

| Metric Name                                                    | Description                                                                                                     |
| -------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| **custom.googleapis.com/knfsd/nfs_connections**                | The number of NFS Clients connected to the Knfsd filer (used for autoscaling).                                  |
| **custom.googleapis.com/knfsd/nfs_inode_cache_active_objects** | The number of active objects in the Linux NFS inode Cache.                                                      |
| **custom.googleapis.com/knfsd/dentry_cache_active_objects**    | The number of active objects in the Linux Dentry Cache.                                                         |
| **custom.googleapis.com/knfsd/nfs_inode_cache_objsize**        | The total size of the objects in the Linux NFS inode Cache in bytes.                                            |
| **custom.googleapis.com/knfsd/dentry_cache_objsize**           | The total size of the objects in the Linux Dentry Cache in bytes.                                               |
| **custom.googleapis.com/knfsd/nfsiostat_mount_read_exe**       | The average read operation EXE per NFS client mount over the past 60 seconds (Knfsd --> Source Filer).          |
| **custom.googleapis.com/knfsd/nfsiostat_mount_read_rtt**       | The average read operation RTT per NFS client mount over the past 60 seconds (Knfsd --> Source Filer).          |
| **custom.googleapis.com/knfsd/nfsiostat_mount_write_exe**      | The average write operation EXE per NFS client mount over the past 60 seconds (Knfsd --> Source Filer).         |
| **custom.googleapis.com/knfsd/nfsiostat_mount_write_rtt**      | The average write operation RTT per NFS client mount over the past 60 seconds (Knfsd --> Source Filer)..        |
| **custom.googleapis.com/knfsd/nfsiostat_ops_per_second**       | The number of NFS operations per second per NFS client mount over the past 60 seconds (Knfsd --> Source Filer). |
| **custom.googleapis.com/knfsd/nfsiostat_rpc_backlog**          | The RPC Backlog per NFS client mount over the past 60 seconds (Knfsd --> Source Filer).                         |

### Dashboards

The Knfsd Monitoring Dashboard is created automatically by the metrics initialisation Terraform that is detailed in the [Metrics Prerequisites](#metrics).

Once ran, you can then access the dashboard from [https://console.cloud.google.com/monitoring/dashboards/](https://console.cloud.google.com/monitoring/dashboards/)

## Autoscaling

**Important: To use Autoscaling you MUST have metrics enabled as they are used as a scaling metric**

The Knfsd Deployment Scripts also support configuring AutoScaling for the Knfsd Managed Instance Group. Scaling based on the standard metric of CPU Usage is not optimal for the caching use case. Instead the custom metric `custom.googleapis.com/knfsd/nfs_connections` is used for triggering an autoscaling event.

Autoscaling can be enabled by setting the `ENABLE_KNFSD_AUTOSCALING` environment variable to `true` (defaults to `false`). There are also some other configuration options detailed in the [Configuration Variables](#configuration-variables) section below such as how many NFS Connections a Knfsd node should be handling before a scale-up.

To avoid interruptions to existing NFS client mounts, and by extension render operations the autoscaler behaviour is set to **SCALE UP ONLY**. When a Knfsd Client exceeds the number of connections defined in the `KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD` variable a new instance will be added to the Knfsd Managed Instance group. If the number of Knfsd Connections subsequently falls significantly the Knfsd cluster **will not** automatically scale down. This is not a GCP limitation but an intentional design consideration to avoid:

- Loss of FS-Cache data that would need to be re-pulled on scale up
- Interruption to existing NFS Client Connections

You can change this behaviour if you wish, but it is not recommended.

There is a slight delay for metric ingestion (1-2 mins) and then for a new node to spin up and initialise (~2 mins). When a scaling event occurs new traffic will continue to be sent to the existing healthy nodes in the cluster until there is a new node ready to handle the connections. It is therefore recommended that you set your `KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD` slightly lower than the maximum number of connections a single Knfsd node can handle. This will start the scaling event early and make sure a new node is ready before your existing nodes become overloaded.

## Knfsd Agent

By default, each Knfsd node will also run the [Knfsd Agent](../../image/knfsd-agent/README.md). This is a small Golang application that exposes a web server with some API Methods. Currently the Knfsd Agent only supports a basic informational API method (`/api/v1.0/nodeinfo`). This method provides basic information on the Knfsd node. It is useful for determining which backend node you are connected to when connecting to the Knfsd Cluster via the Internal Load Balancer.

Over time this will API will be expanded with additional capabilities.

This agent listens on port `80` and can be disabled by setting `ENABLE_KNFSD_AGENT` to `false` in the Terraform.

For information on the API Methods, see the [Knfsd Agent README.md](../../image/knfsd-agent/README.md).

## Usage

**Note: See the [Configuration Variables](#Configuration-Variables) section for advance configuration options**

Basic usage of this module is as follows:

```terraform
module "nfs_proxy" {
    source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.4.0"

    # Google Cloud Project Configuration
    PROJECT                        = "my-gcp-project"
    REGION                         = "us-west1"
    ZONE                           = "us-west1-a"

    # Network Configuration
    NETWORK                        = "my-vpc"
    SUBNETWORK                     = "my-subnet"
    AUTO_CREATE_FIREWALL_RULES     = false
    LOADBALANCER_IP                = "10.67.4.5"

    # Knfsd Proxy Configuration
    PROXY_IMAGENAME                = "knfsd-base-image"
    EXPORT_MAP                     = "10.0.5.5;/remoteexport;/remoteexport"
    PROXY_BASENAME                 = "rendercluster1"
    KNFSD_NODES                    = 3
}

// Prints the IP address of the Load Balancer
output "load_balancer_ip_address" {
    value = module.nfs_proxy.nfsproxy_loadbalancer_ipaddress
}

// Prints the DNS address of the Load Balancer
output "load_balancer_dns_address" {
    value = module.nfs_proxy.nfsproxy_loadbalancer_dnsaddress
}
```

Edit the above [configuration variables](#Configuration-Variables) to match your desired configuration.

### Provider Default Values

The Terraform module also supports supplying the project, region and zone using provider default values. Set the project, region, and/or zone properties on the Google Terraform provider. Omit these properties from the module.

```terraform
provider "google" {
  project     = "my-gcp-project
  region      = "us-west1"
  zone        = "us-west1-a
}

module "nfs_proxy" {
    source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.4.0"

    # Network Configuration
    NETWORK                        = "my-vpc"
    SUBNETWORK                     = "my-subnet"
    AUTO_CREATE_FIREWALL_RULES     = false
    LOADBALANCER_IP                = "10.67.4.5"

    # Knfsd Proxy Configuration
    PROXY_IMAGENAME                = "knfsd-base-image"
    EXPORT_MAP                     = "10.0.5.5;/remoteexport;/remoteexport"
    PROXY_BASENAME                 = "rendercluster1"
    KNFSD_NODES                    = 3
}
```

### Generate Service Account

**If you are using [Google Cloud Shell](https://cloud.google.com/shell), or a local workstation with [gcloud](https://cloud.google.com/sdk/gcloud) installed and authenticated then you can skip this step and instead comment out the `credentials` field in the Terraform Google Provider to make Terraform use your credentials. Your account will need a minimum of the permissions described below**

You will need to generate a [Service Account](https://cloud.google.com/iam/docs/service-accounts) that Terraform can use to deploy resources in your Google Cloud Project.

This can be achieved with the following commands (these should be run in the same directory as the `deploy.tf` file):

```
GOOGLE_PROJECT=<YOUR PROJECT ID>
gcloud config set project $GOOGLE_PROJECT
PROJECT_NUMBER=$(gcloud projects describe $GOOGLE_PROJECT --format 'value(projectNumber)')
gcloud iam service-accounts create terraform-deployment-sa --description="Service Account for Terraform Knfsd Deployment"
gcloud projects add-iam-policy-binding $GOOGLE_PROJECT --member serviceAccount:terraform-deployment-sa@$GOOGLE_PROJECT.iam.gserviceaccount.com --role='roles/compute.admin'
gcloud iam service-accounts keys create service-account-key.json --iam-account terraform-deployment-sa@$GOOGLE_PROJECT.iam.gserviceaccount.com
```

If you are using metrics you should also run the below additional commands:

```
gcloud iam service-accounts add-iam-policy-binding $PROJECT_NUMBER-compute@developer.gserviceaccount.com  --member="serviceAccount:terraform-deployment-sa@$GOOGLE_PROJECT.iam.gserviceaccount.com" --role='roles/iam.serviceAccountUser'
gcloud projects add-iam-policy-binding $GOOGLE_PROJECT --member serviceAccount:terraform-deployment-sa@$GOOGLE_PROJECT.iam.gserviceaccount.com --role='roles/monitoring.admin'
```

Create or update the Google provider in Terraform to use the service account key file:

```terraform
provider "google" {
  credentials = file("service-account-key.json")
}
```

**Warning: The service-account-key.json should be stored securely - do not commit to Git or share with any unauthorised party**

### Deploy Knfsd

Once you have created your `deploy.tf` and created your Service Account you can deploy Knfsd with:

```
terraform init
terraform apply
```

## Configuration Variables

### Google Cloud Project Configuration

| Variable | Description                                                                                                                                                                     | Required | Default |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------- |
| PROJECT  | The Google Cloud Project that the Knfsd Cluster is being deployed to. If it is not provided, the provider project is used.                                                      | False    | N/A     |
| REGION   | The [Google Cloud Region](https://cloud.google.com/compute/docs/regions-zones) to use for deployment of regional resources. If it is not provided, the provider region is used. | False    | N/A     |
| ZONE     | The [Google Cloud Zone](https://cloud.google.com/compute/docs/regions-zones) to use for deployment of zonal resources. If it is not provided, the provider zone is used.        | False    | N/A     |

### Network Configuration

| Variable                   | Description                                                                                                                                                       | Required | Default   |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | --------- |
| NETWORK                    | The network name (VPC) to use for the deployment of the Knfsd Compute Engine Instances.                                                                           | False    | `default` |
| SUBNETWORK                 | The subnetwork name (subnet) to use for the deployment of the Knfsd Compute Engine Instances.                                                                     | False    | `default` |
| AUTO_CREATE_FIREWALL_RULES | Should firewall rules automatically be created to allow [healthcheck connectivity](https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges)? | False    | `true`    |
| LOADBALANCER_IP            | The IP address to use for the Internal Load Balancer. If not specified, a random IP address will be assigned within the subnet.                                   | False    | null      |

### Knfsd Proxy Configuration

| Variable                        | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Required                                                                             | Default         |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------| --------------- |
| PROXY_BASENAME                  | A nickname to use for this Knfsd deployment (used to ensure uniquely named resources for multiple deployments).                                                                                                                                                                                                                                                                                                                                                 | False                                                                                | `nfsproxy`      |
| EXPORT_MAP                      | A list of NFS Exports to mount from on-premise and re-export in the format `<SOURCE_IP>;<SOURCE_EXPORT>;<TARGET_EXPORT>`.<br><br> For example to mount `10.0.0.1/export` from on-premise and re-export as `10.100.100.1/reexport` you would set the `EXPORT_MAP` variable to `10.100.100.1;/export;/reexport`.<br><br>You can specify multiple re-exports using a comma, for example `10.100.100.1;/assets;/assetscache,10.100.100.1;/textures;/texturescache`. | `EXPORT_MAP`, `EXPORT_HOST_AUTO_DETECT` or NetApp Auto-Discovery must be configured. | N/A             |
| EXPORT_HOST_AUTO_DETECT         | A list of IP addresses or hostnames of NFS Filers that respond to the `showmount` command. Knfsd will automatically detect and re-export mounts from this filer. Exports paths on the cache will match the export path on the source filer.<br><br> You can specify multiple filers using a comma, for example `10.100.100.1,10.100.200.1`.                                                                                                                     | `EXPORT_MAP`, `EXPORT_HOST_AUTO_DETECT` or NetApp Auto-Discovery must be configured. | N/A             |
| EXCLUDED_EXPORTS                | A list of export paths excluded from auto-discovery (comma separated). Auto-discovery will ignore any mounts using these paths. Can be used to ignore protected paths such as `/bin`. Does not apply to mounts specified in the `EXPORT_MAP`. Paths excluded from auto-discovery can be explicitly exported using `EXPORT_MAP`, this can be used to change the export path.                                                                                     | `false`                                                                              | `""`            |
| EXPORT_CIDR                     | The CIDR to use in `/etc/exports` of the Knfsd Node for filesystem re-export.                                                                                                                                                                                                                                                                                                                                                                                   | False                                                                                | `10.0.0.0/8`    |
| PROXY_IMAGENAME                 | The name of the Knfsd base [image](https://cloud.google.com/compute/docs/images).                                                                                                                                                                                                                                                                                                                                                                               | True                                                                                 | N/A             |
| KNFSD_NODES                     | The number of Knfsd nodes to deploy as part of the cluster.                                                                                                                                                                                                                                                                                                                                                                                                     | False                                                                                | 3               |
| SERVICE_LABEL                   | The Service Label to use for the Forwarding Rule.                                                                                                                                                                                                                                                                                                                                                                                                               | False                                                                                | `dns`           |
| VFS_CACHE_PRESSURE              | The value to set for `vfs_cache_pressure` Rule.                                                                                                                                                                                                                                                                                                                                                                                                                 | False                                                                                | `100`           |
| NUM_NFS_THREADS                 | The number of NFS Threads to use for KNFSD.                                                                                                                                                                                                                                                                                                                                                                                                                     | False                                                                                | `512`           |
| DISABLED_NFS_VERSIONS           | The versions of NFS that should be disabled in `nfs-kernel-server`. Explicitly disabling unwanted NFS versions prevents clients from accidentally auto-negotiating an undesired NFS version. Specify multiple versions to disable with a comma separated list. Acceptable values are `3`, `4`, `4.0`, `4.1`, `4.2`. NFS Version 2 is always diabled.                                                                                                            | False                                                                                | `4.0,4.1,4.2`   |
| READ_AHEAD                      | The number of bytes to read ahead. Must be a multiple of the kernel page size (8 KiB for 5.11). The kernel will round this down to the nearest page.                                                                                                                                                                                                                                                                                                            | False                                                                                | `8388608`       |
| ENABLE_UDP                      | Create a load balancer to support UDP traffic to the NFS proxy instances. UDP is not recommended for the main NFS traffic as it can cause data corruption. However, this maybe useful for older clients that default to using UDP for the mount protocol.                                                                                                                                                                                                       | False                                                                                | `false`         |
| ENABLE_AUTOHEALING_HEALTHCHECKS | Should failed healthchecks lead to instance replacement?                                                                                                                                                                                                                                                                                                                                                                                                        | False                                                                                | `true`          |
| ENABLE_STACKDRIVER_METRICS      | Should Knfsd metrics be exported into Stackdriver?                                                                                                                                                                                                                                                                                                                                                                                                              | False                                                                                | `true`          |
| CUSTOM_PRE_STARTUP_SCRIPT       | Optional bash script to run before the [proxy-startup.sh](proxy-startup.sh) script. For example `file("/home/ben/myscript.sh")`.                                                                                                                                                                                                                                                                                                                                | `false`                                                                              | empty script    |
| CUSTOM_POST_STARTUP_SCRIPT      | The path to a bash script to run after the [proxy-startup.sh](proxy-startup.sh) script. For example `file("/home/ben/myscript.sh")`.                                                                                                                                                                                                                                                                                                                            | `false`                                                                              | empty script    |
| LOCAL_SSDS                      | The number of Local SSD's to assign to each cache instance. This can be either 1 to 8, 16, or 24 local SSDs for up to 9TB of capacity ([see here](https://cloud.google.com/compute/docs/disks/local-ssd#choosing_a_valid_number_of_local_ssds)). If you are setting this to 24 Local SSD's you should also change the `MACHINE_TYPE` variable to an instance with 32 CPU's, for example `n1-highmem-32`                                                         | `false`                                                                              | `4`             |
| MACHINE_TYPE                    | The GCP Machine type to use for the Knfsd cache. Currently only N1 instances can be used.                                                                                                                                                                                                                                                                                                                                                                       | `false`                                                                              | `n1-highmem-16` |
| MIG_MAX_UNAVAILABLE_PERCENT     | The maximum number of instances that can be unavailable during automated MIG updates ([see docs](https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#max_unavailable)). Defaults to 100% to ensure consistent cache instances within the MIG.                                                                                                                                                                  | `false`                                                                              | `100`           |
| MIG_REPLACEMENT_METHOD          | The instance replacement method for managed instance groups. Valid values are: `RECREATE`, `SUBSTITUTE`.<br><br>If `SUBSTITUTE` (default), the group replaces VM instances with new instances that have randomly generated names. If `RECREATE`, instance names are preserved. You must also set `MIG_MAX_UNAVAILABLE_PERCENT` to be greater than 0 (default is already `100` so this only applies if you have modified this variable).                         | `false`                                                                              | `SUBSTITUTE`    |
| MIG_MINIMAL_ACTION              | Minimal action to be taken on an instance. You can specify either RESTART to restart existing instances or REPLACE to delete and create new instances from the target template. If you specify a RESTART, the Updater will attempt to perform that action only. However, if the Updater determines that the minimal action you specify is not enough to perform the update, it might perform a more disruptive action.                                          | `false`                                                                              | `RESTART`       |
| ENABLE_KNFSD_AGENT              | Should the [Knfsd Agent](../../image/knfsd-agent/README.md) be started at Proxy Startup?                                                                                                                                                                                                                                                                                                                                                                        | `false`                                                                              | `true`          |

### NetApp Exports Auto-Discovery

| Variable                  | Description                                                                                                                                                                                                                                                                           | Required                                 | Default                               |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| ---------------------------------------- | ------------------------------------- |
| ENABLE_NETAPP_AUTO_DETECT | Enables automatic discovery of exports using the NetApp REST API.                                                                                                                                                                                                                     | False                                    | `false`                               |
| NETAPP_HOST               | DNS or IP of the NetApp server. This is the DNS or IP name clients use when mounting the NFS shares.                                                                                                                                                                                  | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_URL                | URL of the NetApp REST API. This *must* include the API version and end with a slash, for example `https://netapp.example/api/v1/`.                                                                                                                                                   | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_USER               | The username used to authenticate with the NetApp REST API.                                                                                                                                                                                                                           | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_SECRET             | The name of a GCP Secret containing the NetApp REST API password.                                                                                                                                                                                                                     | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_SECRET_PROJECT     | The GCP project containing the secret.                                                                                                                                                                                                                                                | False                                    | The project the cluster is running in |
| NETAPP_SECRET_VERSION     | The version of the secret.                                                                                                                                                                                                                                                            | False                                    | `latest`                              |
| NETAPP_CA                 | PEM encoded certificate containing the root certificate for the NetApp REST API. This can also include intermediate certificates to provide the full certificate chain. To read this from a file use the [Terraform file function](https://www.terraform.io/language/functions/file). | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_ALLOW_COMMON_NAME  | Allows using the Common Name (CN) field of the certificate as a DNS name when the certificate does not include a Subject Alternate Name (SAN) field.                                                                                                                                  | False                                    | `false`                               |

For instructions on how to get the NetApp root CA certificate, and verify the `netapp-exports` command works see the [netapp-exports README](../image/resources/netapp-exports/README.md).

#### NetApp Self-Signed Certificates

Modern SSL certificates use the Subject Alternate Name (SAN) field to provide a list of DNS names and IPs that are valid for the certificate.

However, older certificates relied on the Common Name (CN) field. This use has been deprecated and is no longer supported by default as the Common Name field was ambiguous.

If you have a certificate that does not contain a Subject Alternate Name then you can set `NETAPP_ALLOW_COMMON_NAME=true`. When this is enabled the Common Name *must* be the DNS name or IP address of the NetApp cluster. This DNS name or IP address *must* be used for the `NETAPP_URL` host.

If the certificate contains a Subject Alternate Name then the Common Name will be ignored.

#### Updating NetApp secret

Normally using the `latest` version for secrets in Terraform is discouraged because Terraform will not detect when a new version is added to the secret. However, in this case using `latest` does not cause any issues because the secret is only used when a proxy instance is starting up.

To update the NetApp secret, just add a new version and disable the old version. Once the new version has been verified as valid the old version can be destroyed.

Changing the password and updating the secret will not affect any running instances as the password is only required to generate the list of exports when the instance starts.


### Mount Options

These mount options are for the proxy to the source server.

| Variable       | Description                                                                                                                                                                                                                            | Required | Default   |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | --------- |
| NCONNECT_VALUE | The number of TCP connections to use when connecting to the source.                                                                                                                                                                    | False    | `16`      |
| ACREGMIN       | The minimum time (in seconds) that the NFS client caches attributes of a regular file.                                                                                                                                                 | False    | `600`     |
| ACREGMAX       | The maximum time (in seconds) that the NFS client caches attributes of a regular file.                                                                                                                                                 | False    | `600`     |
| ACDIRMIN       | The minimum time (in seconds) that the NFS client caches attributes of a directory.                                                                                                                                                    | False    | `600`     |
| ACDIRMAX       | The maximum time (in seconds) that the NFS client caches attributes of a directory. This can be reduced to improve the cache coherency for `readdir` operationns (e.g `ls`) at the cost of increasing metadata requests to the source. | False    | `600`     |
| RSIZE          | The maximum number of bytes the proxy will read from the source in a single request. The actual value will be negotiated with the source server to determine the maximum value support by both machines.                               | False    | `1048576` |
| WSIZE          | The maximum number of bytes the proxy will write to the source in a single request. The actual value will be negotiated with the source server to determine the maximum value support by both machines.                                | False    | `1048576` |
| MOUNT_OPTIONS  | Any additional NFS mount options not covered by existing variables. These options will be applied to all NFS mounts.                                                                                                                   | False    | `""`      |
| EXPORT_OPTIONS | Any custom NFS exports options. These options will be applied to all NFS exports.                                                                                                                                                      | False    | `""`      |

### Autoscaling Configuration

| Variable                                    | Description                                                                                                                            | Required | Default |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------- |
| ENABLE_KNFSD_AUTOSCALING                    | Should autoscaling be enabled for Knfsd? **You MUST set the `ENABLE_STACKDRIVER_METRICS` variable to `true` if enabling autoscaling**. | False    | `false` |
| KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD | The number of Client Connections to Knfsd that should be targeted for each instance (exceeding will trigger a scale-up).               | False    | `250`   |
| KNFSD_AUTOSCALING_MIN_INSTANCES             | The minimum number of Knfsd instances to set regardless of the traffic volumes.                                                        | False    | `1`     |
| KNFSD_AUTOSCALING_MAX_INSTANCES             | The maximum number of Knfsd instances to set regardless of the traffic volumes.                                                        | False    | `10`    |

## Caveats

### Combining auto-discovery and explicit mounts

While auto-discovery and explicit mounts can be combined the system does not have any special handling for duplicate paths.

As such it is not recommended to combine multiple auto-discovery methods, or explicit (`EXPORT_MAP`).

The behaviour of duplicates is undefined. The system might overwrite one mount with another, or it may error.

### Limitations on export names

The proxy cannot re-export paths that match standard Linux system directories such as `/bin`, `/dev`, `/usr/local/lib`, etc.

While some NFS servers may support creating an export with at path such as `/bin`, attempting to re-export the path via the cache would result in the cache overlaying its local `/bin` directory with the mount.

Also included in the list of protected directories is the `/home` directory. This is because some commands such as `gcloud compute ssh` will look for ssh keys in the home directory, and may create user home directories and ssh keys within the home directories. To prevent this undesired, and unexpected behaviour the `/home` directory is not supported.

The proxy will fail to start if it attempts to export a protected path. Check the logs for errors such as:

> startup-script: ERROR: Cannot mount 10.0.0.2:/home because /home is a system path

If you are providing a manual export list, specify a different path for the export, such as `10.0.0.2;/home;/mnt/home`.

If you're using auto-discovery add the path to the list of excluded exports, for example `EXCLUDED_EXPORTS = "/home,/dev"`.
