#!/usr/bin/env python3
"""Reuse portable artifact storage/dependency checks with bundled personal data."""
from pathlib import Path
root = Path(__file__).resolve().parent.parent
source = (root/'tools/smoke_pf12_portable.py').read_text()
source = source.replace("root/'analysis/pf12-validation/portable'", "root/'.build-personal/storage'")
source = source.replace('        for p in originals:\n            shutil.copy2(p, folder/p.name)\n', '')
source = source.replace("release/linux/", "release/personal/linux/").replace("release/windows/", "release/personal/windows/")
source = source.replace('        for p in originals:\n            assert hashlib.sha256((folder/p.name).read_bytes()).hexdigest() == before[p.name], p.name\n', '')
source = source.replace('game = folder/binary.name', "game = folder/('Фантазии 日本 '+binary.name)")
source = source.replace('default_external_data', 'default_bundled_data')
exec(compile(source, str(root/'tools/smoke_pf12_portable.py'), 'exec'))
