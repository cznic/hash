// Copyright 2017 The hash Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	"github.com/cznic/mathutil"
)

const threshold = 2

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
//
// The iteration order is not specified and is not guaranteed to be the same
// from one iteration to the next. If a map entry that has not yet been reached
// is removed during iteration, the corresponding iteration value will not be
// produced. If a map entry is created during iteration, that entry may be
// produced during the iteration or may be skipped. The choice may vary for
// each entry created and from one iteration to the next.
func (c *Cursor) Next() bool {
	if c.m == nil {
		return false
	}

	if c.hasMoved {
		c.j++
	}
	c.hasMoved = true

	for n := len(c.m.items); c.i < n; c.i, c.j = c.i+1, 0 {
		b := c.m.items[c.i]
		for ; c.j < len(b); c.j++ {
			if b[c.j].k != nil {
				c.K = b[c.j].k
				c.V = b[c.j].v
				return true
			}
		}
	}

	c.m = nil
	return false
}

// Map is a hash table.
type Map struct {
	eq    func(a, b interface{} /*K*/) bool
	hash  func(interface{} /*K*/) int64
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
func New(hash func(interface{} /*K*/) int64, eq func(a, b interface{} /*K*/) bool, initialCapacity int) *Map {
	initialCapacity = mathutil.Max(1, initialCapacity)
	initialCapacity = 1 << uint(mathutil.Log2Uint64(uint64(initialCapacity)))
	r := &Map{
		eq:    eq,
		hash:  hash,
		items: make([][]item, initialCapacity),
		n:     uint(initialCapacity),
	}
	r.setL(0)
	return r
}

func (m *Map) addr(k interface{} /*K*/) uint {
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
func (m *Map) Delete(k interface{} /*K*/) {
	a := m.addr(k)
	b := m.items[a]
	for i, v := range b {
		if v.k != nil && m.eq(v.k, k) {
			m.len--
			n := len(b) - 1
			if n == 0 {
				m.items[a] = nil
				return
			}

			if i == n {
				m.items[a] = b[:n]
				return
			}

			b[i] = item{}
			return
		}
	}
}

// Get returns the value associated with k and a boolean value indicating
// whether the key is in the map.
func (m *Map) Get(k interface{} /*K*/) (r interface{} /*V*/, ok bool) {
	a := m.addr(k)
	for _, v := range m.items[a] {
		if v.k != nil && m.eq(v.k, k) {
			return v.v, true
		}
	}

	return nil, false
}

// Insert inserts v into the map associating it with k.
func (m *Map) Insert(k interface{} /*K*/, v interface{} /*V*/) {
	a := m.addr(k)
	b := m.items[a]
	j := -1
	for i, bv := range b {
		switch {
		case bv.k == nil:
			j = i
		default:
			if m.eq(bv.k, k) {
				b[i].v = v
				m.items[a] = b
				return
			}
		}
	}

	m.len++
	if j >= 0 {
		b[j] = item{k, v}
		return
	}

	b = append(b, item{k, v})
	m.items[a] = b
	if len(b) <= threshold {
		return
	}

	m.items = append(m.items, nil)
	b = m.items[m.s]
	m.items[m.s] = nil
	if m.s == 0 {
		m.setL(m.l + 1)
	}
outer:
	for _, v := range b {
		if v.k == nil {
			continue
		}

		a := m.addr(v.k)
		c := m.items[a]
		for i, w := range c {
			if w.k == nil {
				c[i] = v
				continue outer
			}
		}

		m.items[a] = append(c, v)
	}
	m.s++
	if m.s-1 == m.mask2 {
		m.s = 0
	}
}

// Len returns the number of items in the map.
func (m *Map) Len() int { return m.len }

// Vacuum rebuilds m, repacking it into a possibly smaller amount of memory.
func (m *Map) Vacuum() {
	m2 := New(m.hash, m.eq, m.Len())
	c := m.Cursor()
	for c.Next() {
		m.Delete(c.K)
		m2.Insert(c.K, c.V)
	}
	*m = *m2
}
