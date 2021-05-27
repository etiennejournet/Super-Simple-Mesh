#!/bin/sh
set -xe

GOMPLATE_VERSION="v3.9.0"

wget -O /usr/local/bin/gomplate https://github.com/hairyhenderson/gomplate/releases/download/${GOMPLATE_VERSION}/gomplate_linux-amd64-slim
chmod +x /usr/local/bin/gomplate

chown -R envoy /etc/envoy
