# Terraform MIG Scaler IAM Module

This module creates a custom role for the MIG scaler workflow and grants the
service account permission to scale MIGs at the project level. This will allow
the MIG scaler workflow to scale any MIG in the project.

## Usage

```terraform
# Create a service account
resource "google_service_account" "mig_scaler" {
  project      = "my-gcp-project"
  account_id   = "mig-scaler"
  display_name = "Scales up client MIGs"
}

# Grant permissions for the service account to scale MIGs
module "iam" {
  source          = "github.com/GoogleCloudPlatform/knfsd-cache-utils//tools/mig-scaler/deployment/modules/iam?ref=v1.0.0"
  project         = "my-gcp-project"
  service_account = google_service_account.mig_scaler.email
}
```

## Inputs

* `project` - (Optional) The ID of the project that the workflow will be created in. Defaults to the provider project configuration.

* `service_account` - (Required) The email of the service account for the workflow's identity. This service account needs to have permission to scale the MIGs.
