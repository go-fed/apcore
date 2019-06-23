// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2019 Cory Slep
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package apcore

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Uses a password and salt to hash with the given strength value.
//
// Strength is dependent on the bcrypt library, which has built-in protections
// against under-strength values.
func hashPasswordWithSalt(pass string, salt []byte, strength int) (b []byte, err error) {
	salty := append([]byte(pass), salt...)
	b, err = bcrypt.GenerateFromPassword(salty, strength)
	return
}

// Uses time constant comparison to determine if a password and salt are equal
// to the hash.
func passEquals(pass string, salt, hash []byte) bool {
	salty := append([]byte(pass), salt...)
	err := bcrypt.CompareHashAndPassword(hash, salty)
	return err == nil
}

// Creates a new salt of the given byte size.
//
// The smallest supported salt length is 16 bytes, any shorter request will be
// 16 bytes long.
func newSalt(size int) (b []byte, err error) {
	if size < 16 {
		size = 16
	}
	b = make([]byte, size)
	var n int
	n, err = rand.Read(b)
	if err != nil {
		return
	} else if n != size {
		err = fmt.Errorf("salt generation: crypto/rand only read %d of %d bytes", n, size)
		return
	}
	return
}
