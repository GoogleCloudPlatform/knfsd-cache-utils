* [Go Tests](#go-tests)
* [How to run the tests](#how-to-run-the-tests)
  * [Simple example](#simple-example)
  * [Advanced example](#advanced-example)
* [Test Options](#test-options)
  * [Skip terraform destroy](#skip-terraform-destroy)
* [Validations](#validations)
  * [Instance group](#instance-group)
  * [CloudSQL instace](#cloudsql-instace)
  * [FSIDD database](#fsidd-database)
    * [fsidd\_mode=static](#fsidd_modestatic)
    * [fsidd\_mode=external](#fsidd_modeexternal)
* [Costs](#costs)
* [timeouts](#timeouts)
---
# Go Tests
These tests are desgiend to do the following using terratest framework
- Terraform init
- Terraform Apply
- Validations
- Terraform Destroy

# How to run the tests
Example are available under "../examples" directory. Currently simple and advanced examples are available. 

## Simple example
Set the `EXAMPLE_TO_TEST` variable in your shell and run go tests from `test` directory
```shell
export EXAMPLE_TO_TEST=simple
make test
```
## Advanced example
Set the `EXAMPLE_TO_TEST` variable in your shell and run go tests from `test` directory
```shell
export EXAMPLE_TO_TEST=advanced
make test
```
# Test Options
## Skip terraform destroy
you can set SKIP_destroy to true and run the tests, this will skip the terraform destroy phase of the test
```shell
export SKIP_destroy=true
```

# Validations
The following validations are in place 
## Instance group
Verifies that the number of proxy instances as defined in terraform.tfvars have been created
## CloudSQL instace
when the fsidd_mode is set as external, test verifies that the name of the cloudSQL instances matches what is defined and verifies the DBConnection string matches what us expected based on the inputs from terraform.tfvars
## FSIDD database
### fsidd_mode=static
When fsidd is set to static mode (simple example case), tests verify the status of "fsidd" and "knfsd-fsidd" linux services status. 
If both the status matches "inactive" then the validation would succeed. 
### fsidd_mode=external
When the fsidd mode is set to external (advanced example case), test verifies the status of knfsd-fsidd linux services status. 
If the service status is "active" and substate is "running" then the test would succeed. 
This validates that the proxy is able to communicate to the CloudSQL instace

# Costs
- These tests create real resources in a GCP project and then try to clean those resources up at the end of a test run. That means these tests may cost you money to run! 
- Never forcefully shut the tests down (e.g. by hitting CTRL + C) or the cleanup tasks won't run!

# timeouts
We set -timeout 30m on all tests not because they necessarily take that long, but because Go has a default test timeout of 10 minutes, after which it forcefully kills the tests with a SIGQUIT, preventing the cleanup tasks from running. Therefore, we set an overlying long timeout to make sure all tests have enough time to finish and clean up.