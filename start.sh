#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")"
if [ ! -x ./bin/pulsenet-linux-amd64 ]; then
  printf 'PulseNet binary not found. Building it now...\n'
  ./build.sh
fi
exec ./bin/pulsenet-linux-amd64 "$@"
