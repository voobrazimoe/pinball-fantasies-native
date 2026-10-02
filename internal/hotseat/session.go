// Package hotseat implements FANTASIE's eight player slots and round rotation.
// T is the table's executed PLAYER_STRUC save boundary, never a game engine.
package hotseat

import "pinballfantasies/internal/tablelogic"

const MaxPlayers = 8

type Session[T any] struct {
	SelectionOpen              bool
	PlayerCount, CurrentPlayer int // DOS-facing, one based
	Players                    [MaxPlayers]T
	remaining                  uint8
}

func (s *Session[T]) Initialize(count int, initial T) {
	if count < 1 {
		count = 1
	}
	if count > MaxPlayers {
		count = MaxPlayers
	}
	*s = Session[T]{PlayerCount: count, CurrentPlayer: 1, SelectionOpen: true}
	for i := range s.Players {
		s.Players[i] = initial
	}
}
func (s *Session[T]) Select(count int) {
	if count >= 1 && count <= MaxPlayers {
		s.PlayerCount = count
	}
}
func (s *Session[T]) Save(state T) { s.Players[s.CurrentPlayer-1] = state }
func (s *Session[T]) Load() T      { return s.Players[s.CurrentPlayer-1] }

// Advance increments the ball round only after the last player. On exhaustion
// DOS leaves PLAYER on the last player, with BALLS[11] one beyond the limit.
func (s *Session[T]) Advance(ball *uint8, limit uint8) bool {
	s.SelectionOpen = false
	if s.CurrentPlayer < s.PlayerCount {
		s.CurrentPlayer++
		return true
	}
	*ball++
	if *ball > limit {
		return false
	}
	s.CurrentPlayer = 1
	return true
}
func (s *Session[T]) Scores(score func(T) tablelogic.Decimal, live tablelogic.Decimal) []tablelogic.Decimal {
	out := make([]tablelogic.Decimal, s.PlayerCount)
	for i := range out {
		out[i] = score(s.Players[i])
	}
	out[s.CurrentPlayer-1] = live
	return out
}

// StartMatch consumes each qualifying player's final tens digit once, in player
// order. Ordinary earned extra balls during a matched turn retain that player.
func (s *Session[T]) StartMatch(digit uint8, score func(T) tablelogic.Decimal) bool {
	s.remaining = 0
	for i := 0; i < s.PlayerCount; i++ {
		if score(s.Players[i])[10] == digit {
			s.remaining |= 1 << i
		}
	}
	return s.NextMatch()
}
func (s *Session[T]) NextMatch() bool {
	for i := 0; i < s.PlayerCount; i++ {
		if s.remaining&(1<<i) != 0 {
			s.remaining &^= 1 << i
			s.CurrentPlayer = i + 1
			return true
		}
	}
	return false
}
