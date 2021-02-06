package bot

// ErrUnregisteredUnauthorizedUser represents failure in authorization of unregistered user
type ErrUnregisteredUnauthorizedUser struct {
	UserID string
}

func newErrUnregisteredUnauthorizedUser(UserID string) *ErrUnregisteredUnauthorizedUser {
	return &ErrUnregisteredUnauthorizedUser{
		UserID: UserID,
	}
}
func (e *ErrUnregisteredUnauthorizedUser) Error() string {
	return "User must be registered as unauthorized before finalizing the authorization"
}

// ErrFilteredOut represtents failure in passing the filter
type ErrFilteredOut struct {
	UserID string
}

func newErrFilteredOut(UserID string) *ErrFilteredOut {
	return &ErrFilteredOut{
		UserID: UserID,
	}
}
func (e *ErrFilteredOut) Error() string {
	return "The given user was filtered based on his usos information"
}
