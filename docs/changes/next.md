# Next

* Pin APT packages for the HWE kernel

## Pin APT packages for the HWE kernel

The image build script started failing with the error:

```text
The following packages have unmet dependencies:
 linux-generic-hwe-20.04 : Depends: linux-image-generic-hwe-20.04 (= 5.13.0.39.44~20.04.24) but 5.13.0.40.45~20.04.25 is to be installed
                           Depends: linux-headers-generic-hwe-20.04 (= 5.13.0.39.44~20.04.24) but 5.13.0.40.45~20.04.25 is to be installed
```

Even though the `linux-image-hwe-20.04` package for `5.13.0.39.44~20.04.24` explicitly states the dependencies' versions, APT is trying to then install the latest version of the dependencies. This then fails because the newer versions do not meet the package constraints.

To avoid this issue, the script now explicitly installs the dependant packages with the correct version.
