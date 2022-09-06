# MIG Scaler

mig-scaler is a command line tool that submits and manage jobs to scale MIGs.

The jobs are executed by the MIG scaler workflow. This workflow can be deployed using the [MIG scaler Terraform module](./deployment/).

The MIG scaler workflow will scale up a MIG in increments over a period of time to avoid starting too many instances at the same time.

**NOTE:** mig-scaler only scales MIGs up, to scale a MIG back down use the `gcloud compute instance-groups managed resize` command.

## Building mig-scaler

To build mig-scaler you need go 1.17 or later:

```sh
cd tools/mig-scaler
go build
```

Alternatively, if you do not have go locally you can use the `build-docker.sh` script to build the tool using a docker container. The docker container is only used to build mig-scaler.

```sh
cd tools/mig-scaler
./build-docker.sh
```

## Running mig-scaler

For instructions on how to use mig-scaler:

```sh
./mig-scaler help | less
```

For instructions on configuring mig-scaler:

```sh
./mig-scaler help config | less
```

## Configuring MIGs using Terraform

If you use Terraform to deploy your MIGs you should set the `google_compute_instance_group_manager` resource to ignore changes to `target_size` for any MIGs you plan to scale using mig-scaler.

You can also set the initial `target_size` for the MIG to zero, after creating the MIG, use mig-scaler to scale the MIG incrementally.

```terraform
resource "google_compute_instance_group_manager" "example" {
  project            = "my-gcp-project"
  name               = "example"
  base_instance_name = "example"
  zone               = "us-central1-a"

  # Set the initial target_size to zero when creating the MIG
  target_size = 0

  version {
    name              = "v1"
    instance_template = google_compute_instance_template.example.self_link
  }

  lifecycle {
    # Ignore changes to the target_size so that Terraform doesn't scale the MIG
    # back to zero.
    ignore_changes = [
      target_size
    ]
  }
}
```
