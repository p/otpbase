// Derived from goathtool which includes the following copyright notice:
//
// Copyright 2015 Reed O'Brien <reed@reedobrien.com>.
// All rights reserved. Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.
//
// Copyright (c) 2015 Reed O'Brien. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Reed O'Brien, nor the names of any
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const step = 30
const digits = 6

func secret_to_key(secret string) ([]byte, error) {
	secret = strings.ToUpper(secret)
	secret = strings.Replace(secret, " ", "", -1)
	// repad base 32 strings if they are short.
	for len(secret) < 32 && len(secret) > 16 {
		secret = secret + "="
	}
	key, err := base32.StdEncoding.DecodeString(secret)
	return key, err
}

func gen_hotp(key []byte, counter int64) (string, error) {
	var code uint32

	hash := hmac.New(sha1.New, key)

	err := binary.Write(hash, binary.BigEndian, counter)
	if err != nil {
		return "", err
	}

	h := hash.Sum(nil)
	offset := h[19] & 0x0f

	trunc := binary.BigEndian.Uint32(h[offset : offset+4])
	trunc &= 0x7fffffff
	code = trunc % uint32(math.Pow(10, float64(digits)))
	passcodeFormat := "%0" + strconv.Itoa(digits) + "d"

	return fmt.Sprintf(passcodeFormat, code), nil
}

func gen_totp(key []byte) (string, error) {
	var code string
	now := time.Now().UTC().Unix()
	counter := now / step
	code, err := gen_hotp(key, counter)
	return code, err
}
