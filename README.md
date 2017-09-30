# hash

Package hash implements a hash map.

Installation:

    $ go get github.com/cznic/hash

Documentation: [godoc.org/github.com/cznic/hash](http://godoc.org/github.com/cznic/hash)

# Purpose

Map provided by this package can be useful when using a key type that is not comparable at the language level, like for example a slice.

# Generic types

Keys and their associated values are interface{} typed, similar to all of the containers in the standard library.

Semiautomatic production of a type specific variant of this package is supported via

     $ make generic

This command will write to stdout a version of the hmap.go file where every key type occurrence is replaced by the word 'KEY' and every value type occurrence is replaced by the word 'VALUE'. Then you have to replace these tokens with your desired type(s), using any technique you're comfortable with.

This is how, for example, 'example/int.go' was created:

     $ mkdir example
     $ make generic | sed -e 's/KEY/int64/g' -e 's/VALUE/int64/g' > example/int.go

No other changes to int.go are necessary, it compiles just fine.
