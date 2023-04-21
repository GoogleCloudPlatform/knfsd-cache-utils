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

server "basic-attributes" {
    url      = "https://10.0.0.2:8080"
    user     = "nfs-proxy"
    password = file("./netapp-password")
}

server "tls" {
    url      = "https://10.0.0.2:8080"
    user     = "nfs-proxy"
    password = "secret"

    tls {
        ca_certificate    = file("./netapp-ca.pem")
        allow_common_name = true
    }
}

server "empty-tls" {
    url      = "https://10.0.0.2:8080"
    user     = "nfs-proxy"
    password = "secret"
    tls {}
}
