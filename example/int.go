// Copyright 2017 The hash Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	"math/big"

	"github.com/cznic/mathutil"
)

const threshold = 2

type item struct {
	k *big.Int
	v *big.Int
}

// Cursor provides enumerating of Map items.
type Cursor struct {
	K        *big.Int
	V        *big.Int
	hasMoved bool
	i        int
	j        int
	m        *Map
}

// Next moves the cursor to the next item in the map and sets the K and V
// fields accordingly. It returns true on success, or false if there is no next
// item.
//
// Every use of the K/V fields, even the first one, must be preceded by a call
// to Next, for example
//
//	for c := m.Cursor(); c.Next(); {
//		... c.K, c.V valid here
//	}
func (c *Cursor) Next() bool {
	if c.m == nil {
		return false
	}

	if c.hasMoved {
		c.j++
	}
	c.hasMoved = true

next:
	if c.i >= len(c.m.items) {
		c.m = nil
		return false
	}

	b := c.m.items[c.i]
	if c.j < len(b) {
		c.K = b[c.j].k
		c.V = b[c.j].v
		return true
	}

	c.j = 0
	c.i++
	goto next
}

// Map is a hash table.
type Map struct {
	bits  uint
	eq    func(a, b *big.Int) bool
	hash  func(*big.Int) int64
	items [][]item
	l     uint
	len   int
	mask  uint
	mask2 uint
	n     uint
	s     uint
}

// New returns a newly created Map. The hash function takes a key and returns
// its hash. The eq function takes two keys and returns whether they are
// equal.
func New(hash func(*big.Int) int64, eq func(a, b *big.Int) bool, initialCapacity int) *Map {
	initialCapacity = mathutil.Max(1, initialCapacity)
	bits := uint(mathutil.Log2Uint64(uint64(initialCapacity)))
	initialCapacity = 1 << bits
	r := &Map{
		bits:  bits,
		eq:    eq,
		hash:  hash,
		items: make([][]item, initialCapacity),
		n:     uint(initialCapacity),
	}
	r.setL(0)
	return r
}

func (m *Map) addr(k *big.Int) uint {
	h := uint(m.hash(k))
	a := h & m.mask
	if a < uint(len(m.items)) {
		return a
	}

	return h & m.mask2
}

func (m *Map) setL(l uint) {
	m.mask = m.n<<l - 1
	m.mask2 = m.mask >> 1
	m.l = l
}

// Cursor returns a new map Cursor.
func (m *Map) Cursor() *Cursor { return &Cursor{m: m} }

// Delete removes the element with key k from the map.
func (m *Map) Delete(k *big.Int) {
	a := m.addr(k)
	b := m.items[a]
	for i, v := range b {
		if m.eq(v.k, k) {
			m.len--
			n := len(b) - 1
			if n == 0 {
				m.items[a] = nil
				return
			}

			b[i] = b[n]
			m.items[a] = b[:n]
			return
		}
	}
}

// Get returns the value associated with k and a boolean value indicating
// whether the key is in the map.
func (m *Map) Get(k *big.Int) (r *big.Int, ok bool) {
	a := m.addr(k)
	for _, v := range m.items[a] {
		if m.eq(v.k, k) {
			return v.v, true
		}
	}

	return r, false
}

// Insert inserts v into the map associating it with k.
func (m *Map) Insert(k *big.Int, v *big.Int) {
	a := m.addr(k)
	b := m.items[a]
	for i, bv := range b {
		if m.eq(bv.k, k) {
			b[i].v = v
			m.items[a] = b
			return
		}
	}

	b = append(b, item{k, v})
	m.items[a] = b
	m.len++
	if len(b) <= threshold {
		return
	}

	m.items = append(m.items, nil)
	b = m.items[m.s]
	m.items[m.s] = nil
	if m.s == 0 {
		m.setL(m.l + 1)
	}
	for _, v := range b {
		a := m.addr(v.k)
		m.items[a] = append(m.items[a], v)
	}
	m.s++
	if m.s-1 == m.mask2 {
		m.s = 0
	}
}

// Len returns the number of items in the map.
func (m *Map) Len() int { return m.len }
