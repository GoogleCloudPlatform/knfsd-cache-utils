#!/bin/sh
# extract the name of the GCP Compute Image that was built by Packer
exec jq -r '.builds[] |
	select(.name == "nfs-proxy" and .builder_type == "googlecompute") |
	.artifact_id' image.manifest.json
