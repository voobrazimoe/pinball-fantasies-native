package datalayout

import "fmt"

// runtimeProfile is a compiled, reviewed data-layout adapter, never user code.
// A future possessed/validated edition can supply a source-layout validator
// and translate its consumed records to the canonical decoder address space.
// Gameplay, physics and presentation engines keep their existing canonical
// data contract. Pointer/control translations belong to the adapter's audit.
type runtimeProfile struct {
	name         string
	validate     func(string, []byte) error
	canonicalize func(string, []byte) ([]byte, error)
}

var runtimeProfiles = []runtimeProfile{{
	name:         "dos-retail-linked-v1",
	validate:     Validate,
	canonicalize: func(_ string, data []byte) ([]byte, error) { return data, nil },
}}

// PreparePRG is the common host boundary. Only the possessed canonical layout
// is registered today; this does not claim compatibility with unknown editions.
func PreparePRG(name string, data []byte) ([]byte, error) {
	return preparePRG(runtimeProfiles, name, data)
}
func preparePRG(known []runtimeProfile, name string, data []byte) ([]byte, error) {
	var reasons []error
	for _, profile := range known {
		if err := profile.validate(name, data); err != nil {
			reasons = append(reasons, err)
			continue
		}
		canonical, err := profile.canonicalize(name, data)
		if err != nil {
			return nil, fmt.Errorf("%s: unsupported layout %s: %w", name, profile.name, err)
		}
		// Adapters must produce records satisfying the canonical engine semantics.
		if err := Validate(name, canonical); err != nil {
			return nil, err
		}
		return canonical, nil
	}
	return nil, fmt.Errorf("%s: unsupported data layout: %v", name, reasons)
}
