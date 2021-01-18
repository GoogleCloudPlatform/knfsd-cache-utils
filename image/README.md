# Instructions

This directory contains scripts for building an image for Knfsd. We start with a vanilla Ubuntu 20.10 image but make a number of modifications, namely:

* Installation of a newer kernel which has better support for NFS re-exporting
* Installation of supporting components such as `nfs-kernel-server`, `cachefilesd` and the `stackdriver-agent`

This directory contains scripts that will automatically take a vanilla Ubuntu 20.10 image and build the Knfsd Image.

For details of the patches that are applied, see [1_build_image.sh](scripts/1_build_image.sh).

## Usage

### Naviate to Scripts Directory
```
cd image/scripts
```

### Set Variables
```
export BUILD_MACHINE_ZONE=europe-west2-a
export GOOGLE_CLOUD_PROJECT=knfsd-deployment-test
export BUILD_MACHINE_NETWORK=knfsd-test
export BUILD_MACHINE_SUBNET=europe-west2-subnet
```

### Create a Build Machine
```
gcloud compute instances create knfsd-build-machine \
    --zone=$BUILD_MACHINE_ZONE \
    --machine-type=c2-standard-30 \
    --project=$GOOGLE_CLOUD_PROJECT \
    --image=ubuntu-2010-groovy-v20201022a \
    --image-project=ubuntu-os-cloud \
    --network=$BUILD_MACHINE_NETWORK \
    --subnet=$BUILD_MACHINE_SUBNET \
    --boot-disk-size=20GB \
    --boot-disk-type=pd-ssd \
    --metadata-from-file=BUILD_IMAGE_SCRIPT=1_build_image.sh,startup-script=0_init.sh \
    --metadata=serial-port-enable=TRUE
```

### (Optional) Create Firewall Rule for IAP SSH Access
```
gcloud compute firewall-rules create allow-ssh-ingress-from-iap --direction=INGRESS --action=allow --rules=tcp:22 --source-ranges=35.235.240.0/20 --network=$BUILD_MACHINE_NETWORK --project=$GOOGLE_CLOUD_PROJECT
```

### SSH to Build Machine
```
gcloud beta compute ssh knfsd-build-machine --zone=$BUILD_MACHINE_ZONE --tunnel-through-iap
```

### Switch to Root
```
sudo su
```

### Run Build Modified Kernel Script

**This takes a very long time due to the Ubuntu Kernel clone. However it does complete.**

```
cd
./1_build_image.sh
```

### Reboot
```
reboot
```

### SSH to Instance
```
gcloud beta compute ssh knfsd-build-machine --zone=$BUILD_MACHINE_ZONE --tunnel-through-iap
```

### Shutdown Instance
```
sudo shutdown -h now
```

### Create Image
```
gcloud compute images create knfsd-image-5-11-rc3 --source-disk=knfsd-build-machine --source-disk-zone=$BUILD_MACHINE_ZONE
```