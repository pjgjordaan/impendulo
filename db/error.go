package db

import (
	"fmt"
)

type (
	//DBGetError represents errors encountered
	//when retrieving data from the db.
	DBGetError struct {
		tipe    string
		err     error
		matcher interface{}
	}

	//DBAddError represents errors encountered
	//when adding data to the db.
	DBAddError struct {
		msg string
		err error
	}

	//DBRemoveError represents errors encountered
	//when removing data from the db.
	DBRemoveError struct {
		tipe    string
		err     error
		matcher interface{}
	}
)

func (this *DBGetError) Error() string {
	return fmt.Sprintf(
		"Encountered error %q when retrieving %q matching %q from db",
		this.err, this.tipe, this.matcher,
	)
}

//Error
func (this *DBAddError) Error() string {
	return fmt.Sprintf(
		"Encountered error %q when adding %q to db",
		this.err, this.msg,
	)
}

//Error
func (this *DBRemoveError) Error() string {
	return fmt.Sprintf(
		"Encountered error %q when removing %q matching %q from db",
		this.err, this.tipe, this.matcher,
	)
}
