# Cloud Build should be configured to be able to connect directly to the VM
# using it's internal IP address.
# This requires configuring Cloud Build to use a private pool:
# https://cloud.google.com/build/docs/private-pools/private-pools-overview
omit_external_ip = true
use_internal_ip  = true
use_iap          = false
