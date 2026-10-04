# Host diagnostics

Public Windows, Linux, macOS and Android launches are quiet by default. Routine
host, audio, lifecycle and mode/table messages require explicit opt-in. Fatal
failures and user-relevant errors remain enabled, including Windows fatal dialogs,
macOS alerts and Android import error status.

Set `PF_DIAGNOSTICS=1` in the process environment before launching a desktop host.
Console builds write routine diagnostics to stdout. Windows/Linux GUI release
builds then create `native.log` in their state directory (`-config-dir` if supplied).
No diagnostic file is created during a default launch. macOS writes enabled routine
messages to the console through NSLog.

Android accepts the environment variable when supplied to its process, or the
boolean Activity launch extra:

```sh
adb shell am force-stop io.github.voobrazimoe.pinballfantasies
adb shell am start -W -n io.github.voobrazimoe.pinballfantasies/.PinballActivity --ez PF_DIAGNOSTICS true
```

The extra is read before GameActivity starts its native thread. Android host smoke
first checks a default launch for routine markers, then explicitly enables them
for its A1/A2/A3 assertions. Packaged import instrumentation explicitly enables
native diagnostics before opening a session. Desktop marker-based smoke scripts
also set `PF_DIAGNOSTICS=1`.

Specialized logs remain independently opt-in: `PF12_PACING_LOG`,
`PF12_TRANSITION_LOG`, `PF12_INPUT_LOG=1` (optional `PF12_INPUT_LOG_PATH`), and macOS
`PF_PACING_LOG` / `--pacing-log`. Explicit developer/test tools retain their own
requested output.
