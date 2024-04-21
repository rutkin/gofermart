package my_errors

import "errors"

var ErrExists = errors.New("exists")
var ErrNotFound = errors.New("not found")
