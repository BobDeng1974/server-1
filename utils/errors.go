package utils

// ErrInitPanic defines a panic during the Init call.
type ErrInitPanic struct {
}

// Error formats output.
func (*ErrInitPanic) Error() string {
	return "panic in init"
}

// ErrNoEntryPoint defines an absent Load method in a plugin.
type ErrNoEntryPoint struct {
}

// Error formats output.
func (*ErrNoEntryPoint) Error() string {
	return "entry point not found"
}

// ErrWrongSignature defines that Load method has wrong params.
type ErrWrongSignature struct {
}

// Error formats output.
func (*ErrWrongSignature) Error() string {
	return "wrong entry point signature"
}

// ErrWrongInterface defines unexpected interface implemented by plugin.
type ErrWrongInterface struct {
}

// Error formats output.
func (*ErrWrongInterface) Error() string {
	return "requested interface is not implemented"
}

// ErrWrongSettingsSignature defines unexpected interface implemented by plugin's settings object.
type ErrWrongSettingsSignature struct {
}

// Error formats output.
func (*ErrWrongSettingsSignature) Error() string {
	return "wrong settings signature"
}

// ErrInvalidConfig defines wrong configuration error.
type ErrInvalidConfig struct {
}

// Error formats output.
func (*ErrInvalidConfig) Error() string {
	return "config validation error"
}

// ErrNoInit defines an absent Init method in a plugin interface.
type ErrNoInit struct {
}

// Error formats output.
func (*ErrNoInit) Error() string {
	return "init method not found"
}

// ErrDownload defines download error.
type ErrDownload struct {
}

// Error formats output.
func (*ErrDownload) Error() string {
	return "proxy download failed"
}
