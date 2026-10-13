package frontend

import (
	"image"

	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"pinballfantasies/internal/tablelogic"
)

// Session is the minimum the model drives. Everything else the frontend asks
// of a table is an optional capability below: tests may supply a bare
// Session, so the model probes each one. The shipped tables implement table,
// which the compiler checks at the bottom of this file.
type (
	matrixOwner        interface{ MatrixDisplay() *presentation.Display }
	cheatRuleSetter    interface{ SetCheatRules(bool, bool, int) }
	playerStarter      interface{ StartPlayers(count int) }
	playerCounter      interface{ PlayerCount() int }
	playerSelector     interface{ PlayerSelectionReady() bool }
	playerScorer       interface{ PlayerScores() []tablelogic.Decimal }
	chuteReporter      interface{ InChute() bool }
	musicToggler       interface{ ToggleMusic() }
	musicReporter      interface{ MusicDisabled() bool }
	scoreEntryReporter interface{ ScoreEntryPending() bool }
	scoreEntryFinisher interface{ FinishScoreEntry() }
	presentationAudio  interface{ PresentationAudioSync() }
	frameReuser        interface{ FrameInto(*image.RGBA) *image.RGBA }
	attractRenderer    interface{ AttractFrame(counter int) *image.RGBA }
	playfieldOwner     interface{ PlayfieldPhysics() *physics.Game }
	attractCuer        interface{ AttractCue() string }
	loadedStateCarrier interface{ CarryLoadedTableState(previous any) }
	demoSession        interface{ DemoFinished() bool }
)

// table is the contract every shipped table fulfils.
type table interface {
	Session
	matrixOwner
	cheatRuleSetter
	playerStarter
	playerCounter
	playerSelector
	playerScorer
	chuteReporter
	musicToggler
	musicReporter
	scoreEntryReporter
	scoreEntryFinisher
	presentationAudio
	frameReuser
	attractRenderer
	playfieldOwner
}

var (
	_ table = (*partyland.Game)(nil)
	_ table = (*speeddevils.Game)(nil)
	_ table = (*gameshow.Game)(nil)
	_ table = (*stones.Game)(nil)

	// Party Land's attract keeps the default S_MAIN cue.
	_ attractCuer = (*speeddevils.Game)(nil)
	_ attractCuer = (*gameshow.Game)(nil)
	_ attractCuer = (*stones.Game)(nil)

	_ loadedStateCarrier = (*gameshow.Game)(nil)
	_ demoSession        = (*partyland.Game)(nil)
)
