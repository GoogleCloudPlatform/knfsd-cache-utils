# Next

* (GCP) Changed `LOCAL_SSDS` to a simple count of the number of drives.
* (GCP) Prevent mounting over system directories

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

> startup-script: WARN: skipping export /home because it is a system path

The `/home` directory is included in the list of protected directories to avoid unintended interactions, or issues with the GCP infrastructure such as SSH keys. These can be provisioned automatically on compute instances via OS Login or metadata. Commands such as `gcloud compute ssh` can also create SSH keys. These keys will be created in user home folders in the `/home` directory.

For a full list of the paths, see `PROTECTED_PATHS` in [proxy-startup.sh](../../deployment/terraform-module-knfsd/resources/proxy-startup.sh).
