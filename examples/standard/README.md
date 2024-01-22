# Standard Example

This example shows a standard KNFSD proxy deployment.

As a starting point, the example deploys 3 KNFSD proxy instances, with 4 local SSDs.

The deployment uses an external Cloud SQL database to store FSIDs to ensure all instances in the cluster allocate the same FSID to each export.

## Firewall Rules

You will need to create the firewall rules for:

* KNFSD proxy cluster Health checks
* NFS traffic; proxy to source.
* NFS traffic; clients to proxy.

See [Firewall configuration](../../deployment/firewall.md).

## KNFSD Proxy Service Account

The KNFSD proxy service account will need the following project level IAM permissions:

* Logs Writer (`roles/logging.logWriter`)
* Monitoring Metric Writer (`roles/monitoring.metricWriter`)

## Inputs

* `project` - (Required) The ID of the GCP project where the KNFSD caching proxy will be deployed.

* `region` - (Required) The GCP region where the KNFSD caching proxy will be deployed.

* `zone` - (Required) The GCP zone where the KNFSD caching proxy will be deployed.

* `name` - (Required) The name of the deployment. Resources created will be prefixed with this name.

* `network` - (Optional) The GCP network to attach the KNFSD caching proxy to. Defaults to "default".

* `subnetwork` - (Optional) The GCP subnetwork to attach the KNFSD caching proxy to. Defaults to "default".

* `proxy_image` - (Required) The Compute Image to use for the KNFSD caching proxy. This should be built using the image build script in <../../image/>.

* `proxy_service_account` - (Required) The GCP Service Account to use for the KNFSD proxy cluster.

* `export_map` - (Required) A list of NFS exports to mount from the source and re-export in the format `<SOURCE_IP>;<SOURCE_EXPORT>;<TARGET_EXPORT>`. See [KNFSD Deployment](../../deployment/README.md) for more details.

## Outputs

* `project` - GCP project where the resources were created.

* `region` - GCP region where the resources were created.

* `zone` - GCP zone where the resources were created.

* `proxy_instance_group` - Name of the KNFSD proxy instance group.

* `proxy_host` - DNS name of the KNFSD proxy group.
