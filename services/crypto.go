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

package services

import (
	"crypto/rand"
	"database/sql"
	"fmt"

	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
	"golang.org/x/crypto/bcrypt"
)

// Crypto service provides high level service methods relating to crypto
// operations.
type Crypto struct {
	DB    *sql.DB
	Users *models.Users
}

// Valid determines whether the provided password is valid for the user
// associated with the email address.
func (c *Crypto) Valid(ctx util.Context, email, pass string) (valid bool, err error) {
	var su *models.SensitiveUser
	err = doInTx(ctx, c.DB, func(tx *sql.Tx) error {
		su, err = c.Users.SensitiveUserByEmail(ctx, tx, email)
		return err
	})
	if err != nil {
		return
	}
	valid = passEquals(pass, su.Salt, su.Hashpass)
	return
}

// HashPasswordParameters contains values used in generating secrets.
type HashPasswordParameters struct {
	// Size of the salt in number of bytes.
	SaltSize int
	// Strength of the bcrypt hashing.
	BCryptStrength int
}

// hashPass hashes a password with a salt using the provided parameters and
// bcrypt.
func hashPass(h HashPasswordParameters, secret string) (salt, hashpass []byte, err error) {
	salt, err = newSalt(h.SaltSize)
	if err != nil {
		return
	}
	hashpass, err = hashPasswordWithSalt(secret, salt, h.BCryptStrength)
	return
}

// Uses a password and salt to hash with the given strength value.
//
// Strength is dependent on the bcrypt library, which has built-in protections
// against under-strength values.
func hashPasswordWithSalt(pass string, salt []byte, strength int) (b []byte, err error) {
	salty := append([]byte(pass), salt...)
	b, err = bcrypt.GenerateFromPassword(salty, strength)
	return
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

// Uses time constant comparison to determine if a password and salt are equal
// to the hash.
func passEquals(pass string, salt, hash []byte) bool {
	salty := append([]byte(pass), salt...)
	err := bcrypt.CompareHashAndPassword(hash, salty)
	return err == nil
}
