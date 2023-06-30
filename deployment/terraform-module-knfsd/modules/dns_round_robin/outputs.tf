/*
 * Copyright 2020 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

output "dns_name" {
  # Even when setting var.dns_name explicitly this is still useful as it
  # allows other resources to depend upon the A records having been created.
  value      = google_dns_managed_zone.proxy.dns_name

  # Wait until the A records have been created in the zone so that the DNS is
  # fully initialized before Terraform updates dependent resources.
  # Declaring a dependency instead of using "google_dns_record_set.proxy[0].dns_name"
  # directly as the value because the record set might not always exist.
  depends_on = [
    google_dns_record_set.proxy
  ]
}
