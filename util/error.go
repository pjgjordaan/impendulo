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
