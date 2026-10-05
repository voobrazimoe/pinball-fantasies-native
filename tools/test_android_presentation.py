#!/usr/bin/env python3
"""Keep the production wiring connected to the tested display policy."""
from pathlib import Path
import unittest
ROOT = Path(__file__).resolve().parents[1]

class PresentationWiring(unittest.TestCase):
    def test_loop_only_dispatches_events(self):
        source = (ROOT/'hosts/android/app/src/main/cpp/main.cpp').read_text()
        loop = source.split('while (app->destroyRequested == 0)')[1].split('renderer.stopPacing();')[0]
        self.assertIn('presentation::Policy::pollTimeoutMillis()', loop)
        self.assertNotIn('.draw(', loop)
        self.assertNotIn('? 16', source)
        callback = source.split('static void callback64')[1].split('static void callbackLegacy')[0]
        self.assertEqual(callback.count('self->draw('), 1)
        self.assertIn('policy_.consume() && self->active()', callback)
        self.assertIn('self->arm()', callback)

    def test_clock_and_input_remain_independent(self):
        source = (ROOT/'hosts/android/app/src/main/cpp/engine_host.cpp').read_text()
        self.assertIn('pf_engine_advance(persistent, now(),', source)
        self.assertNotIn('vsync', source)
        input_body = source.split('JNI_METHOD(nativeInput)')[1].split('bool androidEngineFrame')[0]
        self.assertIn('pf_engine_set_action(persistent, a, b)', input_body)
        self.assertNotIn('Choreographer', input_body)

if __name__ == '__main__': unittest.main()
