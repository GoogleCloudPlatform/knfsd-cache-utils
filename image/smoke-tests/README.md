# Smoke Tests

The smoke tests are intended to be executed as part of an automated build process (e.g. using Cloud Build) to verify the basic functionality of new knfsd proxy images.

## Prerequisites

You will need a GCP project with the following resources:

* GCP Services:
  * Cloud DNS API (`dns.googleapis.com`)
  * Compute Engine API (`compute.googleapis.com`)
  * Cloud Filestore API (`file.googleapis.com`)
  * Service Networking API (`servicenetworking.googleapis.com`)

* VPC Network

  This VPC network will be used to host temporary infrastructure such as a knfsd proxy instances and NFS clients. These resources will be created as part of the test, and removed at the end.

  * [Private service access](https://cloud.google.com/vpc/docs/private-services-access) configured to allow Cloud Filestore.

* Firewall Rules:

  * Health Check (GCP health check servers to knfsd proxy)
    * **TCP port**: `2049`
    * **Targets**: *Specified target tags*
    * **Target Tags**: `knfsd-cache-server`
    * **Source IPv4 Ranges**: `130.211.0.0/22`, `35.191.0.0/16`, `209.85.152.0/22`, `209.85.204.0/22`

  * NFS (client to knfsd proxy)
    * **TCP ports**: `111`, `2049`, `20048`, `20050`, `20051`, `20052`, `20053`
    * **Targets**: *Specified target tags*
    * **Target Tags**: `knfsd-cache-server`
    * **Source Tags**: `nfs-client`

  * HTTP (client to knfsd proxy)
    * **TCP Ports**: `80`
    * **Targets**: *Specified target tags*
    * **Target Tags**: `knfsd-cache-server`
    * **Source Tags**: `nfs-client`

* Compute Images:
  * [Knfsd Proxy Image](../)
  * [NFS Client Image](../../testing/images/client/)

## Local Development

### Development Prerequisites

You will need the following software on your local machine:

* [Terraform](https://www.terraform.io/)
* [GCloud SDK](https://cloud.google.com/sdk/docs/install)
* [Bash](https://www.gnu.org/software/bash/)
* [Go 1.20](https://go.dev/) or higher
* [GNU Make](https://www.gnu.org/software/make/)

You will also need to enable and configure [IAP for TCP forwarding](https://cloud.google.com/iap/docs/using-tcp-forwarding) so that the tests can access the NFS client running in GCP. Remember to include a firewall rule to allow TCP port 22 from `35.235.240.0/20`.

### Configure Terraform

The Terraform will create an isolated KNFSD proxy for testing, including an independent VPC network so as not to conflict with any other machines. The intent is that the resources will be created to run the smoke tests, then destroyed once the smoke test is complete.

**NOTE: The Terraform state is stored on your local machine.** Do not remove this state until you have finished the tests so that you can easily destroy the resources that were created.

Create a `terraform.tfvars` file in `smoke-tests/terraform`.

```terraform
project     = "my-project"
region      = "us-central1"
zone        = "us-central1-a"
proxy_image = "knfsd-image"
```

**IMPORTANT**: Normally you should not run terraform directly. The test harness uses terratest to automatically deploy the required GCP resources as part of the test.

### Run the tests

```sh
make test
```

This will create the GCP resources, run the tests then destroy the resources.

When developing tests though, it is useful to deploy the infrastructure once, then run the tests several times before removing the infrastructure.

* `make apply` runs `terraform apply` creates or updates the infrastructure without running any tests.

* `make check` runs the tests without creating or destroying any infrastructure.

* `make destroy` runs `terraform destroy` to remove the infrastructure.
