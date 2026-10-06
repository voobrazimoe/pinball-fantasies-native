package stones

import (
	"bytes"
	"os"
	"path/filepath"
	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/physics"
	"reflect"
	"testing"
)

func TestPrivateDeluxeAreaSemantics(t *testing.T) {
	var games []*Game
	var inputs [][]byte
	for _, env := range []string{"PF_RUNTIME_DATA", "PF_DELUXE_CD_DATA"} {
		dir := os.Getenv(env)
		if dir == "" {
			t.Skip("supply " + env)
		}
		raw, e := os.ReadFile(filepath.Join(dir, "TABLE4.PRG"))
		if e != nil {
			t.Fatal(e)
		}
		data, e := datalayout.PreparePRG("TABLE4.PRG", raw)
		if e != nil {
			t.Fatal(e)
		}
		table, e := physics.DecodeStones(data)
		if e != nil {
			t.Fatal(e)
		}
		games = append(games, New(table, data))
		inputs = append(inputs, data)
	}
	if !reflect.DeepEqual(games[0].areas, games[1].areas) {
		t.Fatal("typed rectangles/handlers differ")
	}
	count := 0
	roles := map[string]bool{}
	for _, list := range games[1].areas {
		for _, r := range list {
			count++
			roles[r.Handler] = true
		}
	}
	if count != 44 || len(roles) != 34 {
		t.Fatal(count, len(roles))
	}
	// Exercise every typed native handler with identical initial model state.
	// The comparison includes capture queues, gates, collision and control state.
	for role := range roles {
		t.Run(role, func(t *testing.T) {
			var states []*Game
			for i, g := range games {
				copy := New(g.Physics.Table, inputs[i])
				copy.area(role)
				states = append(states, copy)
			}
			// Edition bytes/art are storage, while handlers operate on shared state.
			if !bytes.Equal(states[0].Frame().Pix, states[1].Frame().Pix) {
				t.Fatal("handler presentation differs")
			}
			for _, g := range states {
				g.Physics.OnEvent = nil
				g.Physics.BeforeLate = nil
				g.Physics.BeforeTargets = nil
				g.Physics.AfterTargets = nil
				g.Physics.ScrollForce = nil
			}
			if !reflect.DeepEqual(states[0].Physics, states[1].Physics) {
				t.Fatal("handler collision/capture/gate state differs")
			}
			for _, g := range states {
				g.tasks = [20]func() bool{}
			}
			states[0].Display = nil
			states[1].Display = nil
			states[0].Physics = nil
			states[1].Physics = nil
			if !reflect.DeepEqual(states[0], states[1]) {
				t.Fatal("handler control semantics differ")
			}
		})
	}
}
