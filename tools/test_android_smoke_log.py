#!/usr/bin/env python3
"""Crash attribution regressions using brief logcat records, with no device."""
import unittest
from android_smoke_log import host_errors, PACKAGE

class SmokeLogTests(unittest.TestCase):
    def test_unrelated_google_process_is_not_host_crash(self):
        self.assertEqual(host_errors('''D/GameActivity( 4683): GameActivity_register
E/AndroidRuntime( 3464): FATAL EXCEPTION: main
E/AndroidRuntime( 3464): Process: com.google.android.gms, PID: 3464
E/AndroidRuntime( 3464): java.lang.NoSuchFieldError
'''),[])

    def test_early_java_crash_before_native_startup(self):
        self.assertTrue(host_errors(f'''E/AndroidRuntime( 101): FATAL EXCEPTION: main
E/AndroidRuntime( 101): Process: {PACKAGE}, PID: 101
E/AndroidRuntime( 101): java.lang.UnsatisfiedLinkError
'''))

    def test_native_signal_and_remote_tombstone(self):
        self.assertTrue(host_errors('''D/GameActivity( 101): GameActivity_register
F/libc( 101): Fatal signal 11 (SIGSEGV)
'''))
        self.assertTrue(host_errors(f'F/DEBUG( 202): pid: 101, tid: 105, name: worker >>> {PACKAGE} <<<'))
        self.assertFalse(host_errors('F/DEBUG( 202): pid: 101, tid: 105, name: worker >>> com.google.android.gms <<<'))

    def test_host_errors_and_overlapping_foreign_crash(self):
        self.assertTrue(host_errors('I/PinballFantasies( 101): A1_RENDER_ERROR'))
        self.assertTrue(host_errors('I/PinballFantasies( 101): A2_BOOTSTRAP_REJECTED'))
        self.assertTrue(host_errors(f'''E/AndroidRuntime( 202): FATAL EXCEPTION: main
E/AndroidRuntime( 202): Process: com.google.android.gms, PID: 202
E/AndroidRuntime( 101): FATAL EXCEPTION: main
E/AndroidRuntime( 101): Process: {PACKAGE}, PID: 101
'''))

if __name__=='__main__': unittest.main()
