//go:build dmoimpl1

package partyland

import (
	"errors"
	"os"
	"os/exec"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"
	"reflect"
	"strings"
	"testing"
)

func demoOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func demoWait(n int) []presentation.Command {
	return []presentation.Command{{Op: "_WAIT", Args: []string{"ticks"}, Nums: map[int]int{0: n}}, {Op: "_WAIT", Args: []string{"ticks"}, Nums: map[int]int{0: 32000}}}
}
func TestDemoTimerLifetime(t *testing.T) {
	d := newDemoCore()
	if d.counter != 0 || d.expired || d.holdStill {
		t.Fatal("fresh TABLE1 lifetime")
	}
	demoOK(t, d.calculation(false, false))
	if d.counter != 1 || d.game.Random != 1 {
		t.Fatal("one electronics admission")
	}
	// Explicit post-drain suffix, not a fabricated physical drain witness.
	d.game.Phase = BallLost
	demoOK(t, d.calculation(false, false))
	if d.counter != 2 {
		t.Fatal("BallLost suffix must count")
	}
	before := d.counter
	demoOK(t, d.calculation(true, true))
	if d.counter != before || d.game.Random != 2 {
		t.Fatal("paused update admitted")
	}
	d.game.Phase = NewBall // ordinary gameplay phase change is not a new table lifetime
	demoOK(t, d.calculation(false, false))
	if d.counter != 3 {
		t.Fatal("NEW_BALL phase reset lifetime")
	}
}
func TestDemoEqualityOrderAndBudget(t *testing.T) {
	d := newDemoCore()
	d.counter = 35997
	d.game.Physics.Ball.Hold = false
	d.game.SkillTime = 2
	demoOK(t, d.queue(demoTask{Site: "pending", Delay: 100}))
	demoOK(t, d.calculation(false, false))
	if !d.expired || !d.holdStill || d.game.Physics.Ball.Hold || d.cueRequests != 1 {
		t.Fatal("expiry flags/capture separation")
	}
	if d.game.matrix.op != "_CLEAR4" || d.game.matrix.remaining != 5 {
		t.Fatal("budget advanced cursor")
	}
	if d.game.tasks[0] == nil || d.game.waitCounters["pending"] != 1 {
		t.Fatal("expiry purged tasks")
	}
	if d.game.Audio.Position != 13 || d.game.Audio.Priority != 255 || d.game.Audio.JumpCount != 0 {
		t.Fatal("S_GAMEOVER2")
	}
	want := []string{"early/drain-observed", "UPDATE_COUNTERS", "timer", "electronics-empty", "KEYTASK-empty", "DO_TASKS", "matrix-budget", "late-physics-empty"}
	if !reflect.DeepEqual(d.trace, want) || d.game.SkillTime != 1 {
		t.Fatal(d.trace)
	}
	demoOK(t, d.calculation(false, true))
	if d.game.matrix.remaining != 4 || !d.expired || d.cueRequests != 1 {
		t.Fatal("first matrix visit or sticky expired")
	}
	d.game.Physics.Ball.Hold = true
	demoOK(t, d.queue(demoTask{Site: "release", Action: demoReleaseSourceHold}))
	demoOK(t, d.calculation(false, false))
	if d.holdStill || !d.game.Physics.Ball.Hold {
		t.Fatal("source release changed capture Hold")
	}
}
func TestDemoWrapAndRepeatedEquality(t *testing.T) {
	d := newDemoCore()
	// Exercise every admitted counter value, without assigning a threshold state.
	for i := 1; i <= 101534; i++ {
		demoOK(t, d.calculation(false, false))
		d.trace = nil
		if i == 35997 && d.expired {
			t.Fatal("early equality")
		}
		if i == 35998 && (!d.expired || d.cueRequests != 1) {
			t.Fatal("first equality")
		}
		if i == 65536 && (d.counter != 0 || !d.expired) {
			t.Fatal("uint16 wrap/sticky")
		}
	}
	if d.counter != 35998 || d.cueRequests != 2 || d.game.matrix.remaining != 5 {
		t.Fatal("repeated equality must reinstall")
	}
}
func TestDemoTaskReplacementAndPreservation(t *testing.T) {
	for _, replace := range []bool{false, true} {
		t.Run(map[bool]string{false: "preserve", true: "replace"}[replace], func(t *testing.T) {
			d := newDemoCore()
			d.counter = 35997
			task := demoTask{Site: "interleaving", Action: demoPreserve}
			if replace {
				task.Action = demoReplaceMatrix
				task.Program = demoWait(9)
			}
			demoOK(t, d.queue(task))
			demoOK(t, d.calculation(false, true))
			if !d.expired || d.game.tasks[0] != nil {
				t.Fatal("task did not run after equality")
			}
			if replace {
				if d.game.matrix.op != "_WAIT" || d.game.matrix.remaining != 8 {
					t.Fatal("replacement cursor/budget")
				}
			} else if d.game.matrix.remaining != 4 {
				t.Fatal("preserved expiry cursor")
			}
			demoOK(t, d.calculation(false, false))
			want := uint16(4)
			if replace {
				want = 8
			}
			if d.game.matrix.remaining != want {
				t.Fatal("missed budget")
			}
		})
	}
}
func TestDemoSharedTaskPrimitives(t *testing.T) {
	d := newDemoCore()
	order := []int{}
	// Shared scan executes a child allocated above the current slot in this scan.
	d.game.task(func() bool {
		order = append(order, 0)
		d.game.task(func() bool { order = append(order, 1); return true })
		return true
	})
	d.game.runTasks()
	if !reflect.DeepEqual(order, []int{0, 1}) {
		t.Fatal(order)
	}
	for i := 0; i < 50; i++ {
		if slot := tablelogic.Add(d.game.tasks[:], d.game.taskIDs[:], &d.game.nextTaskID, func() bool { return false }); slot != i {
			t.Fatal("first-free", slot)
		}
	}
	d.game.tasks[7] = nil
	if slot := tablelogic.Add(d.game.tasks[:], d.game.taskIDs[:], &d.game.nextTaskID, func() bool { return false }); slot != 7 {
		t.Fatal("first-free hole", slot)
	}
	d = newDemoCore()
	demoOK(t, d.queue(demoTask{Site: "shared", Delay: 1}))
	demoOK(t, d.queue(demoTask{Site: "shared", Delay: 1}))
	d.game.runTasks()
	if d.game.tasks[0] == nil || d.game.tasks[1] != nil || d.game.waitCounters["shared"] != 0 {
		t.Fatal("shared compare-before-increment/reset")
	}
	demoOK(t, d.queue(demoTask{Site: "reset", Action: demoResetTasks}))
	d.counter = 42
	d.expired = true
	d.game.runTasks()
	if d.counter != 42 || !d.expired || len(d.game.waitCounters) != 0 {
		t.Fatal("task reset corrupted table lifetime")
	}
	for _, f := range d.game.tasks {
		if f != nil {
			t.Fatal("task reset failed")
		}
	}
}
func TestDemoUnsupportedBoundary(t *testing.T) {
	d := newDemoCore()
	d.counter = 35997
	demoOK(t, d.calculation(false, true)) // clear 5 -> 4
	for i := 0; i < 3; i++ {
		demoOK(t, d.calculation(false, true))
	}
	if d.game.matrix.remaining != 1 {
		t.Fatal("clear prefix")
	}
	err := d.calculation(false, true)
	var unsupported *demoUnsupported
	if !errors.As(err, &unsupported) || unsupported.Producer != "NEXT_A" || !strings.Contains(err.Error(), "UNSUPPORTED_DEMO_TRANSITION") || unsupported.Guard == "" || unsupported.Reason == "" {
		t.Fatal(err)
	}
	counter, matrix := d.counter, d.game.matrix
	if d.calculation(false, true) != err || d.counter != counter || !reflect.DeepEqual(matrix, d.game.matrix) {
		t.Fatal("continued after unsupported")
	}
	for _, producer := range []string{"WAIT_FOR_SPIN_TASK/reward", "DO_ELECTRONICS/areas", "KEYTASK/restart", "LOOSE_BALL/bonus", "late-physics"} {
		x := newDemoCore()
		err := x.unsupported(producer, "consumed=true")
		if x.calculation(false, true) != err || x.counter != 0 {
			t.Fatal("unknown producer admitted")
		}
	}
	x := newDemoCore()
	demoOK(t, x.queue(demoTask{Site: "invalid-replacement", Action: demoReplaceMatrix, Program: []presentation.Command{{Op: "_CHANGE_PLAYER"}}}))
	if err := x.calculation(false, false); err == nil || x.game.matrix.active {
		t.Fatal("unsupported replacement fell through to A", err)
	}
}
func TestDemoPriorityAdmission(t *testing.T) {
	a := tablelogic.MusicClock{Priority: 255}
	if !a.Play(tablelogic.JingleSpec{Position: 13, Priority: 255}, 62, timing.Cues) {
		t.Fatal("equal rejected")
	}
	if a.Play(tablelogic.JingleSpec{Position: 7, Priority: 90}, 62, timing.Cues) || a.Position != 13 {
		t.Fatal("lower accepted")
	}
	// No expiry priority clamp: a later validated priority store can admit lower.
	a.Priority = 0
	if !a.Play(tablelogic.JingleSpec{Position: 7, Priority: 90}, 62, timing.Cues) {
		t.Fatal("priority immunity")
	}
}
func TestDemoPrivateCandidate(t *testing.T) {
	if os.Getenv("PF_10MIN_DEMO_DATA") == "" {
		t.Skip("private demo inputs NOT AVAILABLE")
	}
	// Existing identity helper only; no audit CLI, search, replay or exported bytes.
	cmd := exec.Command("python3", "-c", `import os,sys,struct
from pathlib import Path
sys.path.insert(0,'../../tools')
from audit_10min_demo import FILES,digest,require,TABLE_DS
root=Path(os.environ['PF_10MIN_DEMO_DATA'])
for name,(size,identity) in FILES.items():
 b=(root/name).read_bytes()
 require(len(b)==size and digest(b)==identity,'private demo identity mismatch')
b=(root/'TABLE1.PRG').read_bytes()
require(struct.unpack_from('<HB',b,TABLE_DS+0x34cd)==(0,0),'fresh state')
require(struct.unpack_from('<H',b,0x5ce3)[0]==35998,'equality threshold')
require(tuple(b[0x1aa4d:0x1aa50])==(13,0,255),'expiry cue')
# Reviewed 2B2 linked operands, not an instruction-domain research pass.
for site,ds,value,width in ((0x515,0x3026,255,1),(0x51b,0x3481,369,2),(0x52c,0x3481,259,2),(0x532,0x2fdc,15,2),(0x538,0x2fde,47,2),(0x53e,0x3416,0,1),(0x560,0x2fea,0,2),(0x566,0x2fe8,0,2),(0x56c,0x34cb,0,1),(0x572,0x34e1,0,1),(0x578,0x5b1,0,1),(0x57d,0x5b2,0,1),(0x58c,0x34ca,255,1),(0x59c,0x3485,0,1),(0x5ae,0xd1,255,1)):
 require(struct.unpack_from('<H',b,site+2)[0]==ds,'drain store target')
 require(struct.unpack_from('<B' if width==1 else '<H',b,site+4)[0]==value,'drain store value')
for site,value in ((0x5a2,0x1493),(0x5a8,0xc8b),(0x5b3,0x2ba),(0x5ba,30),(0x5bd,0x36c9)):
 require(struct.unpack_from('<H',b,site+1)[0]==value,'drain producer operand')
require(tuple(b[TABLE_DS+0xc8b:TABLE_DS+0xc8e])==(0,0,1),'spring request')
u16=lambda at: struct.unpack_from('<H',b,at)[0]
require(u16(0x1b243)==u16(0x1ba17),'PARTY_ONTS CLEAR4 handler')
require(tuple(u16(0x1b243+2*i) for i in (2,4,6,7,9))==(3,1,7647,336,1),'PARTY_ONTS operands')
require(struct.unpack_from('<H',b,0x592+2)[0]==0x34dd and b[0x592+4]==0,'actual score guard')
# 2B3 bounded consumers: source correspondence already reviewed; no graph search.
require(tuple(u16(0x1b243+2*i) for i in (0,1,3,5,8,10))==(0x4ebe,0x459b,0x2bf8,0x4ba4,0x2c33,0),'party handlers/terminator')
for site,target,value,width in ((0x48a4,0x34e6,0x52d6,2),(0x48aa,0x348c,None,2),(0x48ad,0x348e,None,2),(0x48b0,0x3490,255,2),(0x48b6,0x3492,255,1),(0x2efc,0xd0,255,1),(0x2f06,0x1df1,None,1),(0x2f09,0x34e8,1,2),(0x2f0f,0x34e6,0x52d6,2),(0x2f33,0xd0,255,1),(0x2f3e,0x34e6,0x531d,2),(0x4eaa,0x34e6,0x52d6,2),(0x4eb0,0x457e,13,2),(0x4eb6,0x4580,0xf77c,2),(0x4ebc,0x4582,0x6300,2),(0x4ec2,0x457a,0,2),(0x4ec8,0x457c,0,2),(0x4edc,0x37f9,0x6d1a,2)):
 require(u16(site+2 if value is not None else site+1)==target,'matrix store target')
 if value is not None: require(struct.unpack_from('<B' if width==1 else '<H',b,site+4)[0]==value,'matrix store value')
require(u16(0x2f01+1)==0x3819 and b[0x2f04+1]==0x37,'PLAYER read/encoding')
require(b[0x561d]==0xc3,'PARTYRUT retains SI')
# 2B5: known linked macro operands and suicide, no traversal/research.
for base,record in ((0xfb7,0xc39),(0xfdd,0xc51)):
 require(b[base:base+2]==bytes((0xb7,0)),'SOUNDEFFECT volume zero')
 for delta,op,ptr in ((2,0x0e,record),(6,0x1e,record+1),(10,0x16,record+3)):
  require(b[base+delta:base+delta+2]==bytes((0x8a,op)) and u16(base+delta+2)==ptr,'SOUNDEFFECT record read')
 require(b[base+14:base+20]==bytes((0xfe,0xc2,0xb0,17,0xcd,0x66)),'channel4/INT66 sound consumer')
 at=base+20
 require(b[at]==0xe9 and (at+3+struct.unpack_from('<h',b,at+1)[0])==0x5aba,'source SUICIDE')
for at,sample,note in ((0xc39,23,23),(0xc51,28,18)):
 require(b[TABLE_DS+at]==sample and b[TABLE_DS+at+1]==note and b[TABLE_DS+at+3]==3,'effect operands')
for site,target,value,width in ((0x1003,0x2fdc,297,2),(0x1009,0x2fde,530,2),(0x100f,0x3416,0,1),(0x1031,0x2fea,0,2),(0x1037,0x2fe8,10,2),(0x103d,0x3026,0,1),(0x1043,0x3481,65535,2)):
 require(u16(site+2)==target and struct.unpack_from('<B' if width==1 else '<H',b,site+4)[0]==value,'SETBALL source stores')
require(b[0x1049]==0xe9 and 0x104c+struct.unpack_from('<h',b,0x104a)[0]==0x5aba,'SETBALL SUICIDE')
# 2B4: bounded previously reviewed reset/task operands, no graph/replay search.
for site,value in ((0xf7d,0xcce),(0xf83,0xcf4),(0xf89,0xca8),(0xfa8,5),(0xfab,0x36cf),(0xfce,50),(0xfd1,0x36d1),(0xff4,80),(0xff7,0x36d3)):
 require(u16(site+1)==value,'new-ball child producer/wait')
for site,target,value,width in ((0x5c9,0xd0,255,1),(0xece,0x34ca,0,1),(0xed4,0x2498,0,1),(0xeda,0x34df,255,1),(0xef5,0x3026,255,1),(0xefb,0x2fdc,282,2),(0xf01,0x2fde,530,2),(0xf07,0x3416,0,1),(0xf29,0x2fea,0,2),(0xf2f,0x2fe8,0,2),(0xf35,0x3481,65535,2),(0xf8f,0x34cb,255,1),(0xf95,0x240c,0,1),(0xf9b,0x240e,0,2),(0x3ae1,0x34ca,0,1),(0x3ae6,0x34e1,0,1),(0x3aeb,0x34f6,0,1),(0x3af0,0x34f8,255,1),(0x3ae,0x2330,56,1),(0x4c7,0x1fc,1,1)):
 require(u16(site+2)==target,'reset target')
 require(struct.unpack_from('<B' if width==1 else '<H',b,site+4)[0]==value,'reset value')
require(u16(0x3ab3+1)==0x36c9 and u16(0x3ab6+1)==50,'WAITLIST extent')
require(u16(0x3b45+1)==50,'TASKLIST extent')
require(u16(0x3afe+2)==0xd0 and u16(0x3b10+2)==0x34f1,'matrix guards')



`)
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Fatal("private candidate identity/semantic input validation failed", err)
	}
	d := newDemoCore()
	for i := 0; i < 35998; i++ {
		demoOK(t, d.calculation(false, false))
		d.trace = nil
	}
	if !d.expired || d.cueRequests != 1 {
		t.Fatal("candidate timer")
	}
	t.Log("pinned private candidate: staged timer/equality only; no gameplay witness claimed")
}

func TestDemoCanonicalNewBallLifetime(t *testing.T) {
	d := newDemoCore()
	d.game = newTestGame(t)
	d.counter = 123
	d.expired = true
	d.holdStill = true
	d.game.newBall()
	if d.counter != 123 || !d.expired || !d.holdStill {
		t.Fatal("ordinary NEW_BALL reset table lifetime")
	}
	if d.game.Phase != NewBall {
		t.Fatal("NEW_BALL not executed")
	}
	// This tests lifetime ownership, not demo SETBALL or NEW_BALL conformance.
}

func TestDemoBoundaryRejectsMalformedConsumers(t *testing.T) {
	for _, p := range [][]presentation.Command{
		{{Op: "_WAIT", Args: []string{"100"}}},
		{{Op: "_WAIT", Args: []string{"100"}, Nums: map[int]int{0: 65536}}},
		{{Op: "_CHANGE_PLAYER"}},
	} {
		d := newDemoCore()
		if d.install(p) == nil || d.game.matrix.active {
			t.Fatal("unvalidated consumer admitted")
		}
	}
	d := newDemoCore()
	for i := 0; i < 50; i++ {
		demoOK(t, d.queue(demoTask{Site: "shared", Delay: 100}))
	}
	if err := d.queue(demoTask{Site: "overflow"}); err == nil {
		t.Fatal("slot 51 admitted")
	}
	d = newDemoCore()
	if err := d.queue(demoTask{Site: "reward", Action: demoAction(255)}); err == nil {
		t.Fatal("unknown action admitted")
	}
}
func TestDemoSameScanChildBelowCursorWaits(t *testing.T) {
	d := newDemoCore()
	calls := 0
	d.game.task(func() bool { return true })
	d.game.task(func() bool { d.game.task(func() bool { calls++; return true }); return true })
	d.game.runTasks()
	if calls != 0 || d.game.tasks[0] == nil {
		t.Fatal("child below cursor ran in same scan")
	}
	d.game.runTasks()
	if calls != 1 {
		t.Fatal("child missing next scan")
	}
}
