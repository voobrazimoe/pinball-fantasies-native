#!/usr/bin/env python3
"""2B7: rerun saved fixed inputs only; preserve all historical artifacts.

Use the accepted research source base in a private snapshot, since the working
checkout also contains later candidate changes. Only fixed replay functions are run, including the historically named
TestResearchSetballSearch, whose body uses the saved Release35764 script.
"""
import argparse
import io
import json
import os
import shutil
from pathlib import Path
import subprocess
import tarfile
import audit_10min_demo_post_collision as post
import audit_10min_demo_party_on as party
import audit_10min_demo_setball as setball
import audit_10min_demo_droptask2 as drop
import audit_10min_demo_duringflash as flash
import audit_10min_demo_deterministic_replay as replay
from audit_10min_demo_graph import pinned, require
from audit_10min_demo_programs import identities

ROOT = replay.ROOT


def run(harness, test, env):
    p = subprocess.run([str(ROOT/'.tools/go/bin/go'), 'test', './internal/partyland',
                        '-run', '^'+test+'$', '-count=1', '-v'],
                       cwd=harness, env=env, capture_output=True, text=True)
    (harness/(test+'.log')).write_text(p.stdout+p.stderr)
    require(p.returncode == 0, 'fixed replay failed: '+str(harness/test))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--resume', action='store_true', help='reuse this correction run only; never historical outputs')
    parser.add_argument('--output', type=Path, default=Path('/private/tmp/pf-dmo-impl2b7'))
    args = parser.parse_args()
    out = args.output
    out.mkdir(exist_ok=args.resume)
    data, canonical, historical = [Path(os.environ[n]) for n in
        ('PF_10MIN_DEMO_DATA', 'PF_RUNTIME_DATA', 'PF_DMO0_HISTORICAL_SOURCE')]
    demo, _ = pinned(data, canonical, historical)
    h, _ = identities(demo['TABLE1.PRG'], historical)
    cfg, _ = post.configuration(demo['TABLE1.PRG'], h)
    # Keep Git/source base checks in replay.prepare intact. The snapshot is
    # read only input to each separate generated research harness.
    snapshot = out/'source-base'
    if not args.resume:
        snapshot.mkdir()
        archive = subprocess.check_output(['git', 'archive', replay.BASE], cwd=ROOT)
        with tarfile.open(fileobj=io.BytesIO(archive)) as tar:
            tar.extractall(snapshot)
        tracked = subprocess.check_output(['git', 'ls-files'], cwd=ROOT, text=True).splitlines()
        for name in tracked:
            if name.endswith('.go') and not (snapshot/name).exists():
                require(name.endswith('_test.go'), 'unexpected later production source')
                (snapshot/name).parent.mkdir(parents=True, exist_ok=True)
                shutil.copyfile(ROOT/name, snapshot/name)
        (snapshot/'.git').symlink_to(ROOT/'.git', target_is_directory=True)
    replay.ROOT = snapshot
    summary = {'timing_oracle_basis': 'FF-inclusive-v1', 'source_base': replay.BASE, 'configuration': cfg, 'runs': {}}
    for name, module, test, variable, filename in [
        ('post', post, 'TestResearchPostCollision', 'PF_POST_OUTPUT', 'post-output.json'),
        ('party', party, 'TestResearchPartyFixedCorrection', 'PF_PARTY_OUTPUT', 'party-output.json'),
        ('setball', setball, 'TestResearchSetballSearch', 'PF_SETBALL_OUTPUT', 'setball-output.json'),
        ('drop', drop, 'TestResearchDropWitness', 'PF_DROP_OUTPUT', 'drop-output.json'),
        ('flash', flash, 'TestResearchFlashWitness', 'PF_FLASH_OUTPUT', 'flash-output.json')]:
        harness = out/name
        reuse = args.resume and (harness/filename).exists()
        if reuse:
            require(json.loads((harness/'post-config.json').read_text()) == cfg, 'resume configuration differs')
            identity = json.loads((harness/'identity.json').read_text())
        else:
            harness.mkdir(exist_ok=True); (harness/'internal/partyland').mkdir(parents=True, exist_ok=True)
            identity = module.prepare(harness, canonical, cfg)
            (harness/'identity.json').write_text(json.dumps(identity))
        if name == 'party' and not reuse:
            p = harness/'internal/partyland/dmo0_party_on_test.go'
            p.write_text(p.read_text()+'''
func TestResearchPartyFixedCorrection(t *testing.T){researchSetup()
 a,b:=partyFresh(t,35842,0,0,true,""),partyFresh(t,35842,0,0,true,"")
 if !reflect.DeepEqual(a,b){t.Fatal("determinism")}
 raw,_:=json.Marshal(map[string]any{"witness":a,"deterministic_equal":true})
 if err:=os.WriteFile(os.Getenv("PF_PARTY_OUTPUT"),raw,0600);err!=nil{t.Fatal(err)}
}
''')
        env = os.environ.copy()
        env.update(PF_POST_CONFIG=str(harness/'post-config.json'),
                   PF_POST_PAUSE_OUTPUT=str(harness/'pause-output.json'),
                   GOCACHE='/private/tmp/pf-dmo0-go-cache')
        env[variable] = str(harness/filename)
        if name in ('drop', 'flash'):
            script = harness/'fixed-script.json'
            script.write_text(json.dumps(module.TARGET))
            env['PF_DROP_SCRIPT' if name == 'drop' else 'PF_FLASH_SCRIPT'] = str(script)
        if not reuse:
            run(harness, test, env)
        raw = json.loads((harness/filename).read_text())
        if name == 'post':
            if not reuse:
                run(harness, 'TestResearchPostCollisionPauseCycle', env)
            summary['pause_cycle'] = json.loads((harness/'pause-output.json').read_text())
            traces = {k: v for k, v in raw.items() if not k.startswith('mutation')}
        else:
            traces = {'witness': raw['witness']}
        saved_name = {'post': 'post-collision-termination', 'party': 'first-equality-party-on',
                      'setball': 'first-equality-setball', 'drop': 'first-equality-droptask2',
                      'flash': 'first-equality-duringflash'}[name]
        saved = json.loads(Path('/private/tmp/pf-dmo0-'+saved_name+'.json').read_text())
        if name == 'post':
            old_traces = saved['continuations']
        else:
            old_traces = {'witness': saved['replay' if name == 'party' else 'fresh_input_replay']['witness']}
        result = {}
        for key, trace in traces.items():
            final = trace['final']
            require(final['QUIT'], name+' concrete continuation did not reach QUIT')
            require(final['expiry_visits'] == (1048 if name == 'flash' else 1050), 'observed admitted visits')
            # Compare every saved prefix/equality row and boundary field. Extra
            # phase/admission observers are excluded from this historical check.
            matched = {}
            for field in ('rows', 'boundaries'):
                if field not in old_traces[key]:
                    continue
                old_rows = [x for x in old_traces[key][field] if x['calculation'] <= 35998]
                new_rows = [x for x in trace[field] if x['calculation'] <= 35998]
                require(len(old_rows) == len(new_rows), 'saved prefix row count')
                for old, new in zip(old_rows, new_rows):
                    if name == 'drop':
                        new = dict(new, linked_HOLDSTILL=new['expired'] and new['ball']['Hold'])
                    require(all(new.get(k) == v for k, v in old.items() if k != 'linked_matrix'), 'saved prefix/equality state differs: '+name)
                matched[field] = len(old_rows)
            # Observe command transitions, do not calculate a deadline to drive Sync.
            rows = trace['rows']
            transitions = []
            prior = None
            for row in rows:
                if row.get('boundary') or row['calculation'] < 35998:
                    continue
                state = (row['matrix_pc'], row['matrix_op'])
                if state != prior:
                    transitions.append({k: row[k] for k in ('calculation', 'matrix_pc', 'matrix_op', 'matrix_text_left', 'matrix_remaining', 'expiry_visits', 'scroll_phase')})
                    prior = state
            result[key] = dict(QUIT=final['calculation'], admitted_visits=final['expiry_visits'], prefix_equality_matched=matched, transitions=transitions,
                               final=final, source=str(harness/filename))
        summary['runs'][name] = result
        (out/'summary.json').write_text(json.dumps(summary, separators=(',', ':'))+'\n')
        print(name, {k: v['QUIT'] for k, v in result.items()}, flush=True)


if __name__ == '__main__':
    main()
