# KNFSD proxy load balancer module

This module supports deploying an Internal TCP Load Balancer TCP (and optionally UDP) for the KNFSD proxy cluster.

This module is not generally intended to be used directly, and is included by the main KNFSD proxy module when `LOADBALANCING_MODE = "loadbalancer"`.

## Prerequisites

These prerequisites are normally created by the main KNFSD proxy module.

### Reserved IP Address

This module expects that `IP_ADDRESS` is a reserved internal IP address. The IP address must have a purpose of `SHARED_LOADBALANCER_VIP` to allow the same IP address to be used by multiple forwarding rules.

### Health Checks

This module does not create the health checks required by the load balancer. This allows the same health check to be shared between both the load balancer and MIG auto-healing.

### Firewall Rules

This module does not create any firewall rules. Any firewall rules required by the health checks must created separately.

## Inputs

* PROJECT - (optional) The Google Cloud Project that the load balancer is being deployed to. If it is not provided, the provider project is used.

* REGION - (optional) The [Google Cloud Region](https://cloud.google.com/compute/docs/regions-zones) to use for deployment of the load balancer. If it is not provided, the provider region is used.

* PROXY_BASENAME - A nickname to use for this Knfsd deployment (used to ensure uniquely named resources for multiple deployments).

* NETWORK - The network name (VPC) to use for the deployment of the Internal Load Balancer.

* SUBNETWORK - The subnetwork name (subnet) to use for the deployment of the Internal Load Balancer.

* SERVICE_LABEL - (optional) The Service Label to use for the Forwarding Rule. The default is `"dns"`.

* IP_ADDRESS - The IP address to use for the Internal Load Balancer.

* ENABLE_UDP - (optional) Create a load balancer to support UDP traffic to the NFS proxy instances. UDP is not recommended for the main NFS traffic as it can cause data corruption. However, this maybe useful for older clients that default to using UDP for the mount protocol. The default is `false`.

* HEALTH_CHECK - The URL (self link) to the health check to use for the load balancer.

* INSTANCE_GROUP - The URL (self link) to the managed instance group of the KNFSD proxy cluster.

## Outputs

* dns_address - The internal DNS address of the load balancer.
