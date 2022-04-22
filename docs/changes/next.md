# Next

* Use latest 5.13 HWE kernel
* Use metric labels for mount stats

## Use latest 5.13 HWE kernel

The image build script started failing with the error:

```text
E: Version '5.13.0.39.44~20.04.24' for 'linux-generic-hwe-20.04' was not found
E: Version '5.13.0.39.44~20.04.24' for 'linux-image-generic-hwe-20.04' was not found
E: Version '5.13.0.39.44~20.04.24' for 'linux-headers-generic-hwe-20.04' was not found
```

The `linux-image-hwe-20.04` package only keeps the binaries for the latest HWE kernel. As such the HWE kernels cannot be pinned to a specific kernel version.

## Use metric labels for mount stats

The mount stats were previously using resource labels for `server`, `path` and `instance`. This was intended to reduce the volume of data being logged by reducing repeated values.

However, GCP Cloud Monitoring does not support custom resource labels. This is likely to be a common issue with other reporting systems either handling resource labels differently, or ignoring them completely.

To avoid issues the labels for `server`, `path` and `instance` are now reported as metric level labels.
