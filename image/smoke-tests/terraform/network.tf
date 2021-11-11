resource "google_compute_network" "this" {
  project                 = var.project
  name                    = var.prefix
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "this" {
  project = google_compute_network.this.project
  network = google_compute_network.this.id
  region  = var.region
  name    = var.prefix

  ip_cidr_range = "10.0.0.0/20"

  private_ip_google_access = true
}

resource "google_compute_router" "this" {
  name    = var.prefix
  network = google_compute_network.this.id
  region  = var.region
}

resource "google_compute_router_nat" "this" {
  name   = var.prefix
  router = google_compute_router.this.name
  region = google_compute_router.this.region

  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}
