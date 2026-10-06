package datalayout

import "fmt"

const (
	RetailProfile    = "dos-retail-linked-v1"
	PowerPackProfile = "dos-powerpack-linked-v1"
)

// runtimeProfile confines linked addresses and consumed edition semantics to
// the loader/decode boundary. All profiles use one gameplay engine.
type runtimeProfile struct {
	name         string
	validate     func(string, []byte) error
	canonicalize func(string, []byte) ([]byte, error)
}

var runtimeProfiles = []runtimeProfile{
	{name: RetailProfile, validate: Validate, canonicalize: func(_ string, data []byte) ([]byte, error) { return data, nil }},
	{name: PowerPackProfile, validate: func(name string, data []byte) error {
		if _, ok := profiles[name]; !ok {
			return fmt.Errorf("%s: no supported data layout", name)
		}
		return validateProfile(name, data, linkedProfile(name, true))
	}, canonicalize: translateLinked},
}

var prgNames = []string{"INTRO.PRG", "TABLE1.PRG", "TABLE2.PRG", "TABLE3.PRG", "TABLE4.PRG"}

// DetectInstallation requires one coherent registered profile across all five
// PRGs. TABLE3/4 are shared by A/B; they cannot establish an edition by themselves.
// Required-file reads and role-specific MOD validation remain the host boundary's
// responsibility. No persisted profile identity is needed on adoption/relaunch.
func DetectInstallation(prgs map[string][]byte) (string, error) {
	for _, name := range prgNames {
		if _, ok := prgs[name]; !ok {
			return "", fmt.Errorf("%s: incomplete installation", name)
		}
	}
	var reasons []error
	for _, profile := range runtimeProfiles {
		accepted := true
		for _, name := range prgNames {
			if err := profile.validate(name, prgs[name]); err != nil {
				reasons = append(reasons, err)
				accepted = false
				break
			}
		}
		if accepted {
			return profile.name, nil
		}
	}
	return "", fmt.Errorf("unsupported layout: installation does not match a coherent registered runtime profile: %v", reasons)
}

// PreparePRG supports independent asset tools. Installation callers must first
// detect a coherent profile and use PreparePRGForProfile for every input.
func PreparePRG(name string, data []byte) ([]byte, error) {
	return preparePRG(runtimeProfiles, name, data)
}
func PreparePRGForProfile(profileID, name string, data []byte) ([]byte, error) {
	for _, p := range runtimeProfiles {
		if p.name == profileID {
			return preparePRG([]runtimeProfile{p}, name, data)
		}
	}
	return nil, fmt.Errorf("%s: unsupported layout profile %s", name, profileID)
}
func preparePRG(known []runtimeProfile, name string, data []byte) ([]byte, error) {
	var reasons []error
	for _, profile := range known {
		if err := profile.validate(name, data); err != nil {
			reasons = append(reasons, err)
			continue
		}
		decoded, err := profile.canonicalize(name, data)
		if err != nil {
			return nil, fmt.Errorf("%s: unsupported layout %s: %w", name, profile.name, err)
		}
		if err := ValidateDecoded(name, decoded); err != nil {
			return nil, err
		}
		return decoded, nil
	}
	return nil, fmt.Errorf("%s: unsupported layout: %v", name, reasons)
}
