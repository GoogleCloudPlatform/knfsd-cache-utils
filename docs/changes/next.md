# Next

* Use latest 5.13 HWE kernel

## Use latest 5.13 HWE kernel

The image build script started failing with the error:

```text
E: Version '5.13.0.39.44~20.04.24' for 'linux-generic-hwe-20.04' was not found
E: Version '5.13.0.39.44~20.04.24' for 'linux-image-generic-hwe-20.04' was not found
E: Version '5.13.0.39.44~20.04.24' for 'linux-headers-generic-hwe-20.04' was not found
```

The `linux-image-hwe-20.04` package only keeps the binaries for the latest HWE kernel. As such the HWE kernels cannot be pinned to a specific kernel version.
