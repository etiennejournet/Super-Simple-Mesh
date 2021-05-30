#!/bin/sh
set -xe

gomplate --config /etc/envoy/gomplate.yaml
envoy -c /etc/envoy/envoy.yaml
