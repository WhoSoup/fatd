// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package fat

import (
	"unicode/utf8"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// ValidTokenNameIDs returns true if the nameIDs match the pattern for a valid
// token chain.
func ValidTokenNameIDs(nameIDs []factom.Bytes) bool {
	if len(nameIDs) == 4 && len(nameIDs[1]) > 0 &&
		string(nameIDs[0]) == "token" && string(nameIDs[2]) == "issuer" &&
		factom.ValidIdentityChainID(nameIDs[3]) &&
		utf8.Valid(nameIDs[1]) {
		return true
	}
	return false
}

// NameIDs returns valid NameIDs
func NameIDs(tokenID string, issuerChainID factom.Bytes32) []factom.Bytes {
	return []factom.Bytes{
		[]byte("token"), []byte(tokenID),
		[]byte("issuer"), issuerChainID[:],
	}
}

// ChainID returns the chain ID for a given token ID and issuer Chain ID.
func ChainID(tokenID string, issuerChainID factom.Bytes32) factom.Bytes32 {
	return factom.ChainID(NameIDs(tokenID, issuerChainID))
}
