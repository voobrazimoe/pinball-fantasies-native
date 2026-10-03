//go:build !matrixdebug

package frontend

func matrixTestCheat(*Model, Key) bool { return false }
func matrixTestTrace(*Model)           {}
