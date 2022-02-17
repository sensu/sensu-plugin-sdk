package sensu

// ErrValidationFailed should be returned when a configuration validation
// function fails.
type ErrValidationFailed string

func (e ErrValidationFailed) Error() string {
	return string(e)
}
