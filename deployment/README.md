# terraform-module-knfsd

This directory contains a [Terraform Module](https://www.terraform.io/docs/modules/index.html) for deploying a Knfsd cluster on Google Cloud.

**Note:** The `main` branch may be updated at any time with the latest changes which could be breaking. You should always configure your module to use a release. This can be configured in the modules Terraform Configuration block.

```
source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.5.1"
```

## Prerequisites

Before continuing with the deployment and configuration of Knfsd you should review the deployment [prerequisites](prerequisites.md).

## Features

There are a number of optional features that can be enabled and configured for Knfsd. If you are planning on using any of these features then please review the appropriate documentation section.

* [Metrics](metrics.md) - System and proxy metrics for monitoring and observing Knfsd
* [Autoscaling](autoscaling.md) - Automatic scale up of Knfsd in response to the number of clients
* [Agent](agent.md) - A lightweight API that provides information on Knfsd nodes

## Usage

**Note: See the [Configuration Variables](#Configuration-Variables) section for advance configuration options**

Basic usage of this module is as follows:

```terraform
module "nfs_proxy" {
    source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.5.1"

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
    source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.5.1"

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
| EXCLUDED_EXPORTS                | A list of filter patterns to be excluded from auto-discovery (see [Filter Patterns](#filter-patterns)). Auto-discovery will ignore any exports that match any of the exclude patterns. Does not apply to mounts specified in the `EXPORT_MAP`. Paths filtered from auto-discovery can be explicitly exported using `EXPORT_MAP`, this can be used to change the export path.                                                                                    | `false`                                                                              | `[]`            |
| INCLUDED_EXPORTS                | If set, auto-discovery will only include paths matching a filter pattern from the include list (see [Filter Patterns](#filter-patterns)). Does not apply to mounts specified in the `EXPORT_MAP`. Paths filtered from auto-discovery can be explicitly exported using `EXPORT_MAP`, this can be used to change the export path.                                                                                                                                 | `false`                                                                              | `[]`            |
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
| SERVICE_ACCOUNT                 | Service account the NFS proxy compute instances will run with.                                                                                                                                                                                                                                                                                                                                                                                                  | False                                                                                | see service account |

#### Service Account

The default `SERVICE_ACCOUNT` depends on `ENABLE_STACKDRIVER_METRICS`.

* When `false` the NFS proxy instances will a service account
* When `true` the NFS proxy instances will use the default compute service account with the following scopes:
  * `https://www.googleapis.com/auth/logging.write`
  * `https://www.googleapis.com/auth/monitoring.write`

If a custom service account is assigned the compute instances will use the `https://www.googleapis.com/auth/cloud-platform` scope. This allows the proxy instances access to any GCP API permitted by IAM.

The service account will need the following project level IAM permissions:

* Logs Writer (`roles/logging.logWriter`)
* Monitoring Metric Writer (`roles/monitoring.metricWriter`)

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

**NOTE:** A service account must be assigned to allow the proxy access to the GCP secret. The service account will need to be granted the following permissions on the NetApp secret:

* Secret Manager Viewer (`roles/secretmanager.viewer`)
* Secret Manager Secret Accessor (`roles/secretmanager.secretAccessor`)

**IMPORTANT:** *Do not* assign the permissions at the project level as this will allow the NetApp proxy to read any secret in the project. Assign the IAM permissions directly on the NetApp secret.

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

### Export Options

| Variable       | Description                                                                       | Required | Default   |
| -------------- | --------------------------------------------------------------------------------- | -------- | --------- |
| NOHIDE         | When `true`, adds the `nohide` option to all the exports.                         | False    | `true`    |
| EXPORT_OPTIONS | Any custom NFS exports options. These options will be applied to all NFS exports. | False    | `""`      |

### Autoscaling Configuration

| Variable                                    | Description                                                                                                                            | Required | Default |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------- |
| ENABLE_KNFSD_AUTOSCALING                    | Should autoscaling be enabled for Knfsd? **You MUST set the `ENABLE_STACKDRIVER_METRICS` variable to `true` if enabling autoscaling**. | False    | `false` |
| KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD | The number of Client Connections to Knfsd that should be targeted for each instance (exceeding will trigger a scale-up).               | False    | `250`   |
| KNFSD_AUTOSCALING_MIN_INSTANCES             | The minimum number of Knfsd instances to set regardless of the traffic volumes.                                                        | False    | `1`     |
| KNFSD_AUTOSCALING_MAX_INSTANCES             | The maximum number of Knfsd instances to set regardless of the traffic volumes.                                                        | False    | `10`    |

## Filter Patterns

Filter patterns use simple glob-style wildcard patterns. A single asterisk `*` will match any character except `/`. A double asterisk will match all the descendants of path.

|                        | `/home` | `/home/*` | `/home/**` |
| ---------------------- | :-----: | :-------: | :--------: |
| `/home`                | **✔**   | **✘**     | **✘**      |
| `/home/alice`          | **✘**   | **✔**     | **✔**      |
| `/home/alice/projects` | **✘**   | **✘**     | **✔**      |

NOTE: Filter patterns ending in a wildcard *will not* match the parent path. You need to add both the parent path, and the child patterns.

```terraform
# To exclude /home and all its descendants:
EXCLUDED_EXPORTS = ["/home", "/home/**"]
```

| Special Terms | Meaning
| ------------- | -------
| `*`           | matches any sequence of characters except `/`
| `/**`         | matches zero or more directories
| `?`           | matches any single character except `/`
| `[class]`     | matches any single character except `/` against a class of characters
| `{alt1,...}`  | matches a sequence of characters if one of the comma-separated alternatives matches

### Character Classes

Character classes support the following:

| Class      | Meaning
| ---------- | -------
| `[abc]`    | matches any single character within the set
| `[a-z]`    | matches any single character in the range
| `[^class]` | matches any single character which does *not* match the class
| `[!class]` | same as `^`: negates the class

### Combining include and exclude patterns

Include and exclude patterns can be combined. For an export to be accepted (and re-exported), the export *must* match an include pattern, and *must not* match an exclude pattern.

## Caveats

### Excluding nested mounts

If you exclude a nested mount but still export the parent mount you may get I/O errors when accessing the nested mount.

The exact behaviour will depend on how the source server has exported the nested mount.

If the source server exports the mount with the `crossmnt`, or `nohide` options then trying to access the nested mount, or list the directory containing the nested mount will result in I/O errors.

If the source server exports the mount without `crossmnt`, or `hide` options then the directory for the nested mount will be visible, but empty.

It is advised that if you exclude a nested mount, you also exclude the parent mount. You may however exclude a parent mount but include a nested mount.

For example, if you have the following mounts:

```text
/assets
/assets/common
/assets/common/textures
```

You could exclude `/assets`, but still export `/assets/common` and `/assets/common/textures`. You could also export only `/assets/common/textures`.

However, exporting `/assets` but excluding `/assets/common` could cause errors.

Exporting `/assets` and `/assets/common/textures`, but excluding `/assets/common` will likely fail, and can have unintended side-effects as the proxy will try to create the directory `/assets/common`.

### Combining auto-discovery and explicit mounts

While auto-discovery and explicit mounts can be combined the system does not have any special handling for duplicate paths.

As such it is not recommended to combine multiple auto-discovery methods, or explicit (`EXPORT_MAP`).

The behaviour of duplicates is undefined. The system might overwrite one mount with another, or it may error.

### Limitations on export names

The proxy cannot re-export any path that matches a symlink on the local server.

The most likely symlinks that will cause conflicts are:

* `/bin`
* `/lib`
* `/lib32`
* `/lib64`
* `/libx32`
* `/sbin`

The proxy will fail to start if it attempts to export a path that matches a symlink. Check the logs for errors such as:

```text
ERROR: Cannot mount 10.0.0.2:/bin because /bin matches a symlink
```

If you are providing a manual export list, specify a different path for the export, such as `10.0.0.2;/bin;/binaries`.

If you're using auto-discovery add the path to the list of excluded exports, for example `EXCLUDED_EXPORTS = ["/bin"]`

For a full list of symlinks, start a compute instance using the proxy image (without the standard startup script) and run the command:

```bash
find / -type l
```

Most of the symlinks listed are unlikely to cause issues, such as `/usr/lib/x86_64-linux-gnu/libc.so.6`.
