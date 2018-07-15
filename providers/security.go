//go:generate enumer -type=SecVerb -transform=kebab -trimprefix=SecVerb -json -text -yaml
//go:generate enumer -type=SecSystem -transform=kebab -trimprefix=SecSystem -json -text -yaml

package providers

import "regexp"

// ISecurityProvider defines secutiry provider.
type ISecurityProvider interface {
	GetUser(map[string][]string) (*AuthenticatedUser, error)
}

// SecVerb describes allowed rules for the role.
type SecVerb int

const (
	// SecVerbAll describes all allowed operation rules.
	SecVerbAll SecVerb = iota
	// SecVerbGet describes get operation rule.
	SecVerbGet
	// SecVerbCommand describes execute command rule.
	SecVerbCommand
	// SecVerbHistory describes get history command rule
	SecVerbHistory
)

// SecSystem describes possible role's rule system.
type SecSystem int

const (
	// SecSystemAll describes all possible systems.
	SecSystemAll SecSystem = iota
	// SecSystemDevice describes devices' system.
	SecSystemDevice
)

// SecRoleRule has data, describing single security rule.
type SecRoleRule struct {
	System    string    `yaml:"system" validate:"required,oneof=* device"`
	Resources []string  `yaml:"resources" validate:"unique,min=1"`
	Verbs     []SecVerb `yaml:"verbs" validate:"unique,min=1,oneof=* get command history"`
}

// SecRole has data, describing single security role.
type SecRole struct {
	Name  string        `yaml:"name" validate:"required"`
	Users []string      `yaml:"users" validate:"unique,min=1"`
	Rules []SecRoleRule `yaml:"rules" validate:"min=1"`
}

// BakedRule is a helper type with pre-compiled regexps.
type BakedRule struct {
	Resources []*regexp.Regexp
	System    SecSystem
	Get       bool
	Command   bool
	History   bool
}

// AuthenticatedUser has data with authenticated user, returned by user store.
type AuthenticatedUser struct {
	Username string
	Rules    map[SecSystem][]*BakedRule
}
