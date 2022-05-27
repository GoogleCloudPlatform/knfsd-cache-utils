# Knfsd Proxy Metrics

This modules configures Google Cloud Monitoring [Metric Descriptors](https://cloud.google.com/monitoring/custom-metrics/creating-metrics#creating_a_metric_descriptor) and Monitoring Dashboard.

**This module only needs to be applied once per Google Cloud Project, regardless of the number of Knfsd deployments you have in the project**.

## Usage

```terraform
module "metrics" {
    source  = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/metrics?ref=v0.6.1"
    project = "my-gcp-project"
}
```

## Inputs

| Variable | Description | Required | Default
| -------- | ----------- | -------- | -------
| project  | The Google Cloud Project that the Knfsd Metrics are being deployed to. If it is not provided, the provider project is used. | False

### Provider Default Values

The Terraform module also supports supplying the project, region and zone using provider default values. Set the project, region, and/or zone properties on the Google Terraform provider. Omit these properties from the module.

```terraform
provider "google" {
  project = "my-gcp-project
}

module "metrics" {
    source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/metrics?ref=v0.6.1"
}
```

## Filtering

If you have deployed multiple knfsd proxy clusters within the same project then the dashboard will show all the knfsd proxy clusters.

To filter by a specific knfsd proxy cluster add the following filters to the dashboard with the value `{PROXY_BASENAME}-proxy-group`.

* System Metadata Label, `instance_group`
* Resource, `instance_group_name`
* Resource, `backend_name`

For example, if the knfsd proxy cluster was deployed with `PROXY_BASENAME = "example"` then the filter value is `example-proxy-group`.

![Example Filters](./filters.png)

## Upgrading from earlier versions to use the module

There are three ways to upgrade from the earlier deployments to the new module based approach:

1. Continue deploying using the old method
2. Re-create metric descriptors
3. Refactor existing state

### Continue deploying using the old method

This module can still be deployed using the original method. Clone the module using git and add a `provider.tf` file.

### Re-create metric descriptors

Use `terraform destroy` on the old state, before running `terraform apply` on the new Terraform configuration.

**WARNING:** This might also delete any historical data collected for these metrics.

**NOTE:** You need to stop any knfsd proxy instances (and clients using custom knfsd metrics) first. Otherwise the new Terraform may fail to apply.

### Refactor existing state

You will need to copy your Terraform state from the old metrics deployment to the new Terraform configuration.

If you're using Terraform v1.1 or greater you can use the built in [refactoring support](https://www.terraform.io/language/modules/develop/refactoring).

Copy [refactor.tf.example](./refactor.tf.example) into your Terraform configuration and rename it to `refactor.tf`. This assumes you named your module "metrics" (same as the example usage). If you have used a different name for your module replace "module.metrics" with your module name, e.g. "module.knfsd_metrics".

If you're using a version earlier than v1.1 you can use the [`terraform state mv`](https://www.terraform.io/cli/commands/state/mv) to rename the resources.

## Caveats

### Dashboard shows unrelated load balancers and instance groups

Metrics based on resources that do not support labels such as the knfsd proxy latency (load balancer) and knfsd proxy cluster size (instance group) are filtered based upon the instance group name ending with the suffix `-proxy-group`.

If you have other instance groups that end with the suffix `-proxy-group` these instances groups will also be included in some of the graphs.
