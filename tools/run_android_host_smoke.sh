#!/bin/sh
# Asset-free A1/A3 runtime checks. A fresh log snapshot per phase prevents stale
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
    timeout 10 adb shell wm user-rotation >&2 || true
    timeout 10 adb shell settings get system accelerometer_rotation >&2 || true
    timeout 10 adb shell settings get system user_rotation >&2 || true
    timeout 10 adb shell dumpsys window displays >&2 || true
    timeout 10 adb shell dumpsys activity activities >&2 || true
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
timeout 10 adb shell wm user-rotation lock 0
timeout 10 adb shell input keyevent KEYCODE_WAKEUP
timeout 10 adb shell wm dismiss-keyguard
# Public launch must be quiet before the marker-based opt-in smoke.
timeout 10 adb shell am force-stop "$PACKAGE"
begin_phase
timeout 30 adb shell am start -W -n "$COMPONENT"
sleep 3
test -n "$(timeout 10 adb shell pidof "$PACKAGE" | tr -d '\r')"
snapshot
if grep -Eq 'A[123]_(HOST|SURFACE|VIEWPORT|FRAME|ACTIVE|ACTIVITY|NATIVE|INPUT|SHELL|DATA|IMPORT)' "$scratch/log"; then
    echo 'Routine diagnostics emitted by default launch' >&2
    cat "$scratch/log" >&2
    exit 1
fi
timeout 10 adb shell am force-stop "$PACKAGE"
begin_phase
timeout 30 adb shell am start -W -n "$COMPONENT" --ez PF_DIAGNOSTICS true
wait_for A1_FRAME_PRESENTED
wait_for A2_SHELL_NO_DATA
wait_for 'A3_INPUT_STATE resumed=1 focused=1'
wait_for 'A1_VIEWPORT.*orientation=portrait'
test "$(grep -c A1_HOST_STARTED "$scratch/log")" -eq 1
initial_pid=$(timeout 10 adb shell pidof "$PACKAGE" | tr -d '\r')
test -n "$initial_pid"

begin_phase
timeout 10 adb shell wm user-rotation lock 1
wait_for 'A1_VIEWPORT.*orientation=landscape'
require_same_host

begin_phase
timeout 10 adb shell wm user-rotation lock 0
wait_for 'A1_VIEWPORT.*orientation=portrait'
require_same_host

begin_phase
timeout 10 adb shell input keyevent KEYCODE_HOME
wait_for A1_ACTIVITY_PAUSED
begin_phase
timeout 30 adb shell am start -W -n "$COMPONENT" --ez PF_DIAGNOSTICS true
wait_for A1_ACTIVITY_RESUMED
wait_for A1_ACTIVE_FRAME
wait_for 'A3_INPUT_STATE resumed=1 focused=1'
require_same_host

# A3 no-data inputs/Back must keep the same legal native shell alive.
begin_phase
for key in KEYCODE_F1 KEYCODE_F8 KEYCODE_ENTER KEYCODE_DPAD_DOWN KEYCODE_SHIFT_LEFT KEYCODE_SHIFT_RIGHT KEYCODE_SPACE KEYCODE_P KEYCODE_M KEYCODE_BACK; do
    timeout 10 adb shell input keyevent "$key"
done
timeout 10 adb shell input swipe 900 1600 900 1800 150
require_same_host

# A process restart creates fresh EGL resources and presents again.
timeout 10 adb shell am force-stop "$PACKAGE"
begin_phase
timeout 30 adb shell am start -W -n "$COMPONENT" --ez PF_DIAGNOSTICS true
wait_for A1_FRAME_PRESENTED
wait_for A2_SHELL_NO_DATA
wait_for 'A3_INPUT_STATE resumed=1 focused=1'
test "$(grep -c A1_HOST_STARTED "$scratch/log")" -eq 1
echo 'OK: Android A3 no-data inputs/Back and A1 frame, rotation, background/resume and fresh process'
