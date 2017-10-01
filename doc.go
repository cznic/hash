// Copyright 2017 The hash Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package hash implements a hash map.
//
// Purpose
//
// Maps provided by this package can be useful when using a key type that is
// not comparable at the language level, like for example a slice or types
// containing slices etc.
//
// Such types are forbidden as keys of the builtin Go maps for good reasons.
// Care must be taken to not modify keys inserted into a Map.
//
// Generic types
//
// Keys and their associated values are interface{} typed, similar to all of
// the containers in the standard library.
//
// Semiautomatic production of a type specific variant of this package is
// supported via
//
//	$ make generic
//
// This command will write to stdout a version of the hash.go file where every
// key type occurrence is replaced by the word 'KEY' and every value type
// occurrence is replaced by the word 'VALUE'. Then you have to replace these
// tokens with your desired type(s), using any technique you're comfortable
// with.
//
// This is how, for example, 'example/int.go' was created:
//
//	$ mkdir example
//	$ make generic | sed -e 's/KEY/*big.Int/g' -e 's/VALUE/*big.Int/g' > example/int.go
//
// After adding import "math/big", no other changes to int.go are necessary, it compiles just fine.
package hash
