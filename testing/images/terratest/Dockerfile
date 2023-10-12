# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM gcr.io/cloud-builders/gcloud

# the build requires the following tools:
#  - gcloud (from the base image, latest version is fine).
#  - go v1.20.
#  - terraform 1.2.

# install golang (and unzip which we need later for terraform)
RUN apt-get update \
	&& apt-get install -y golang-1.20 unzip \
	&& apt-get clean

# add go to the path
ENV PATH="$PATH:/usr/lib/go-1.20/bin"

# grab terraform from hashicorp and upack
RUN curl https://releases.hashicorp.com/terraform/1.2.0/terraform_1.2.0_linux_amd64.zip --output terraform.zip \
	&& unzip terraform.zip -d /usr/local/bin \
	&& rm terraform.zip
