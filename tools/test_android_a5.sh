#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
javac -d "$scratch" hosts/android/app/src/main/java/io/github/voobrazimoe/pinballfantasies/AudioPolicy.java hosts/android/tests/AudioPolicyTest.java
java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.AudioPolicyTest
# A3 now also exercises actual native focus/route coordination and stale tokens.
sh tools/test_android_a3.sh
sh tools/test_android_a4.sh
