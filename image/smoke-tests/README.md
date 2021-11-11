# Smoke Tests

## Prerequisites

You will need the following software on your local machine:

* [Terraform](https://www.terraform.io/)
* [Docker](https://www.docker.com/)
* [GCloud SDK](https://cloud.google.com/sdk/docs/install)
* [Bash](https://www.gnu.org/software/bash/)
* (optional) [Git](https://github.com/git-guides/install-git)

## Download

To run the smoke tests you need to download them to your local machine.

Click the "Code" dropdown in the top right for the various options. If you do not have git installed, there is an option in the dropdown to download the repository as a ZIP file.

## Deploy

The Terraform will create an isolated KNFSD proxy for testing, including an independent VPC network so as not to conflict with any other machines. The intent is that the resources will be created to run the smoke tests, then destroyed once the smoke test is complete.

The names of all the resources created have a configurable prefix (defaults to `smoke-test`). This can be changed to avoid conflicting with any existing resources.

**NOTE: The Terraform state is stored on your local machine.** Do not remove this state until you have finished the tests so that you can easily destroy the resources that were created.

Create a `terraform.tfvars` file in `smoke-tests/terraform`.

```terraform
project     = "my-project"
region      = "us-central1"
zone        = "us-central1-a"
proxy_image = "knfsd-image"
prefix      = "smoke-tests"
```

Apply the Terraform configuration:

```bash
cd smoke-tests/terraform
terraform init -upgrade
terraform apply
```

## Create BATS image

The smoke tests use the [Bash Automated Testing System (BATS)](https://github.com/bats-core/bats-core) to run a suite of tests on the the NFS proxy.

The simplest way to run BATS is to use a docker image. The standard BATS docker image does not contain openssh, which is required by these tests.

Build the custom BATS docker image:

```bash
cd smoke-tests
docker build -t bats:knfsd-test bats
```

## Run the tests

Create a `run.sh` file in the `smoke-tests/tests` directory to set the parameters for the smoke tests.

```bash
# Customize these to match your environment
export PROJECT=my-project
export ZONE=us-central1-a
export PREFIX=smoke-tests

# These do not need to be changed unless you're modified the test environment
export PROXY_INSTANCE_GROUP="${PREFIX}-proxy-group"
export CLIENT_INSTANCE="${PREFIX}-client"
export NFS_SOURCE="$(terraform -chdir=../terraform output -raw source_ip)":/files
export NFS_PROXY="$(terraform -chdir=../terraform output -raw proxy_ip)":/files
export TEST_PATH=/test
export BATS_IMAGE=bats:knfsd-test

exec ./smoke.sh
```

Run the test script:

```bash
bash run.sh
```

You should seen an output such as:

```text
$ bash run.sh
Resolved PROXY_INSTANCE as smoke-tests-proxy-wgl2
(ssh) client: External IP address was not found; defaulting to using IAP tunneling.
(ssh) proxy: External IP address was not found; defaulting to using IAP tunneling.
(ssh) client: Warning: Permanently added 'compute.3791846025470988987' (ECDSA) to the list of known hosts.
(ssh) proxy: Warning: Permanently added 'compute.2834520282166514666' (ECDSA) to the list of known hosts.
Mounting NFS shares on client
client: debconf: unable to initialize frontend: Dialog
client: debconf: (Dialog frontend will not work on a dumb terminal, an emacs shell buffer, or without a controlling terminal.)
client: debconf: falling back to frontend: Readline
...
```

Once the initial setup has finished the test suite will begin:

```text
Setup Complete
───────────────────────────────────────────────────────────────────────
Running Tests
 ✓ Proxy is running the correct kernel
 ✓ cachefilesd is running
 ✓ fscache is mounted on a separate volume
 ✓ Client can read via proxy
 ✓ Client can write via proxy
 ✓ Metadata caches positive lookups
 ✓ Metadata caches negative lookups
 ✓ Proxy caches file data

8 tests, 0 failures
```

**NOTE: Tests marked with "(slow)" can take a couple of minutes to run.** Though all tests should complete in less than 5 minutes, once the test suite has started.

## Clean up

Once you have finished running tests you can use Terraform to destroy all the resources that were created.

```bash
cd smoke-tests/terraform
terraform destroy
```

## Troubleshooting

### SSH still running in the background

If the tests terminate abnormally they might not close the SSH connections. If the control files (`.proxy.ssh`, or `.client.ssh`) are left behind you can list all SSH processes in the current session by running:

```bash
pgrep -s0 -ax ssh
```

To terminate all the SSH processes run:

```bash
pkill -s0 -x ssh
```
