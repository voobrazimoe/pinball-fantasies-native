#!/usr/bin/env python3
"""Bounded launch correspondence and explicit native reference boundary."""
import os
from pathlib import Path
import tempfile
import unittest
import audit_10min_demo_canonical_a_jitter as a


class NativeReference(unittest.TestCase):
    def test_native_clock_initialization(self):
        r=a.native_policy()
        self.assertEqual((r['constructor_clock'],r['first_sync_clock'],r['first_spring_low_byte']),(0,1030,6))
        self.assertTrue(r['fresh_factory'])
        self.assertTrue(r['pre_game_counter_history_discarded'])

    def test_native_increment_order(self):
        self.assertIn('before audio',a.native_policy()['advancement'])

    def test_velocity_and_rotation_arithmetic(self):
        self.assertIn('VY=-166*charge-low8(clock)',a.native_policy()['arithmetic'])
        self.assertIn('rotation=low4(clock)',a.native_policy()['arithmetic'])

    def test_historical_phase_explicitly_not_authentic(self):
        r=a.historical_policy()
        self.assertEqual(r['PF8']['exact_DOS_callback_phase'],'UNRESOLVED_NOT_AUTHENTIC')
        self.assertFalse(r['historical_pre_game_counter_phase_required_by_these_tests'])
        self.assertIn('NOT AVAILABLE',r['TestOriginalTrajectories'])

    def mutate(self,path,old,new):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp)
            for directory in ('internal/partyland','internal/physics','internal/frontend'):
                for p in (a.ROOT/directory).glob('*.go'):
                    out=root/directory/p.name;out.parent.mkdir(parents=True,exist_ok=True);out.write_text(p.read_text())
            p=root/path;s=p.read_text();self.assertIn(old,s);p.write_text(s.replace(old,new,1))
            with self.assertRaisesRegex(ValueError,'native'):a.native_policy(root)

    def test_mutation_clock_seed(self):
        self.mutate('internal/partyland/game.go','g := &Game{matrixTimeLeft: true','g := &Game{clock: 128, matrixTimeLeft: true')

    def test_mutation_clock_delta(self):
        self.mutate('internal/partyland/game.go','g.clock += 1030','g.clock += 1031')

    def test_mutation_clock_after_physics(self):
        self.mutate('internal/partyland/game.go','g.clock += 1030\n\tg.audioTick()', 'g.audioTick()\n\tdefer func() { g.clock += 1030 }()')

    def test_mutation_attract_carries_history(self):
        self.mutate('internal/partyland/session.go','PresentationAudioSync() { g.audioTick() }','PresentationAudioSync() { g.clock++; g.audioTick() }')

    def test_mutation_spring_velocity(self):
        self.mutate('internal/physics/ball.go','-166*int16(charge)','-165*int16(charge)')

    def test_mutation_rotation_mask(self):
        self.mutate('internal/physics/ball.go','int16(jitter & 15)','int16(jitter & 31)')

    def test_mutation_reset_clock_each_ball(self):
        self.mutate('internal/partyland/game.go','func (g *Game) resetBall() {',
                    'func (g *Game) resetBall() { g.clock = 0')

    def test_mutation_fresh_score(self):
        self.mutate('internal/partyland/game.go','g := &Game{matrixTimeLeft: true','g := &Game{Score: Number(1), matrixTimeLeft: true')


class PrivateLaunch(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        vals=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(vals):raise unittest.SkipTest('private demo/A/source inputs required')
        cls.paths=list(map(Path,vals));demo,full=a.pinned(*cls.paths)
        cls.b=demo['TABLE1.PRG'];cls.c=full['TABLE1.PRG'];cls.d=a.Decoder();cls.report=a.audit(*cls.paths)

    def test_primary_increment_correspondence(self):
        rows=a.correspondence(self.d,self.c,self.b)
        self.assertEqual(next(r for r in rows if r['role']=='primary increment')['classification'],'RELOCATED-IDENTICAL')

    def test_main_increment_correspondence(self):
        r=next(r for r in self.report['A_demo_linked_jitter_comparison'] if r['role']=='MAIN increment')
        self.assertEqual((r['A_site'],r['demo_site']),(0x39e9,0x39c5))

    def test_demomode_after_primary(self):
        r=next(r for r in self.report['A_demo_linked_jitter_comparison'] if r['role']=='DEMOMODE test after increment')
        self.assertEqual(r['demo_site'],0x455c)

    def test_linked_spring_velocity_formula(self):
        self.d.expect(self.b,768,0x6206,'mov','ax, 0xff5a')
        self.d.expect(self.b,768,0x6210,'and','ax, 0xff')
        self.d.expect(self.b,768,0x6213,'sub','bp, ax')

    def test_linked_rotation_mask(self):
        self.d.expect(self.b,768,0x6233,'and','word ptr [0x2fda], 0xf')

    def test_reset_state_preservation(self):
        r=self.report['surviving_pre_game_state_audit']
        self.assertIsNone(r['another_pre_game_trajectory_residue'])
        self.assertIn('SLUMP_COUNTERN',r['only_surviving_historical_timing_residue'])
        self.assertFalse(r['physical_prefix_join_proved'])
        self.assertTrue({'four totals','player/current ball','relevant lights','tasks/waits','spring charge'} <= {v['role'] for v in r['fields']})

    def test_mutation_demo_specific_jitter_difference(self):
        for at in (0x6207,0x6212,0x6237,0x455a,0x39c7):
            with self.subTest(site=at):
                b=bytearray(self.b);b[at]^=1
                with self.assertRaises(ValueError):a.correspondence(self.d,self.c,b)

    def test_mutation_another_pre_game_score_survives(self):
        # Skip all six score stores: a previous score now survives reset.
        b=bytearray(self.b);b[0x35e:0x360]=b'\x90\x90'
        with self.assertRaises(ValueError):a.correspondence(self.d,self.c,b)
        with self.assertRaises(ValueError):a.selected_state(self.d,self.c,b)

    def test_mutation_other_selected_reset_survives(self):
        for at in (0x384,0x38c,0x3b9,0x3c1,0x331,0x3ae6,0x3b5e,0x3b6b,0x103d,0x5bc7):
            b=bytearray(self.b);b[at]^=1
            with self.assertRaises(ValueError):
                a.correspondence(self.d,self.c,b);a.selected_state(self.d,self.c,b)

    def test_mutation_light_reset_startup_guard(self):
        b=bytearray(self.b);b[0x33e1]=0
        with self.assertRaises(ValueError):a.selected_state(self.d,self.c,b)

    def test_mutation_spring_charge_initial_history(self):
        b=bytearray(self.b);b[0x19db0+0x24a1]=1
        with self.assertRaisesRegex(ValueError,'spring charge'):a.selected_state(self.d,self.c,b)

    def test_historical_phase_not_claimed_proved(self):
        r=self.report
        self.assertFalse(r['historical_phase_theorem_proved'])
        self.assertIsNone(r['reachable_C0_M_P_witness'])
        self.assertEqual(r['ATTRACT_PHASE_ALIGNMENT'],'NOT_PROVED')
        self.assertEqual(r['HISTORICAL_ATTRACT_PHASE'],'NOT_PROVED_BUT_BELOW_NATIVE_REFERENCE_BOUNDARY')
        self.assertEqual(r['reference_boundary']['scope'],'launch-jitter initialization ONLY')

    def test_fresh_entry_without_search_or_drain_upgrade(self):
        r=self.report
        self.assertEqual(r['FRESH_BYGEL_ENTRY_TRANSFER'],'PROVED')
        self.assertIsNone(r['remaining_transfer_reason'])
        self.assertEqual(r['search']['BYGEL_scripts_executed'],0)
        self.assertEqual(r['search']['native_trajectory_calculations'],0)
        self.assertEqual(r['unchanged']['DRAIN_35877_MEMBERSHIP'],'UNKNOWN')

if __name__=='__main__':unittest.main()
