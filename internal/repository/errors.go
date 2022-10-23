package repository

type ErrUnknown struct {
	Err error
}

func (e ErrUnknown) Error() string {
	return e.Err.Error()
}

func (e ErrUnknown) Unwrap() error {
	return e.Err
}

type ErrNotificationNotFound struct {
	Message string
}

func (e ErrNotificationNotFound) Error() string {
	return e.Message
}

type ErrInvalidData struct {
	Message string
}

func (e ErrInvalidData) Error() string {
	return e.Message
}
