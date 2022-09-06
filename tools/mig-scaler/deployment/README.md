# Terraform MIG Scaler Module

This module deploys a workflow that can be used to scale up MIGs. This workflow
will scale up a MIG in increments over a period of time to avoid starting too many instances at the same time.

**NOTE:** mig-scaler only scales MIGs up, to scale a MIG back down use the `gcloud compute instance-groups managed resize` command.

This module will create:

* A service account for the workflow.
* A custom IAM role to support scaling MIGs.
* Grants the service account permission to scale any MIG within the project.
* Enabled the workflow API.
* The MIG scaler workflow.

This module provides a simplified deployment of the MIG scaler workflow and will configure the workflow service, service account and IAM permissions. This simplified deployment only supports a single workflow per project, and expects the MIGs to be running in the same project.

Jobs can be submitted to the workflow using the [mig-scaler command line tool](../). A single workflow can be used to control multiple MIGs.

Sub-modules are provided for more control over deployment for scenarios such as:

* Controlling MIGs in other projects from a single central workflow.
* Assigning IAM permissions on individual MIGs instead of at the project level.
* Deploying multiple workflows to the same project.

## Usage

```terraform
module "mig_scaler" {
  source          = "github.com/GoogleCloudPlatform/knfsd-cache-utils//tools/mig-scaler/deployment?ref=v0.9.0"
  project         = "my-gcp-project"
  region          = "us-central1"
  service_account = google_service_account.mig_scaler.email
  depends_on = [
    google_project_service.workflows
  ]
}
```

## Inputs

* `project` - (Optional) The ID of the project that the workflow will be created in. Defaults to the provider project configuration.

* `region` - (Optional) The region of the workflow. Defaults to the provider region configuration.

## Requirements

### Next Steps

Build the [mig-scaler command line tool](../). This can be used to submit jobs to the workflow.

Grant users that need to submit jobs the `roles/workflows.invoker` role on the `mig-scaler` workflow.
