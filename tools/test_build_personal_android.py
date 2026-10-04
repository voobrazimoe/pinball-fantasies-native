"""Asset-free builder regressions. All payload bytes are invented."""
import contextlib
import hashlib
import io
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
import zipfile
import build_personal_android as builder
from personal_assets import PERSONAL_INPUTS, inventory, ROOT


class BuilderTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        subprocess.run(['git', 'init', '-q', str(self.root)], check=True)
        (self.root/'.gitignore').write_text('/.build-personal/\n/hosts/android/app/.personal-assets/\n/release/personal/\n/hosts/android/app/build/\n')
        tracked = self.root/'hosts/android/app/src/main/source-sentinel.txt'
        tracked.parent.mkdir(parents=True)
        tracked.write_text('invented tracked source')
        subprocess.run(['git', 'add', 'hosts/android/app/src/main/source-sentinel.txt'], cwd=self.root, check=True)
        self.data = self.root/'owner'
        self.data.mkdir()
        for name in PERSONAL_INPUTS: (self.data/name).write_bytes(b'invented invalid test bytes')
        (self.data/'TABLE1.HI').write_bytes(b'ignored dummy scores')
        self.output = self.root/'release/personal/android/pinball-fantasies-personal.apk'
        self.assets = self.root/'hosts/android/app/.personal-assets'
        self.work = self.root/'.build-personal/android-work'
        self.commands = []
        self.real_run = subprocess.run
        self.fail = False
        self.public_contamination = False

    def fake_run(self, command, **kwargs):
        if command[0] == 'git': return self.real_run(command, **kwargs)
        if './cmd/personalvalidate' in command: return subprocess.CompletedProcess(command, 0)
        self.commands.append(command)
        if command[0] == 'sh': return subprocess.CompletedProcess(command, 0)
        private = next((v.split('=', 1)[1] for v in command if v.startswith('-PpersonalBuildDir=')), None)
        if private and self.fail: raise subprocess.CalledProcessError(1, command)
        path = Path(private) if private else self.root/'hosts/android/app/build'
        apk = path/'outputs/apk/debug/app-debug.apk'
        apk.parent.mkdir(parents=True, exist_ok=True)
        with zipfile.ZipFile(apk, 'w') as archive:
            for abi in ('arm64-v8a', 'x86_64'):
                for lib in ('libpfengine.so', 'libpinball_android.so', 'libc++_shared.so', 'liboboe.so'):
                    archive.writestr(f'lib/{abi}/{lib}', b'invented native placeholder')
            # Simulate opt-in packaging while the ignored payload still exists:
            # ordinary Gradle must omit it even before builder cleanup.
            if private or self.public_contamination:
                for file in (self.assets/'personal-data').iterdir():
                    archive.write(file, 'assets/personal-data/'+file.name)
        return subprocess.CompletedProcess(command, 0)

    def build(self):
        with patch.object(builder.subprocess, 'run', side_effect=self.fake_run), \
             patch.object(builder.subprocess, 'check_output', return_value='dummy-source-commit\n'), \
             contextlib.redirect_stdout(io.StringIO()):
            return builder.build(self.data, root=self.root)

    def clean(self):
        self.assertFalse(self.assets.exists())
        self.assertFalse(self.work.exists())
        self.assertEqual((self.data/'INTRO.PRG').read_bytes(), b'invented invalid test bytes')
        self.assertEqual(subprocess.check_output(['git', 'ls-files'], cwd=self.root),
                         b'hosts/android/app/src/main/source-sentinel.txt\n')

    def test_optional_cfg_and_opt_in_and_public(self):
        for cfg in (False, True):
            if cfg: (self.data/'PINBALL.CFG').write_bytes(b'invented settings')
            result = self.build()
            self.assertEqual(len(result['included']), 12 if cfg else 11)
            builder.verify_apk(self.root/'hosts/android/app/build/outputs/apk/debug/app-debug.apk')
            self.assertTrue(result['public_apk_asset_free'])
            self.assertTrue(self.output.is_file())
            builder.ignored(self.root, [self.output])
            self.clean()
        gradle = [c for c in self.commands if c[0] == 'gradle']
        self.assertEqual([any(v.startswith('-PpersonalAssetsDir=') for v in c) for c in gradle], [True, False, True, False])
        self.assertFalse(any(p.name in PERSONAL_INPUTS for p in (self.root/'hosts').rglob('*')))

    def test_failure_cleanup_and_no_stale_output(self):
        self.build()
        self.fail = True
        with self.assertRaises(subprocess.CalledProcessError): self.build()
        self.assertFalse(self.output.exists())
        self.clean()

    def test_missing_rejected_before_copy(self):
        (self.data/'TABLE4.MOD').unlink()
        with self.assertRaises(SystemExit): self.build()
        self.assertEqual(self.commands, [])
        self.assertFalse(self.output.exists())
        self.clean()

    def test_actual_validator_rejects_invented_data(self):
        with self.assertRaises(subprocess.CalledProcessError): inventory(self.data)
        self.assertFalse(self.output.exists())
        self.clean()

    def test_every_commercial_destination_must_be_ignored(self):
        for rule in ('/hosts/android/app/.personal-assets/', '/.build-personal/', '/release/personal/'):
            with self.subTest(rule=rule):
                ignore = self.root/'.gitignore'
                original = ignore.read_text()
                ignore.write_text(original.replace(rule+'\n', ''))
                with self.assertRaises(subprocess.CalledProcessError): self.build()
                self.assertFalse(self.output.exists())
                self.assertEqual(self.commands, [])
                ignore.write_text(original)
                self.clean()

    def test_symlink_destination_rejected(self):
        self.assets.parent.mkdir(parents=True, exist_ok=True)
        self.assets.symlink_to(self.data, target_is_directory=True)
        with self.assertRaises(RuntimeError): self.build()
        self.assertTrue((self.data/'INTRO.PRG').exists())

    def test_public_contamination_rejects_personal_output(self):
        self.public_contamination = True
        with self.assertRaises(RuntimeError): self.build()
        self.assertFalse(self.output.exists())
        self.clean()

    def test_apk_missing_library_extra_scores_or_corrupt_payload_rejected(self):
        self.build()
        records = [{'name': n, 'bytes': len((self.data/n).read_bytes()),
                    'sha256': hashlib.sha256((self.data/n).read_bytes()).hexdigest()} for n in PERSONAL_INPUTS]
        damaged = self.root/'damaged.apk'
        with zipfile.ZipFile(self.output) as original, zipfile.ZipFile(damaged, 'w') as apk:
            for name in original.namelist():
                apk.writestr(name, b'corrupt' if name.endswith('INTRO.PRG') else original.read(name))
        with self.assertRaises(RuntimeError): builder.verify_apk(damaged, records)
        for name, data in [('assets/personal-data/TABLE4.HI', b'invented'),
                           ('assets/unexpected/PINBALL.CFG', b'invented')]:
            with zipfile.ZipFile(self.output, 'a') as apk: apk.writestr(name, data)
            with self.assertRaises(RuntimeError): builder.verify_apk(self.output)
        missing = self.root/'missing.apk'
        with zipfile.ZipFile(missing, 'w'): pass
        with self.assertRaises(RuntimeError): builder.verify_apk(missing)
        records = [{'name': n, 'bytes': 1, 'sha256': hashlib.sha256(b'x').hexdigest()} for n in PERSONAL_INPUTS]
        with self.assertRaises(RuntimeError): builder.verify_apk(self.output, records)

    def test_gradle_opt_in_contract(self):
        config = (ROOT/'hosts/android/app/build.gradle').read_text()
        self.assertIn("if (personalAssets != null) main.assets.srcDirs = [personalAssets]", config)
        self.assertIn('layout.buildDirectory.set(file(personalOutput))', config)
        self.assertFalse((ROOT/'hosts/android/app/src/main/assets/personal-data').exists())


if __name__ == '__main__': unittest.main()
