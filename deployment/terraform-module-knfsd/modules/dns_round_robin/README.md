# KNFSD DNS Round Robin Module

This Terraform module configures Cloud DNS to use DNS Round Robin for a KNFSD proxy cluster. This offers an alternative to using the load balancer.

This module is not generally intended to be used directly, and is included by the main KNFSD proxy module when `TRAFFIC_DISTRIBUTION_MODE = "dns_round_robin"`.

## Prerequisites

To use DNS Round Robin the KNFSD proxy cluster must be configured with:

* `ENABLE_KNFSD_AUTOSCALING = false`
* `ASSIGN_STATIC_IPS        = true`

## Inputs

* `project` - (optional) The ID of the project in which the resource belongs. If it is not provided, the provider project is used.

* `networks` - (optional) Set of networks to attach Cloud DNS to. These should be formatted like `projects/{project}/global/networks/{network}` or `https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}`. This defaults to the same network as the KNFSD proxy cluster.

* `instance_group` - The full URL (self link) of the KNFSD proxy instance group.

* `proxy_basename` - The proxy basename of the KNFSD proxy cluster.

* `dns_name` - (optional) The fully qualified domain name (FQDN) to assign the KNFSD proxy cluster. This defaults to `{proxy_basename}.knfsd.internal.`.

* `knfsd_nodes` - The number of instances (size) of the KNFSD instance group.

## Outputs

* `dns_name` - The DNS name that was created for the KNFSD proxy cluster.

**NOTE:** `dns_name` is useful if you're creating other resources in the same Terraform configuration that depend on the DNS entry to create a dependency between the DNS entry and the other resources.
