#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
${CXX:-c++} -std=c++17 -pthread -Wall -Wextra -Werror ${PF_A4_TEST_FLAGS:-} \
    -Ihosts/android/app/src/main/cpp hosts/android/tests/pcm_ring_test.cpp -o "$scratch/ring"
"$scratch/ring"
echo 'PASS: A4 ring ordering, stereo alignment, full/empty, wrap, silence, tail-drop, epoch reset, concurrent SPSC and allocation guard'
java_home=$(java -XshowSettings:properties -version 2>&1 | awk '/java.home =/ {print $3}')
case "$(uname -s)" in Darwin) jni_platform=darwin ;; *) jni_platform=linux ;; esac
${CXX:-c++} -std=c++17 -pthread -Wall -Wextra -Werror ${PF_A4_TEST_FLAGS:-} \
    -Ihosts/android/tests/stubs -I"$java_home/include" -I"$java_home/include/$jni_platform" \
    hosts/android/tests/audio_output_test.cpp -o "$scratch/output"
"$scratch/output"
echo 'PASS: A4 production Oboe controller: shared fallback, callback independence, pause/resume, deferred disconnect restart, open failure, quiet diagnostics'
