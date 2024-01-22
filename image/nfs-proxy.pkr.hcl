// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

packer {
  # Minimum of packer 1.7 is required to support downloading newer plugin versions.
  # Packer 1.8.1 comes pre-bundled with a supported version of the googlecompute plugin.
  required_version = ">= 1.7"

  required_plugins {
    googlecompute = {
      # Need a minimum of 1.0.13 to support the rsa-ssh2-512 key algorithm.
      # The older ssh-rsa (rsa + sha1) algorithm has been removed from the newer
      # versions of OpenSSL as it is no longer secure.
      version = ">= 1.0.13"
      source  = "github.com/hashicorp/googlecompute"
    }
  }
}

locals {
  timestamp = formatdate("YYYY-MM-DD-hhmmss", timestamp())
  image_name = (
    var.image_name == "" ?
    "${var.image_family}-${local.timestamp}" :
    var.image_name
  )
}

source "googlecompute" "nfs-proxy" {
  project_id = var.project
  zone       = var.zone

  # Build machine
  instance_name                   = var.build_instance_name
  machine_type                    = var.machine_type
  disk_size                       = 50
  disk_type                       = "pd-ssd"
  disable_default_service_account = true

  # Networking
  network_project_id = var.network_project
  subnetwork         = var.subnetwork
  omit_external_ip   = var.omit_external_ip
  tags               = var.network_tags

  # Output image
  image_family            = var.image_family
  image_name              = local.image_name
  image_description       = "NFS caching proxy server"
  image_storage_locations = var.image_storage_location == "" ? null : [var.image_storage_location]

  # Source image
  source_image            = "ubuntu-2204-jammy-v20221101a"
  source_image_project_id = ["ubuntu-os-cloud"]

  # Communicator
  communicator    = "ssh"
  ssh_username    = "build"
  use_iap         = var.use_iap
  use_internal_ip = var.use_internal_ip

  skip_create_image = var.skip_create_image
}

build {
  sources = ["googlecompute.nfs-proxy"]

  provisioner "file" {
    source      = "${path.root}/resources/"
    destination = "./"
    timeout     = "10m"
  }

  provisioner "shell" {
    execute_command = "chmod +x {{ .Path }}; {{ .Vars }} sudo {{ .Path }}"
    inline = [
      "chmod +x ./scripts/*.sh",
      "./scripts/1_build_image.sh",
      "reboot",
    ]
    expect_disconnect = true
    pause_after       = "30s"
    timeout           = "1h"
  }

  provisioner "shell" {
    execute_command = "chmod +x {{ .Path }}; {{ .Vars }} sudo {{ .Path }}"
    inline          = ["./scripts/8_custom.sh"]
  }

  provisioner "shell" {
    execute_command = "chmod +x {{ .Path }}; {{ .Vars }} sudo {{ .Path }}"
    inline = [
      "printf 'Kernel Version: %s\n' \"$(uname -r)\"",
      "./scripts/9_finalize.sh",
    ]
    expect_disconnect = true
    skip_clean        = true
    timeout           = "5m"
  }

  # Output the last build to a manifest file in the current directory.
  # This can be useful for automated tooling, especially if packer generated
  # the image name with a timestamp.
  post-processor "manifest" {
    output = "image.manifest.json"
  }
}
