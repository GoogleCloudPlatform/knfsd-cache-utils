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
  instance_name                   = "packer-nfs-proxy-{{uuid}}"
  machine_type                    = var.machine_type
  disk_size                       = 20
  disk_type                       = "pd-ssd"
  disable_default_service_account = true

  # Networking
  network_project_id = var.network_project
  subnetwork         = var.subnetwork
  omit_external_ip   = var.omit_external_ip

  # Output image
  image_family            = var.image_family
  image_name              = local.image_name
  image_description       = "NFS caching proxy server"
  image_storage_locations = var.image_storage_location == "" ? null : [var.image_storage_location]

  # Source image
  source_image            = "ubuntu-2004-focal-v20220615"
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
  }

  provisioner "shell" {
    execute_command = "chmod +x {{ .Path }}; {{ .Vars }} sudo {{ .Path }}"
    inline = [
      "chmod +x ./scripts/*.sh",
      "./scripts/1_build_image.sh",
      "reboot",
    ]
    expect_disconnect = true
    pause_after = "30s"
  }

  provisioner "shell" {
    execute_command = "chmod +x {{ .Path }}; {{ .Vars }} sudo {{ .Path }}"
    inline = ["./scripts/8_custom.sh"]
  }

  provisioner "shell" {
    execute_command = "chmod +x {{ .Path }}; {{ .Vars }} sudo {{ .Path }}"
    inline = [
      "printf 'Kernel Version: %s\n' \"$(uname -r)\"",
      "./scripts/9_finalize.sh",
    ]
    expect_disconnect = true
    skip_clean        = true
  }
}
