package util

import (
	"fmt"
)

type (
	//MissingError indicates that a key was not present in a map.
	MissingError struct {
		key string
	}

	//CastError indicates that an interface{} could not be cast to
	//a certain type.
	CastError struct {
		tipe  string
		value interface{}
	}

	//IOError is used to add more context to errors which occur in the util package.
	UtilError struct {
		origin interface{}
		tipe   string
		err    error
	}
)

func (this *MissingError) Error() string {
	return fmt.Sprintf("Error reading value for %q.", this.key)
}

func (this *CastError) Error() string {
	return fmt.Sprintf("Error casting value %q to %q.", this.value, this.tipe)
}

func IsCastError(err error) (ok bool) {
	_, ok = err.(*CastError)
	return
}

//Error
func (this *UtilError) Error() string {
	return fmt.Sprintf(`Encountered error %q while %s %q.`,
		this.err, this.tipe, this.origin)
}
