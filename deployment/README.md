# terraform-module-knfsd

This directory contains a [Terraform Module](https://www.terraform.io/docs/modules/index.html) for deploying a Knfsd cluster on Google Cloud.

**Note:** The master branch may be updated at any time with the latest changes which could be breaking. You should always configure your module to use a release. This can be configured in the modules Terraform Configuration block.

```
source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.1"
```

## Prerequisites

Before deploying a Knfsd Cluster in Google Cloud, you should make sure the following prerequisites are met.

| Prerequisite                                                                   | Details                                                                                                                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Knfsd Disk Image](https://cloud.google.com/compute/docs/images#custom_images) | You must have followed the steps in the [/image](/image) directory to build a Knfsd Image.                                                                                                                                                                                                                                                  |
| [Firewall Rules](https://cloud.google.com/vpc/docs/firewalls)                  | This module only creates the firewall rules for [healthchecks](https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges) (this can be disabled). You must make sure you implement the firewall rules required for NFS Communication between:<br /><br /> - Knfsd --> Source NFS Server<br /> - Render Clients --> Knfsd |
| [gcloud](https://cloud.google.com/sdk/install)                                 | You should have [gcloud](https://cloud.google.com/sdk/install) installed and configured                                                                                                                                                                                                                                                     |
| [Terraform](https://www.terraform.io/downloads.html)                           | You should have Terraform 0.12 or above installed.                                                                                                                                                                                                                                                                                          |

## Metrics

These deployment scripts can optionally configure the exporting a range of metrics from each Knfsd Node into [Google Cloud Operations](https://cloud.google.com/products/operations) (formerly Stackdriver). These are exported via the [Stackdriver Monitoring Agent](https://cloud.google.com/monitoring/agent) which is installed as part of the [build scripts](/image).

These metrics can be enabled via the `ENABLE_STACKDRIVER_METRICS` Terraform Variable as detailed below. **If you wish to use auto-scaling then metrics must be enabled**.

The following additional prerequisites must be met if you wish to enable metrics:

| Prerequisite                                                                                                             | Details                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Metric Descriptors and Dashboard Import](metrics)                                                                                  | If this is the first time you are deploying Knfsd in a Google Cloud Project you need to setup the Metric Descriptors and import the Knfsd Monitoring Dashboard. This is achieved via a standalone Terraform configuration and the process is described in the [metrics](metrics) directory.                                                                                                                                                                                                                                                                                                                                                                      |
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

## Usage

**Note: See the [Configuration Variables](#Configuration-Variables) section for advance configuration options**

### Use with Existing Terraform Configuration

If you have an existing Terraform Environment - you can simply add a configuration block for the Terraform Module - and edit the variables as required.

```hcl
module "terraform-module-knfsd" {

    source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.1"

    # Google Cloud Project Configuration
    PROJECT                                     = "my-gcp-project"
    REGION                                      = "us-west1"
    ZONE                                        = "us-west1-a"

    # Network Configuration
    NETWORK                                     = "my-vpc"
    SUBNETWORK                                  = "my-subnet"
    AUTO_CREATE_FIREWALL_RULES                  = false
    LOADBALANCER_IP                             = "10.67.4.5"

    # Knfsd Proxy Configuration
    PROXY_IMAGENAME                             = "knfsd-base-image"
    EXPORT_MAP                                  = "10.0.5.5;/remoteexport;/remoteexport"
    PROXY_BASENAME                              = "rendercluster1"
    KNFSD_NODES                                 = 3

}

// Prints the IP address of the Load Balancer
output "load_balancer_ip_address" {
    value = module.terraform-module-knfsd.nfsproxy_loadbalancer_ipaddress
}

// Prints the DNS address of the Load Balancer
output "load_balancer_dns_address" {
    value = module.terraform-module-knfsd.nfsproxy_loadbalancer_dnsaddress
}
```

Edit the above [configuration variables](#Configuration-Variables) to match your desired configuration.

### Use without Existing Terraform Configuration

If you **do not** currently use Terraform, follow [this guide](https://learn.hashicorp.com/tutorials/terraform/install-cli) to download and install it.

Create a file called `deploy.tf` and add the following contents:

```
module "terraform-module-knfsd" {

    source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.1"

    # Google Cloud Project Configuration
    PROJECT                                     = "my-gcp-project"
    REGION                                      = "us-west1"
    ZONE                                        = "us-west1-a"

    # Network Configuration
    NETWORK                                     = "my-vpc"
    SUBNETWORK                                  = "my-subnet"
    AUTO_CREATE_FIREWALL_RULES                  = false
    LOADBALANCER_IP                             = "10.67.4.5"

    # Knfsd Proxy Configuration
    PROXY_IMAGENAME                             = "knfsd-base-image"
    EXPORT_MAP                                  = "10.0.5.5;/remoteexport;/remoteexport"
    KNFSD_NODES                                 = 3

}

// Prints the IP address of the Load Balancer
output "load_balancer_ip_address" {
    value = module.terraform-module-knfsd.nfsproxy_loadbalancer_ipaddress
}

// Prints the DNS address of the Load Balancer
output "load_balancer_dns_address" {
    value = module.terraform-module-knfsd.nfsproxy_loadbalancer_dnsaddress
}

provider "google" {
  project     = <PROJECT>
  region      = <REGION>
  zone        = <ZONE>
  credentials = file("service-account-key.json")
}
```

Edit the above [configuration variables](#Configuration-Variables) to match your desired configuration.

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

**Warning: The service-account-key.json should be stored securely - do not commit to Git or share with any unauthorised party**

### Deploy Knfsd

Once you have created your `deploy.tf` and created your Service Account you can deploy Knfsd with:

```
terraform init
terraform apply
```

## Configuration Variables

### Google Cloud Project Configuration

| Variable | Description                                                                                                                                                     | Required | Default      |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------ |
| PROJECT  | The Google Cloud Project that the Knfsd Cluster is being deployed to (must also be set in `provider.tf`).                                                       | True     | N/A          |
| REGION   | The [Google Cloud Region](https://cloud.google.com/compute/docs/regions-zones) to use for deployment of regional resources (must also be set in `provider.tf`). | True     | `us-west1`   |
| ZONE     | The [Google Cloud Zone](https://cloud.google.com/compute/docs/regions-zones) to use for deployment of zonal resources (must also be set in `provider.tf`).      | True     | `us-west1-a` |

### Network Configuration

| Variable                   | Description                                                                                                                                                       | Required | Default   |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | --------- |
| NETWORK                    | The network name (VPC) to use for the deployment of the Knfsd Compute Engine Instances.                                                                           | False    | `default` |
| SUBNETWORK                 | The subnetwork name (subnet) to use for the deployment of the Knfsd Compute Engine Instances.                                                                     | False    | `default` |
| AUTO_CREATE_FIREWALL_RULES | Should firewall rules automatically be created to allow [healthcheck connectivity](https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges)? | False    | `true`    |
| LOADBALANCER_IP            | The IP address to use for the Internal Load Balancer. If not specified, a random IP address will be assigned within the subnet.                                   | False    | null      |

### Knfsd Proxy Configuration

| Variable                        | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Required                                             | Default     |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- | ----------- |
| PROXY_BASENAME                  | A nickname to use for this Knfsd deployment (used to ensure uniquely named resources for multiple deployments).                                                                                                                                                                                                                                                                                                                                                 | False                                                | `nfsproxy`  |
| EXPORT_MAP                      | A list of NFS Exports to mount from on-premise and re-export in the format `<SOURCE_IP>;<SOURCE_EXPORT>;<TARGET_EXPORT>`.<br><br> For example to mount `10.0.0.1/export` from on-premise and re-export as `10.100.100.1/reexport` you would set the `EXPORT_MAP` variable to `10.100.100.1;/export;/reexport`.<br><br>You can specify multiple re-exports using a comma, for example `10.100.100.1;/assets;/assetscache,10.100.100.1;/textures;/texturescache`. | `EXPORT_MAP` or `DISCO_MOUNT_EXPORT_MAP` must be set | N/A         |
| DISCO_MOUNT_EXPORT_MAP          | The same as `EXPORT_MAP` but NFS Exports specified via this variable will have crossmounts automatically discovered and re-exported. Workaround to the issue described in [this thread](https://marc.info/?l=linux-nfs&m=161653016627277&w=2). Specifying re-exports via this option will have a small performance impact on cache initialisation as it involves performing a `tree` on each export. This option can be set alongside `EXPORT_MAP`.             | `EXPORT_MAP` or `DISCO_MOUNT_EXPORT_MAP` must be set | N/A         |
| EXPORT_CIDR                     | The CIDR to use in `/etc/exports` of the Knfsd Node for filesystem re-export.                                                                                                                                                                                                                                                                                                                                                                                   | False                                                | `10.0.0.0/8 |
| PROXY_IMAGENAME                 | The name of the Knfsd base [image](https://cloud.google.com/compute/docs/images).                                                                                                                                                                                                                                                                                                                                                                               | True                                                 | N/A         |
| KNFSD_NODES                     | The number of Knfsd nodes to deploy as part of the cluster.                                                                                                                                                                                                                                                                                                                                                                                                     | False                                                | 3           |
| SERVICE_LABEL                   | The Service Label to use for the Forwarding Rule.                                                                                                                                                                                                                                                                                                                                                                                                               | False                                                | `dns`       |
| NCONNECT_VALUE                  | The nconnect value to set on the mount from Knfsd --> Source Filer Rule.                                                                                                                                                                                                                                                                                                                                                                                        | False                                                | `16`        |
| VFS_CACHE_PRESSURE              | The value to set for `vfs_cache_pressure` Rule.                                                                                                                                                                                                                                                                                                                                                                                                                 | False                                                | `100`       |
| NUM_NFS_THREADS                 | The number of NFS Threads to use for KNFSD.                                                                                                                                                                                                                                                                                                                                                                                                                     | False                                                | `512`       |
| ENABLE_AUTOHEALING_HEALTHCHECKS | Should failed healthchecks lead to instance replacement?                                                                                                                                                                                                                                                                                                                                                                                                        | False                                                | `true`      |
| ENABLE_STACKDRIVER_METRICS      | Should Knfsd metrics be exported into Stackdriver?                                                                                                                                                                                                                                                                                                                                                                                                              | False                                                | `true`      |

### Autoscaling Configuration

| Variable                                    | Description                                                                                                                            | Required | Default |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------- |
| ENABLE_KNFSD_AUTOSCALING                    | Should autoscaling be enabled for Knfsd? **You MUST set the `ENABLE_STACKDRIVER_METRICS` variable to `true` if enabling autoscaling**. | False    | `false` |
| KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD | The number of Client Connections to Knfsd that should be targeted for each instance (exceeding will trigger a scale-up).               | False    | `250`   |
| KNFSD_AUTOSCALING_MIN_INSTANCES             | The minimum number of Knfsd instances to set regardless of the traffic volumes.                                                        | False    | `1`     |
| KNFSD_AUTOSCALING_MAX_INSTANCES             | The maximum number of Knfsd instances to set regardless of the traffic volumes.                                                        | False    | `10`    |
