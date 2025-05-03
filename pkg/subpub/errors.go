package subpub

import "errors"

var (
	ErrInputData       = errors.New("error of the input params: restricted value was got")
	ErrSystemCondition = errors.New("error of the system's condition: the call is restricted")
)
