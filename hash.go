// Copyright 2017 The hash Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	"github.com/cznic/mathutil"
)

const threshold = 3

type item struct {
	k interface{} /*K*/
	v interface{} /*V*/
}

// Cursor provides enumerating of Map items.
type Cursor struct {
	K        interface{} /*K*/
	V        interface{} /*V*/
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
	bits  int
	eq    func(a, b interface{} /*K*/) bool
	hash  func(interface{} /*K*/) int64
	items [][]item
	len   int
	level int
	mask  int64
	mask2 int64
	split int64
}

// New returns a newly created Map. The hash function takes a key and returns
// its hash. The eq function takes two keys and returns whether they are
// equal.
func New(hash func(interface{} /*K*/) int64, eq func(a, b interface{} /*K*/) bool, initialCapacity int) *Map {
	initialCapacity = mathutil.Max(16, initialCapacity)
	bits := mathutil.Log2Uint64(uint64(initialCapacity))
	initialCapacity = 1 << uint(bits)
	return &Map{
		bits:  bits,
		eq:    eq,
		hash:  hash,
		items: make([][]item, initialCapacity),
		mask2: int64(1)<<uint(bits-1) - 1,
		mask:  int64(1)<<uint(bits) - 1,
	}
}

func (m *Map) lv(j int) {
	m.level = j
	m.mask = int64(1)<<uint(j+m.bits) - 1
	m.mask2 = m.mask >> 1
}

func (m *Map) a(h int64) int64 {
	a := h & m.mask
	if a >= int64(len(m.items)) {
		return h & m.mask2
	}

	return a
}

// Cursor returns a new map Cursor.
func (m *Map) Cursor() *Cursor { return &Cursor{m: m} }

// Delete removes the element with key k from the map.
func (m *Map) Delete(k interface{} /*K*/) {
	a := m.a(m.hash(k))
	b := m.items[a]
	for i, item := range b {
		if m.eq(k, item.k) {
			c := len(b) - 1
			if i < c {
				b[i] = b[c]
			}
			b = b[:c]
			m.items[a] = b
			m.len--
			if c < threshold {
				return
			}

			if m.split == 0 {
				m.split = m.mask2 + 1
			}
			m.split--
			c = len(m.items) - 1
			m.items[m.split] = append(m.items[m.split], m.items[c]...)
			m.items = m.items[:c]
			if m.split == 0 {
				m.lv(m.level - 1)
			}
		}
	}
}

// Get returns the value associated with k and a boolean value indicating
// whether the key is in the map.
func (m *Map) Get(k interface{} /*K*/) (interface{} /*V*/, bool) {
	a := m.a(m.hash(k))
	for _, item := range m.items[a] {
		if m.eq(k, item.k) {
			return item.v, true
		}
	}
	return 0, false
}

// Insert inserts v into the map associating it with k.
func (m *Map) Insert(k interface{} /*K*/, v interface{} /*V*/) {
	a := m.a(m.hash(k))
	b := m.items[a]
	for i, item := range b {
		if m.eq(k, item.k) {
			b[i].v = v
			return
		}
	}

	m.len++
	b = append(b, item{k, v})
	m.items[a] = b
	if len(b) <= threshold {
		return
	}

	m.items = append(m.items, []item(nil))
	if m.split == 0 {
		m.lv(m.level + 1)
	}

	b = m.items[m.split]
	m.items[m.split] = nil
	for _, item := range b {
		a := m.a(m.hash(item.k))
		m.items[a] = append(m.items[a], item)
	}
	m.split++
	if m.split == m.mask2+1 {
		m.split = 0
	}
}

// Len returns the number of items in the map.
func (m *Map) Len() int { return m.len }
