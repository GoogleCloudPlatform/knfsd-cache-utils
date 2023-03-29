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

# The Internal DNS name of the Internal Load Balancer
output "dns_name" {
  description = "The internal DNS name of the load balancer."
  value       = "${google_compute_forwarding_rule.default.service_label}.${var.PROXY_BASENAME}.il4.${var.REGION}.lb.${var.PROJECT}.internal"
}
