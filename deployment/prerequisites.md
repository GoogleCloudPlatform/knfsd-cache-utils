# Prerequisites

Before deploying Knfsd to Google Cloud there are a number of prerequisites that you should ensure are met.

These are detailed below.

## Knfsd Disk Image

Before you can deploy Knfsd in your Google Cloud Project you first need to build a disk image. The process for building this image is documented [here](../image).

## Firewall Rules

This Terraform module will automatically create the required firewall rules to allow healthchecking the Knfsd instances (this can be disabled). However it **will not** create any other firewall rules.

You should make sure that you implement firewall rules to allow:

* Knfsd Node --> Source NFS Server Communication
* NFS Clients --> Knfsd Node Communication (See [ports.md](ports.md) for more information on the ports used by Knfsd)

## Metrics

Knfsd supports a range of metrics which can be automatically exported into Google Cloud Operations. In order for these metrics to work correctly you need to complete some setup steps.

### Setup Metric Descriptors and Monitoring Dashboard

**This step only needs to be ran once per GCP Project**. In order for the Knfsd custom metrics to be correctly formatted in Google Cloud Operations you need to create the Metric Descriptors. Instructions on how to do this are available [here](metrics/README.md). This process also imports the custom Google Cloud Monitoring dashboard which can be used to easily understand Knfsd performance.

### Private Google Access
You must have [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access) enabled on the subnet that you will be using for the Knfsd Nodes. This is required to allow connectivity to the Monitoring API for VM's without a Public IP address.

### Service Account Permissions
A Service Account needs to be configured for the Knfsd Nodes with the `logging-write` and `monitoring-write` scopes. This is performed automatically by the Terraform Module when you have metrics enabled. 

By default, the [Compute Engine Default Service Account](https://cloud.google.com/compute/docs/access/service-accounts#default_service_account) will be used. 

You need to make sure the identity used for Terraform has the `roles/iam.serviceAccountUser` role on the Compute Engine Default Service Account so that it can assign it to the Knfsd Nodes.

### Enable Metrics in Terraform Variables
By default Knfsd Metrics are disabled, to enable you need to set the `ENABLE_STACKDRIVER_METRICS` variable to `true` as described in [configuration.md](configuration.md).

## gcloud

You should ensure that you have [gcloud](https://cloud.google.com/sdk/install) instaled and configured on your machine. 

## Terraform

You should ensure that you have [Terraform](https://www.terraform.io/downloads.html) 0.12+ installed.
