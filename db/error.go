//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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
