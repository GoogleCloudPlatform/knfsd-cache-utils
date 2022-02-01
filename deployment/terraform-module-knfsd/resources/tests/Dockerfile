FROM bats/bats:v1.4.1

RUN apk add --no-cache git && \
    git config --global advice.detachedHead false && \
    cd /opt/bats/lib && \
    git clone --depth 1 --branch v0.3.0 https://github.com/bats-core/bats-support.git && \
    git clone --depth 1 --branch v2.0.0 https://github.com/bats-core/bats-assert.git
