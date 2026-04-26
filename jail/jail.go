//go:build freebsd

package jail

import "github.com/zombocoder/go-freebsd-jail/internal/jailops"

// Create creates a new jail with the given configuration.
// Returns the kernel-assigned JID on success.
// Fails with ErrExists if a jail with the same name already exists.
//
// Required fields: Name, Path.
// Requires root privileges.
func Create(cfg Config) (JID, error) {
	jid, err := jailops.Create(cfg)
	if err != nil {
		return 0, err
	}
	return JID(jid), nil
}

// CreateOrUpdate creates a jail if it doesn't exist, or updates it if it does.
// Returns the JID on success.
//
// Required fields: Name, Path.
// Requires root privileges.
func CreateOrUpdate(cfg Config) (JID, error) {
	jid, err := jailops.CreateOrUpdate(cfg)
	if err != nil {
		return 0, err
	}
	return JID(jid), nil
}

// Update modifies parameters of an existing jail identified by name or JID string.
// Requires root privileges.
func Update(nameOrJID string, cfg Config) error {
	return jailops.Update(nameOrJID, cfg)
}

// Remove removes a jail by name or JID string.
// Idempotent: removing a non-existent jail returns nil.
//
// WARNING: This kills all processes in the jail and removes child jails.
// This is a destructive, irreversible operation.
// Requires root privileges.
func Remove(nameOrJID string) error {
	return jailops.Remove(nameOrJID)
}

// Get retrieves information about a jail by name or JID string.
// Returns ErrNotFound if the jail does not exist.
func Get(nameOrJID string) (*Info, error) {
	return jailops.Get(nameOrJID)
}

// List returns all active jails on the system.
func List() ([]Info, error) {
	return jailops.List()
}

// Exists checks whether a jail exists by name or JID string.
func Exists(nameOrJID string) (bool, error) {
	return jailops.Exists(nameOrJID)
}
