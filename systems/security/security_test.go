package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/providers"
)

func getFakeProvider(usr string) providers.ISecurityProvider {
	ctor := &ConstructSecurityProvider{
		PluginLogger: mocks.FakeNewLogger(nil),
		UserProvider: "test",
		Loader:       mocks.FakeNewPluginLoader(mocks.FakeNewUserStorage(usr)),
		Roles: []*providers.SecRole{
			{
				Name: "1",
				Rules: []providers.SecRoleRule{
					{
						System: providers.SecSystemAll.String(),
						StrVerb: []string{providers.SecVerbAll.String(),
							providers.SecVerbHistory.String(), providers.SecVerbGet.String()},
						Resources: []string{"res"},
					},
				},
				Users: []string{"usr[!0-9]*"},
			},
			{
				Name: "2",
				Rules: []providers.SecRoleRule{
					{
						System: providers.SecSystemDevice.String(),
						StrVerb: []string{providers.SecVerbCommand.String(),
							providers.SecVerbHistory.String(), providers.SecVerbGet.String()},
						Resources: []string{"res*"},
					},
				},
				Users: []string{"usr?"},
			},
			{
				Name: "3",
				Rules: []providers.SecRoleRule{
					{
						System:    "wrong",
						StrVerb:   []string{providers.SecVerbAll.String()},
						Resources: []string{"res1"},
					},
				},
				Users: []string{"user"},
			},
		},
	}

	return NewSecurityProvider(ctor)
}

// Tests that provider falls back to default FS implementation.
func TestFallbackToDefaultProvider(t *testing.T) {
	found := false
	ctor := &ConstructSecurityProvider{
		PluginLogger: mocks.FakeNewLogger(func(s string) {
			if s == "Loading default user storage" {
				found = true
			}
		}),
	}

	NewSecurityProvider(ctor)
	assert.True(t, found)
}

// Tests fallback to default FS provider.
func TestWrongProvider(t *testing.T) {
	found := false
	ctor := &ConstructSecurityProvider{
		PluginLogger: mocks.FakeNewLogger(func(s string) {
			if s == "Failed to load user storage, defaulting to basic" {
				found = true
			}
		}),
		UserProvider: "wrong",
		Loader:       mocks.FakeNewPluginLoader(nil),
	}

	NewSecurityProvider(ctor)
	assert.True(t, found)
}

// Tests possible errors with roles.
func TestWrongRoles(t *testing.T) {
	wrongResRegex := false
	emptyResource := false
	wrongUserRegex := false
	emptyUsers := false
	emptyRules := false
	ctor := &ConstructSecurityProvider{
		PluginLogger: mocks.FakeNewLogger(func(s string) {
			switch s {
			case "Failed to compile role's resource regexp":
				wrongResRegex = true
			case "Skipping role since resources are empty":
				emptyResource = true
			case "Failed to compile role's user regexp":
				wrongUserRegex = true
			case "Skipping role since users are empty":
				emptyUsers = true
			case "Skipping role since rules are empty":
				emptyRules = true
			}
		}),
		UserProvider: "wrong",
		Loader:       mocks.FakeNewPluginLoader(nil),
		Roles: []*providers.SecRole{
			{
				Name: "1",
				Rules: []providers.SecRoleRule{
					{
						System:    providers.SecSystemAll.String(),
						StrVerb:   []string{providers.SecVerbAll.String()},
						Resources: []string{"[!]"},
					},
				},
				Users: []string{"usr"},
			},
			{
				Name: "2",
				Rules: []providers.SecRoleRule{
					{
						System:    providers.SecSystemAll.String(),
						StrVerb:   []string{providers.SecVerbAll.String()},
						Resources: []string{"res\\s"},
					},
				},
				Users: []string{"(("},
			},
			{
				Name: "3",
				Rules: []providers.SecRoleRule{
					{
						System:    "wrong",
						StrVerb:   []string{providers.SecVerbAll.String()},
						Resources: []string{"res*"},
					},
				},
				Users: []string{"[!]"},
			},
		},
	}

	NewSecurityProvider(ctor)
	assert.True(t, wrongResRegex, "roles regexp")
	assert.True(t, emptyResource, "empty resource")
	assert.True(t, wrongUserRegex, "wrong user")
	assert.True(t, emptyUsers, "empty users")
	assert.True(t, emptyRules, "empty rules")
}

// Tests correct user validation.
func TestCorrectUsers(t *testing.T) {
	prov := getFakeProvider("usr1")
	usr, err := prov.GetUser(nil)
	require.NoError(t, err)
	assert.Equal(t, 1, len(usr.(*AuthenticatedUser).Rules))
}

// Tests that incorrect user won't pass validation.
func TestIncorrectUsers(t *testing.T) {
	prov := getFakeProvider("user1")
	usr, err := prov.GetUser(nil)
	require.NoError(t, err)
	assert.Equal(t, 0, len(usr.(*AuthenticatedUser).Rules))
}

// Tests user not found scenario.
func TestUserNotFound(t *testing.T) {
	prov := getFakeProvider("")
	_, err := prov.GetUser(nil)
	assert.Error(t, err)
}

// Tests that everything is allowed.
func TestCorrectProcessing(t *testing.T) {
	ctor := &ConstructSecurityProvider{
		PluginLogger: mocks.FakeNewLogger(nil),
		UserProvider: "test",
		Loader:       mocks.FakeNewPluginLoader(mocks.FakeNewUserStorage("test")),
		Roles: []*providers.SecRole{
			{
				Name: "1",
				Rules: []providers.SecRoleRule{
					{
						System:    "*",
						StrVerb:   []string{"*"},
						Resources: []string{"*"},
					},
				},
				Users: []string{"test"},
			},
		},
	}

	prov := NewSecurityProvider(ctor)
	user, err := prov.GetUser(nil)
	require.NoError(t, err)
	checkAllAllowed(t, user)
	// Trying two times.
	user, err = prov.GetUser(nil)
	require.NoError(t, err)
	checkAllAllowed(t, user)
}
