# Terraform MIG Scaler Workflow Module

This module deploys a workflow that can be used to scale up MIGs. This workflow
will scale up the MIG in increments over a period of time to avoid starting too many instances at the same time.

**NOTE:** mig-scaler only scales MIGs up, to scale a MIG back down use the `gcloud compute instance-groups managed resize` command.

This module only creates the MIG scaler workflow and is intended for advanced usage. You will need to enable the workflow service and configure a service account with the correct IAM permissions.

Jobs can be submitted to the workflow using the [mig-scaler command line tool](../../../). A single workflow can be used to control multiple MIGs, including MIGs in other projects.

## Usage

```terraform
# Create a service account
resource "google_service_account" "mig_scaler" {
  project      = "my-gcp-project"
  account_id   = "mig-scaler"
  display_name = "Scales up client MIGs"
}

# Enable the workflow service
resource "google_project_service" "workflows" {
  project            = "my-gcp-project"
  service            = "workflows.googleapis.com"
  disable_on_destroy = false
}

# Grant permissions for the service account to scale MIGs
module "iam" {
  source          = "github.com/GoogleCloudPlatform/knfsd-cache-utils//tools/mig-scaler/deployment/modules/iam?ref=v0.9.0"
  project         = "my-gcp-project"
  service_account = google_service_account.mig_scaler.email
}

# Deploy the workflow
module "workflow" {
  source          = "github.com/GoogleCloudPlatform/knfsd-cache-utils//tools/mig-scaler/deployment/modules/workflow?ref=v0.9.0"
  project         = "my-gcp-project"
  region          = "us-central1"
  name            = "mig-scaler"
  service_account = google_service_account.mig_scaler.email
  depends_on = [
    google_project_service.workflows
  ]
}
```

## Inputs

* `project` - (Optional) The ID of the project that the workflow will be created in. Defaults to the provider project configuration.

* `region` - (Optional) The region of the workflow. Defaults to the provider region configuration.

* `name` - (Optional) The name of the workflow. Defaults to `mig-scaler`.

* `service_account` - (Required) The email of the service account for the workflow's identity. This service account needs to have permission to scale the MIGs.

## Requirements

### Next Steps

Build the [mig-scaler command line tool](../../../). This can be used to submit jobs to the workflow.

Grant users that need to submit jobs the `roles/workflows.invoker` role on the workflow.

### Configure a Service Account

For the workflow to be able to scale MIGs you must have a Service Account with the following permissions:

* compute.instanceGroupManagers.get
* compute.instanceGroupManagers.update

The only standard roles that include these permissions are admin roles such as `roles/compute.admin`. It is advised that you create a custom role with only the permissions required.

### Enable APIs

To deploy the workflow you must activate the following API on the project where the workflow is deployed:

* Workflow API - workflows.googleapis.com
