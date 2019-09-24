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
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"os"
)

func createRSAPrivateKey(n int) (k *rsa.PrivateKey, err error) {
	if n < 1024 {
		err = fmt.Errorf("Creating a key of size < 1024 is forbidden: %d", n)
		return
	}
	k, err = rsa.GenerateKey(rand.Reader, n)
	return
}

func createKeyFile(file string) (err error) {
	c := 64
	k := make([]byte, c)
	var n int
	n, err = rand.Read(k)
	if err != nil {
		return
	} else if n != c {
		err = fmt.Errorf("crypto/rand read %d of %d bytes", n, c)
		return
	}
	err = ioutil.WriteFile(file, k, os.FileMode(0660))
	return
}
