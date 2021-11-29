# Next

* (GCP) Fixed specifying project/region/zone
* (GCP) Changed `LOCAL_SSDS` to a simple count of the number of drives.
* (GCP) Prevent mounting over system directories
* (GCP) Added `EXCLUDED_EXPORTS` option to exclude exports from auto-discovery
* (GCP) Added optional UDP load balancer
* (GCP) Fixed remove duplicate and stale exports when restarting
* (GCP) Added configuration for NFS mount options
* (GCP) Added configuration for NFS export options

## (GCP) Fixed specifying project/region/zone

Project, region and zone no longer need to be configured on the Google provider and provided parameters to the module.

If project, region and/or zone are set on the module, these values will be used instead of the values on the provider.

Project, region and zone will default to the provider's value if not set on the module.

## (GCP) LOCAL_SSDS changed to count

`LOCAL_SSDS` is now configured as a simple count of the number of drives. The local SSDS will be named sequentially with the prefix, `local-ssd-`.

**Old:**

```terraform
LOCAL_SSDS = ["local-ssd-1", "local-ssd-2", "local-ssd-3", "local-ssd-4"]
```

**New:**

```terraform
LOCAL_SSDS = 4
```

## (GCP) Prevent mounting over system directories

The `proxy-startup.sh` script now contains a list of protected directories such as `/bin` and `/usr`. Any exports that

When the proxy starts up, check the logs entries such as:

> startup-script: ERROR: Cannot mount 10.0.0.2:/home because /home is a system path

The `/home` directory is included in the list of protected directories to avoid unintended interactions, or issues with the GCP infrastructure such as SSH keys. These can be provisioned automatically on compute instances via OS Login or metadata. Commands such as `gcloud compute ssh` can also create SSH keys. These keys will be created in user home folders in the `/home` directory.

For a full list of the paths, see `PROTECTED_PATHS` in [proxy-startup.sh](../../deployment/terraform-module-knfsd/resources/proxy-startup.sh).

## (GCP) Added `EXCLUDED_EXPORTS` option to exclude exports from auto-discovery

This can be used to exclude specific exports when using auto-discovery such as
`EXPORT_HOST_AUTO_DETECT`. The main use is to exclude any exports that would
try to mount over a a protected directory such as `/home`.

## (GCP) Added optional UDP load balancer

Added an option `ENABLE_UDP` that will deploy a UDP load balancer for the NFS proxy (sharing the same IP as the TCP load balancer).

This is mainly aimed at support for the mount protocol for older clients that default to using UDP. NFS does not recommend using UDP.

### Upgrading existing deployments to support UDP

Existing deployments created the reserved address using the purpose `GCE_ENDPOINT`. To share an IP with multiple load balancers the reserved address' purpose needs to be changed to `SHARED_LOADBALANCER_VIP`.

**NOTE:** You only need to follow these instructions if you want to use the UDP load balancer for existing deployments using v0.3.0 or earlier.

To avoid breaking existing deployments the `google_compute_address` is set to ignore changes to `purpose` in Terraform as most existing deployments will not require UDP. New deployments will set the purpose to `SHARED_LOADBALANCER_VIP`.

**IMPORTANT:** The purpose cannot be changed while the reserved address is in use. To update the purpose you will first need to delete the current TCP load balancer (forwarding rule). **This will prevent the clients from accessing the NFS proxy during the update.**

If you try and set `UDP_ENABLED = true` on an existing deployment you will receive the following error (the IP will match your load balancer's IP):

```text
Error: Error creating ForwardingRule: operation received error: error code "IP_IN_USE_BY_ANOTHER_RESOURCE", message: IP '10.0.0.2' is already being used by another resource.
```

**IMPORTANT:** These instructions only applies when setting `UDP_ENABLED = true` on an existing deployment. If this error occurs when deploying the proxy for the first time check that `LOADBALANCER_IP` is not set to an IP address that is already in use.

Configure your environment (change these values to match your deployment):

```bash
export CLOUDSDK_CORE_PROJECT=my-project
export CLOUDSDK_COMPUTE_REGION=us-west1
export PROXY_BASENAME=rendercluster1
```

Remove the TCP forwarding rule. You can get a list of forwarding rules by running `gcloud compute forwarding-rules list`. The forwarding rule will have the same name as the `PROXY_BASENAME` variable.

```bash
gcloud compute forwarding-rules delete "$PROXY_BASENAME"
```

Get the reserved IP of the address. You can get a list of addresses by running `gcloud compute addresses list`. The address will be named `PROXY_BASENAME-static-ip`.

```bash
gcloud compute addresses describe "$PROXY_BASENAME-static-ip" --format='value(address)'
```

Update Terraform with the IP address for the proxy:

```terraform
module "proxy" {
  source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.3.0"

  LOADBALANCER_IP = "10.0.0.2" # Use the value from the command above
}
```

To update the address' purpose the address needs to be re-created. Delete the address using the gcloud command. Due to dependencies between resources, `terraform taint` cannot be used to automatically delete and re-create the address.

```bash
gcloud compute addresses delete "$PROXY_BASENAME-static-ip"
```

Use Terraform to re-create the reserved address and the forwarding rules:

```bash
terraform apply
```

## (GCP) Fixed remove duplicate and stale exports when restarting

The `/etc/exports` file was not cleared when running the start up script. When rebooting a proxy instance this would create duplicate entries (or leave stale entries) in the `/etc/exports` file.

The `/etc/exports` file is now cleared by the start up script before appending any exports.

## (GCP) Added configuration for NFS mount options

Added variables to Terraform for:

* `ACDIRMIN`
* `ACDIRMAX`
* `ACREGMIN`
* `ACREGMAX`
* `RSIZE`
* `WSIZE`

Also added `MOUNT_OPTIONS` to allow specifying any additional NFS mount options not covered by existing Terraform variables.

## (GCP) Added configuration for NFS export options

Added `EXPORT_OPTIONS` to allow specifying custom NFS export options.
