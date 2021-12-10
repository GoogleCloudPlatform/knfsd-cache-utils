# Instructions

This directory contains scripts for building an image for Knfsd. We start with a vanilla Ubuntu 20.10 image but make a number of modifications, namely:

* Installation of a newer kernel which has better support for NFS re-exporting
* Installation of supporting components such as `nfs-kernel-server`, `cachefilesd` and the `stackdriver-agent`

This directory contains scripts that will automatically take a vanilla Ubuntu 20.10 image and build the Knfsd Image.

For details of the patches that are applied, see [1_build_image.sh](scripts/1_build_image.sh).

## Usage

### Navigate to Scripts Directory

```bash
cd image/scripts
```

### Update settings in brackets <> below and set variables

```bash
export BUILD_MACHINE_NAME=knfsd-build-machine
export BUILD_MACHINE_ZONE=<europe-west2-a>
export GOOGLE_CLOUD_PROJECT=<knfsd-deployment-test>
export BUILD_MACHINE_NETWORK=<knfsd-test>
export BUILD_MACHINE_SUBNET=<europe-west2-subnet>
```

### Create Build Machine

```bash
gcloud compute instances create $BUILD_MACHINE_NAME \
    --zone=$BUILD_MACHINE_ZONE \
    --machine-type=c2-standard-16 \
    --project=$GOOGLE_CLOUD_PROJECT \
    --image=ubuntu-2010-groovy-v20210323 \
    --image-project=ubuntu-os-cloud \
    --network=$BUILD_MACHINE_NETWORK \
    --subnet=$BUILD_MACHINE_SUBNET \
    --boot-disk-size=20GB \
    --boot-disk-type=pd-ssd \
    --metadata-from-file=BUILD_IMAGE_SCRIPT=1_build_image.sh,startup-script=0_init.sh \
    --metadata=serial-port-enable=TRUE
```

### (Optional) Create Firewall Rule for IAP SSH Access

```bash
gcloud compute firewall-rules create allow-ssh-ingress-from-iap --direction=INGRESS --action=allow --rules=tcp:22 --source-ranges=35.235.240.0/20 --network=$BUILD_MACHINE_NETWORK --project=$GOOGLE_CLOUD_PROJECT
```

### Copy Knfsd Agent Source Code to Machine

```bash
gcloud compute scp --recurse ../knfsd-agent build@$BUILD_MACHINE_NAME: --zone=$BUILD_MACHINE_ZONE --tunnel-through-iap --project=$GOOGLE_CLOUD_PROJECT
```

**NOTE:** You might get some errors when connecting while the instance is still booting. These errors will be generic network errors, or errors exchanging keys such as:

```text
ERROR: (gcloud.compute.start-iap-tunnel) Error while connecting [4047: 'Failed to lookup instance'].

ERROR: (gcloud.compute.start-iap-tunnel) Error while connecting [4003: 'failed to connect to backend']. (Failed to connect to port 22)
```

### SSH to Build Machine

```bash
gcloud compute ssh $BUILD_MACHINE_NAME --zone=$BUILD_MACHINE_ZONE --tunnel-through-iap --project=$GOOGLE_CLOUD_PROJECT
```

### Switch to Root

```bash
sudo su
cd
```

### Move Knfsd Agent Source Code

```bash
mv /home/build/knfsd-agent .
```

### Run the script 1_build_image.sh

```bash
./1_build_image.sh
```

### After the installation is complete, reboot your Build Machine to run the updated code

```bash
reboot
```

**NOTE: When your Build Machine reboots, your Cloud Console will revert to your host machine.**

### SSH to your Build Machine to run subsequent commands

```bash
gcloud compute ssh $BUILD_MACHINE_NAME --zone=$BUILD_MACHINE_ZONE --tunnel-through-iap --project=$GOOGLE_CLOUD_PROJECT
```

### Validate Newer Kernel version is installed

```bash
uname -r
```

**Output from above command should indicate kernel version `5.11.8-051108-generic`.**

### Shutdown Instance

```bash
sudo shutdown -h now
```

### On your Cloud Shell host machine, create the Custom Disk Image

```bash
gcloud compute images create knfsd-image --source-disk=$BUILD_MACHINE_NAME --source-disk-zone=$BUILD_MACHINE_ZONE --project=$GOOGLE_CLOUD_PROJECT
```

### Delete Build Machine

```bash
gcloud compute instances delete $BUILD_MACHINE_NAME --zone=$BUILD_MACHINE_ZONE --project=$GOOGLE_CLOUD_PROJECT
```
