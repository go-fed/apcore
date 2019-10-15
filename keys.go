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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	minKeySize = 1024
)

func createRSAPrivateKey(n int) (k *rsa.PrivateKey, err error) {
	if n < minKeySize {
		err = fmt.Errorf("Creating a key of size < %d is forbidden: %d", minKeySize, n)
		return
	}
	k, err = rsa.GenerateKey(rand.Reader, n)
	return
}

func marshalPublicKey(p crypto.PublicKey) (string, error) {
	pkix, err := x509.MarshalPKIXPublicKey(p)
	if err != nil {
		return "", err
	}
	pb := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkix,
	})
	return string(pb), nil
}

func serializeRSAPrivateKey(k *rsa.PrivateKey) ([]byte, error) {
	return x509.MarshalPKCS8PrivateKey(k)
}

func deserializeRSAPrivateKey(b []byte) (crypto.PrivateKey, error) {
	return x509.ParsePKCS8PrivateKey(b)
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
