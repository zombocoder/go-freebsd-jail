//go:build !freebsd

package jail

// Create creates a new jail. Not supported on this platform.
func Create(cfg Config) (JID, error) {
	return 0, ErrNotSupported
}

// CreateOrUpdate creates or updates a jail. Not supported on this platform.
func CreateOrUpdate(cfg Config) (JID, error) {
	return 0, ErrNotSupported
}

// Update modifies parameters of an existing jail. Not supported on this platform.
func Update(nameOrJID string, cfg Config) error {
	return ErrNotSupported
}

// Remove removes a jail. Not supported on this platform.
func Remove(nameOrJID string) error {
	return ErrNotSupported
}

// Get retrieves jail information. Not supported on this platform.
func Get(nameOrJID string) (*Info, error) {
	return nil, ErrNotSupported
}

// List returns all active jails. Not supported on this platform.
func List() ([]Info, error) {
	return nil, ErrNotSupported
}

// Exists checks whether a jail exists. Not supported on this platform.
func Exists(nameOrJID string) (bool, error) {
	return false, ErrNotSupported
}
