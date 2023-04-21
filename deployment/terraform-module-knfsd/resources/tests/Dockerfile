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

FROM bats/bats:v1.4.1

RUN apk add --no-cache git && \
    git config --global advice.detachedHead false && \
    cd /opt/bats/lib && \
    git clone --depth 1 --branch v0.3.0 https://github.com/bats-core/bats-support.git && \
    git clone --depth 1 --branch v2.0.0 https://github.com/bats-core/bats-assert.git
