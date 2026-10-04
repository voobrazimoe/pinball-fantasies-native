#!/bin/sh
# Physical acceptance only. Python supplies portable timeouts (including macOS).
set -eu
exec python3 "$(dirname "$0")/android_a6_device.py" "$@"
