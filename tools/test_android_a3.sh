#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
javac -d "$scratch" hosts/android/app/src/main/java/io/github/voobrazimoe/pinballfantasies/InteractionGeometry.java hosts/android/app/src/main/java/io/github/voobrazimoe/pinballfantasies/Controls.java hosts/android/app/src/main/java/io/github/voobrazimoe/pinballfantasies/SemanticUi.java hosts/android/tests/ControlsTest.java hosts/android/tests/GeometryTest.java hosts/android/tests/SemanticUiTest.java
java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.GeometryTest
java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.SemanticUiTest
${CC:-cc} -std=c11 -Wall -Wextra -Werror -Ihosts/macos hosts/macos/host_logic.c hosts/android/tests/keyboard_parity.c -o "$scratch/keyboard-parity"
"$scratch/keyboard-parity" > "$scratch/desktop-makes"
java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.ControlsTest "$scratch/desktop-makes"
${CXX:-c++} -std=c++17 -Wall -Wextra -Werror -Ihosts/android/app/src/main/cpp hosts/android/tests/frame_test.cpp -o "$scratch/frame-test"
"$scratch/frame-test"
echo 'PASS: A3 frame validation, stride copy, allocation reuse, aspect and orientation'
java_home=$(java -XshowSettings:properties -version 2>&1 | awk '/java.home =/ {print $3}')
case "$(uname -s)" in Darwin) jni_platform=darwin ;; *) jni_platform=linux ;; esac
${CXX:-c++} -std=c++17 -pthread ${PF_A3_TEST_FLAGS:-} -Wall -Wextra -Werror \
    -Ihosts/android/tests/stubs -Icmd/pfengine \
    -I"$java_home/include" -I"$java_home/include/$jni_platform" \
    hosts/android/tests/session_test.cpp -o "$scratch/session-test"
"$scratch/session-test"
echo 'PASS: A3 actual native session: engine/input/frame/lifecycle serialization, focus, close and stale token guards'

${CXX:-c++} -std=c++17 -Wall -Wextra -Werror -Ihosts/android/tests/stubs -Ihosts/android/app/src/main/cpp hosts/android/tests/presentation_test.cpp -o "$scratch/presentation-test"
"$scratch/presentation-test"
python3 tools/test_android_presentation.py
echo 'PASS: display callback eligibility, stale epochs, blocking Looper and diagnostics'
