//go:build !windows || !personal

package platform

// Public builds and Linux (whose AppRun supplies the mounted data path) embed
// no original game data. The personal Windows builder supplies this variable.
var personalPayload []byte
