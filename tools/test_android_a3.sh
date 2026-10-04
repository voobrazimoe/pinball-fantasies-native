#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
javac -d "$scratch" hosts/android/app/src/main/java/io/github/voobrazimoe/pinballfantasies/Controls.java hosts/android/tests/ControlsTest.java
java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.ControlsTest
${CXX:-c++} -std=c++17 -Wall -Wextra -Werror -Ihosts/android/app/src/main/cpp hosts/android/tests/frame_test.cpp -o "$scratch/frame-test"
"$scratch/frame-test"
echo 'PASS: A3 frame validation, stride copy, allocation reuse, aspect and orientation'
java_home=$(java -XshowSettings:properties -version 2>&1 | awk '/java.home =/ {print $3}')
case "$(uname -s)" in Darwin) jni_platform=darwin ;; *) jni_platform=linux ;; esac
${CXX:-c++} -std=c++17 -pthread -Wall -Wextra -Werror \
    -Ihosts/android/tests/stubs -Icmd/pfengine \
    -I"$java_home/include" -I"$java_home/include/$jni_platform" \
    hosts/android/tests/session_test.cpp -o "$scratch/session-test"
"$scratch/session-test"
echo 'PASS: A3 actual native session: engine/input/frame/lifecycle serialization, focus, close and stale token guards'
