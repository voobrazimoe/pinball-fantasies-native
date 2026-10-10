package gamepad

import "pinballfantasies/internal/frontend"

type Family int

const (
	Generic Family = iota
	Xbox
	PlayStation
	Nintendo
)

type Glyph int

const (
	DefaultGlyph Glyph = iota
	LetterA
	LetterB
	LetterX
	LetterY
	Cross
	Circle
	Square
	Triangle
)

func (p *Input) Connected() bool                  { return p.connected }
func (p *Input) Family() Family                   { return p.family }
func (p *Input) SetFamily(f Family)               { p.family = f; p.glyphs = [4]Glyph{} }
func (p *Input) SetGlyph(button int, glyph Glyph) { p.glyphs[button] = glyph }
func (p *Input) Glyph(button int) Glyph {
	if p.glyphs[button] != DefaultGlyph {
		return p.glyphs[button]
	}
	if p.family == PlayStation {
		return [...]Glyph{Cross, Circle, Square, Triangle}[button]
	}
	if p.family == Nintendo {
		return [...]Glyph{LetterB, LetterA, LetterY, LetterX}[button]
	}
	return [...]Glyph{LetterA, LetterB, LetterX, LetterY}[button]
}

type Hint struct {
	Button int
	Text   string
}

const (
	DPad = 20
	LT   = 21
	RT   = 22
)

func Hints(mode frontend.Mode, launch, players, demo bool) []Hint {
	switch mode {
	case frontend.Startup:
		return []Hint{{A, "CONTINUE"}}
	case frontend.Selector, frontend.SelectorText:
		rows := []Hint{{A, "PARTY LAND"}}
		if !demo {
			rows = append(rows, Hint{B, "SPEED DEVILS"}, Hint{X, "BILLION DOLLAR GAMESHOW"}, Hint{Y, "STONES 'N BONES"})
		}
		return append(rows, Hint{Start, "OPTIONS"}, Hint{Back, "QUIT"})
	case frontend.Options:
		return []Hint{{DPad, "UP / DOWN"}, {A, "CHANGE VALUE / SELECT"}, {B, "BACK"}}
	case frontend.TableAttract:
		return []Hint{{A, "PLAY - ONE PLAYER"}, {X, "START / ADD PLAYER"}, {Back, "EXIT TABLE"}}
	case frontend.Playing:
		if !launch {
			return nil
		}
		rows := []Hint{{LT, "LEFT FLIPPER"}, {RT, "RIGHT FLIPPER"}, {X, "HOLD / RELEASE TO LAUNCH"}, {Y, "NUDGE / TILT"}}
		if players {
			rows = append(rows, Hint{A, "ADD PLAYER"})
		}
		return append(rows, Hint{Start, "PAUSE"})
	case frontend.Paused:
		return []Hint{{A, "RESUME"}, {B, "EXIT TABLE"}}
	case frontend.QuitQuestion:
		return []Hint{{A, "YES - EXIT TABLE"}, {B, "NO - RETURN"}}
	case frontend.Closing:
		return []Hint{{B, "CLOSE"}}
	}
	return nil
}
