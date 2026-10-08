//go:build demodev

package engine

func (e *Engine) Diagnostic() string { return e.runner.Runtime.Diagnostic() }
