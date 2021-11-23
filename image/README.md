# Instructions

This directory contains scripts for building an image for Knfsd. We start with a vanilla Ubuntu 20.10 image but make a number of modifications, namely:

* Installation of a newer kernel which has better support for NFS re-exporting
* Installation of supporting components such as `nfs-kernel-server`, `cachefilesd` and the `stackdriver-agent`

This directory contains scripts that will automatically take a vanilla Ubuntu 20.10 image and build the Knfsd Image.

For details of the patches that are applied, see [1_build_image.sh](scripts/1_build_image.sh).

## Essential Package Versions

For the image we are using a specific version for this packages:

* stackdriver-agent=6.1.4-1.focal
* cachefilesd=0.10.10-0.2ubuntu1
* rpcbind=1.2.5-8
* nfs-kernel-server=1:1.3.4-2.5ubuntu3.4
* nfs-utils=2.5.3
* kernel=5.11.8-051108

## Build Using Packer

Download Packer 1.7.8 or higher from <https://packer.io/downloads>.

### Create Packer Variables File

Create a packer variables file (e.g. `image.pkrvars.hcl`).

#### Required

* project (string) - GCP Project where the image will be built and stored.
* zone (string) -  The zone in which to launch the instance used to create the image. Example: `"us-west1-a"`.

#### Optional

* machine_type (string) - The machine type used to build the image. This can be increased to improve build speeds. Defaults to `"e2-standard-8"`.
* network_project (string) - Project hosting the network when using shared VPC. Defaults to `project`.
* subnetwork (string) - The subnetwork the compute instance will use. Defaults to `"default"`.
* omit_external_ip (bool) - Use a private (internal) IP only. Defaults to `true`.
  * The subnetwork will require a NAT router in the region.
  * IAP tunnelling (default) or a VPN connection to the VPC network is required.
* image_name (string) - The unique name of the resulting image. Defaults to `"{image_family}-{timestamp}"`.
* image_family (string) - The name of the [image family](https://cloud.google.com/compute/docs/images/image-families-best-practices) to which the resulting image belongs. Defaults to `"nfs-proxy"`.
* image_storage_location (string) - [Storage location](https://cloud.google.com/compute/docs/images/create-delete-deprecate-private-images#selecting_image_storage_location), either regional or multi-regional, where snapshot content is to be stored. Defaults to a nearby regional or multi-regional location.
* use_iap (bool) - Whether to use an IAP proxy. Defaults to `true`.
* use_internal_ip (bool) - If true, use the instance's internal IP instead of its external IP during building. Defaults to `true`.
* skip_create_image (bool) - Skip creating the image. Useful for setting to `true` during a build test stage. Defaults to `false`.

#### Example

```hcl
project = "my-gcp-project"
zone    = "us-west1-a"
```

### Run Packer Build

```bash
packer build -vars-file image.pkrvars.hcl image
```

### Run Smoke Tests

You can use the [smoke test suite](smoke-tests/README.md) to verify the basic functionality of the image.

## Build Manually

### Navigate to Image Directory

```bash
cd image
```

### Update settings in brackets <> below and set variables

```bash
export BUILD_MACHINE_NAME=knfsd-build-machine
export BUILD_MACHINE_ZONE=<europe-west2-a>
export GOOGLE_CLOUD_PROJECT=<knfsd-deployment-test>
export BUILD_MACHINE_NETWORK=<knfsd-test>
export BUILD_MACHINE_SUBNET=<europe-west2-subnet>
export IMAGE_FAMILY=knfsd-proxy
export IMAGE_NAME="${IMAGE_FAMILY}-$(date -u +%F-%H%M%S)"
```

### Create Build Machine

The standard machine type is e2-standard-8, you can use a higher one if you want to improve build speed.

**IMPORTANT:** It is recommended not to change the disk boot-disk-size. The performance can increase but the image size will be larger and this will force any VMs to use the larger size drive.

```bash
gcloud compute instances create $BUILD_MACHINE_NAME \
    --zone=$BUILD_MACHINE_ZONE \
    --machine-type=e2-standard-8 \
    --project=$GOOGLE_CLOUD_PROJECT \
    --image-family=ubuntu-2004-lts \
    --image-project=ubuntu-os-cloud \
    --network=$BUILD_MACHINE_NETWORK \
    --subnet=$BUILD_MACHINE_SUBNET \
    --boot-disk-size=20GB \
    --boot-disk-type=pd-ssd \
    --metadata=serial-port-enable=TRUE,block-project-ssh-keys=TRUE
```

### (Optional) Create Firewall Rule for IAP SSH Access

```bash
gcloud compute firewall-rules create allow-ssh-ingress-from-iap --direction=INGRESS --action=allow --rules=tcp:22 --source-ranges=35.235.240.0/20 --network=$BUILD_MACHINE_NETWORK --project=$GOOGLE_CLOUD_PROJECT
```

### Copy Resources to Build Machine

```bash
gcloud compute scp --recurse resources/* build@$BUILD_MACHINE_NAME: --zone=$BUILD_MACHINE_ZONE --tunnel-through-iap --project=$GOOGLE_CLOUD_PROJECT
```

**NOTE:** You might get some errors when connecting while the instance is still booting. These errors will be generic network errors, or errors exchanging keys such as:

```text
ERROR: (gcloud.compute.start-iap-tunnel) Error while connecting [4047: 'Failed to lookup instance'].

ERROR: (gcloud.compute.start-iap-tunnel) Error while connecting [4003: 'failed to connect to backend']. (Failed to connect to port 22)
```

### SSH to Build Machine

```bash
gcloud compute ssh build@$BUILD_MACHINE_NAME --zone=$BUILD_MACHINE_ZONE --tunnel-through-iap --project=$GOOGLE_CLOUD_PROJECT
```

### Run the Build Image Script

```bash
sudo bash scripts/1_build_image.sh
```

When this script completes you should see:

```text
SUCCESS: Please reboot for new kernel to take effect
```

### Reboot Build Machine

Once the build image script has completed, check there were no errors and reboot the machine. This will restart the build machine with the new kernel.

```bash
sudo reboot
```

**NOTE: When your Build Machine reboots, your Cloud Console will revert to your host machine.**

### SSH to Build Machine to run subsequent commands

```bash
gcloud compute ssh build@$BUILD_MACHINE_NAME --zone=$BUILD_MACHINE_ZONE --tunnel-through-iap --project=$GOOGLE_CLOUD_PROJECT
```

### Validate Kernel Version

Verify that the build machine booted using the new kernel version.

```bash
uname -r
```

**Output from above command should indicate kernel version `5.11.8-051108-generic`.**

### Finalize the Build Machine

This will clean up the local disk prior to creating the image (such as removing the build user).

Once the clean up is complete, the instance will shutdown.

```bash
sudo bash scripts/9_finalize.sh
```

You will get the following warnings, these can safely be ignored:

```text
userdel: user build is currently used by process 1431
userdel: build mail spool (/var/mail/build) not found
```

### On your Cloud Shell host machine, create the Custom Disk Image

```bash
gcloud compute images create $IMAGE_NAME --family=$IMAGE_FAMILY --source-disk=$BUILD_MACHINE_NAME --source-disk-zone=$BUILD_MACHINE_ZONE --project=$GOOGLE_CLOUD_PROJECT
```

### Delete Build Machine

```bash
gcloud compute instances delete $BUILD_MACHINE_NAME --zone=$BUILD_MACHINE_ZONE --project=$GOOGLE_CLOUD_PROJECT
```
