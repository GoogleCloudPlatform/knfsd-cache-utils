# Basic Example

This example provides a very simple, minimal, deployment.
This can be used to verify that the KNFSD caching proxy can be deployed, and can connect to the source NFS server.

For simplicity this example uses some settings that are not recommended for production:

* `AUTO_CREATE_FIREWALL_RULES = true`.

  This setting automatically creates the firewall rules required for health checks.
  Normally for production you should create the firewall rules yourself.
  If you try and deploy more than one KNFSD proxy cluster with `AUTO_CREATE_FIREWALL_RULES` set to true Terraform will fail because the firewall rule already exists.

  You will still need to create firewall rules to allow NFS traffic from the proxy to the source, and the clients to the proxy.
  See [Firewall configuration](../../deployment/firewall.md).

* `FSID_MODE = "static"`

  This deploys the KNFSD proxy cluster without a shared FSID database.
  In production it is recommended to use a shared (external) FSID database to ensure that all the KNFSD proxy instances in the cluster allocate the same FSID to each export.

* Uses the default Compute Service Account.

  For best practice you should create a specific Service Account for the KNFSD proxy cluster to use.
  A Service Account is required when using an external FSID database.

## Inputs

* `project` - (Required) The ID of the GCP project where the KNFSD caching proxy will be deployed.

* `region` - (Required) The GCP region where the KNFSD caching proxy will be deployed.

* `zone` - (Required) The GCP zone where the KNFSD caching proxy will be deployed.

* `name` - (Required) The name of the deployment. Resources created will be prefixed with this name.

* `network` - (Optional) The GCP network to attach the KNFSD caching proxy to. Defaults to "default".

* `subnetwork` - (Optional) The GCP subnetwork to attach the KNFSD caching proxy to. Defaults to "default".

* `proxy_image` - (Required) The Compute Image to use for the KNFSD caching proxy. This should be built using the image build script in <../../image/>.

* `export_map` - (Required) A list of NFS exports to mount from the source and re-export in the format `<SOURCE_IP>;<SOURCE_EXPORT>;<TARGET_EXPORT>`. See [KNFSD Deployment](../../deployment/README.md) for more details.

## Outputs

* `project` - GCP project where the resources were created.

* `region` - GCP region where the resources were created.

* `zone` - GCP zone where the resources were created.

* `proxy_instance_group` - Name of the KNFSD proxy instance group.

* `proxy_host` - DNS name of the KNFSD proxy group.
