package store

import "fmt"

type ErrDataNotFound struct {
	Err error
}

func (e *ErrDataNotFound) Error() string {
	return fmt.Sprintf("Data not found, err %v", e.Err)
}
