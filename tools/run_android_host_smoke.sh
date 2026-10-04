#!/bin/sh
# Asset-free A1 runtime checks. A fresh log snapshot per phase prevents stale
# portrait/frame messages from making later transitions pass accidentally.
set -eu

APK=${1:-hosts/android/app/build/outputs/apk/debug/app-debug.apk}
PACKAGE=io.github.voobrazimoe.pinballfantasies
COMPONENT="$PACKAGE/.PinballActivity"
scratch=$(mktemp -d)
collector=''
cleanup() {
    if [ -n "$collector" ]; then
        kill "$collector" 2>/dev/null || true
        wait "$collector" 2>/dev/null || true
    fi
    rm -rf "$scratch"
}
trap cleanup EXIT HUP INT TERM

begin_phase() {
    if [ -n "$collector" ]; then
        kill "$collector" 2>/dev/null || true
        wait "$collector" 2>/dev/null || true
    fi
    timeout 10 adb logcat -c
    # Stream from before the transition, so framework startup chatter cannot
    # evict short-lived native/JNI failures before the next polling snapshot.
    adb logcat -v brief -s PinballFantasies:I AndroidRuntime:E GameActivity:V DEBUG:F '*:S' \
        > "$scratch/log" 2>&1 &
    collector=$!
}

snapshot() {
    if grep -Eq 'A1_.*ERROR|A2_BOOTSTRAP_REJECTED|FATAL EXCEPTION|UnsatisfiedLinkError|Fatal signal' "$scratch/log"; then
        cat "$scratch/log" >&2
        exit 1
    fi
}

wait_for() {
    pattern=$1
    attempt=0
    while [ "$attempt" -lt 60 ]; do
        snapshot
        if grep -q "$pattern" "$scratch/log"; then
            cat "$scratch/log"
            return
        fi
        attempt=$((attempt + 1))
        sleep 1
    done
    echo "Android host smoke timed out waiting for: $pattern" >&2
    cat "$scratch/log" >&2
    exit 1
}

require_same_host() {
    current_pid=$(timeout 10 adb shell pidof "$PACKAGE" | tr -d '\r')
    test -n "$current_pid"
    test "$current_pid" = "$initial_pid"
    snapshot
    if grep -q A1_HOST_STARTED "$scratch/log"; then
        echo 'Activity transition restarted the native host' >&2
        exit 1
    fi
}

test -f "$APK"
timeout 60 adb install -r "$APK"
timeout 10 adb shell settings put system accelerometer_rotation 0
timeout 10 adb shell settings put system user_rotation 0
timeout 10 adb shell input keyevent KEYCODE_WAKEUP
timeout 10 adb shell wm dismiss-keyguard
begin_phase
timeout 30 adb shell am start -W -n "$COMPONENT"
wait_for A1_FRAME_PRESENTED
wait_for A2_SHELL_NO_DATA
wait_for 'A1_VIEWPORT.*orientation=portrait'
test "$(grep -c A1_HOST_STARTED "$scratch/log")" -eq 1
initial_pid=$(timeout 10 adb shell pidof "$PACKAGE" | tr -d '\r')
test -n "$initial_pid"

begin_phase
timeout 10 adb shell settings put system user_rotation 1
wait_for 'A1_VIEWPORT.*orientation=landscape'
require_same_host

begin_phase
timeout 10 adb shell settings put system user_rotation 0
wait_for 'A1_VIEWPORT.*orientation=portrait'
require_same_host

begin_phase
timeout 10 adb shell input keyevent KEYCODE_HOME
wait_for A1_ACTIVITY_PAUSED
begin_phase
timeout 30 adb shell am start -W -n "$COMPONENT"
wait_for A1_ACTIVITY_RESUMED
wait_for A1_ACTIVE_FRAME
require_same_host

# A process restart creates fresh EGL resources and presents again.
timeout 10 adb shell am force-stop "$PACKAGE"
begin_phase
timeout 30 adb shell am start -W -n "$COMPONENT"
wait_for A1_FRAME_PRESENTED
wait_for A2_SHELL_NO_DATA
test "$(grep -c A1_HOST_STARTED "$scratch/log")" -eq 1
echo 'OK: Android A1 frame, rotation, background/resume and fresh process'
