// Copyright 2017 The hash Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/cznic/mathutil"
)

func caller(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(2)
	fmt.Fprintf(os.Stderr, "# caller: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	_, fn, fl, _ = runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# \tcallee: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func dbg(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# dbg %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func TODO(...interface{}) string { //TODOOK
	_, fn, fl, _ := runtime.Caller(1)
	return fmt.Sprintf("# TODO: %s:%d:\n", path.Base(fn), fl) //TODOOK
}

func use(...interface{}) {}

func init() {
	use(caller, dbg, TODO) //TODOOK
}

// ============================================================================

var (
	exp = flag.Int("e", -1, "")
)

func fnv(k *big.Int) int64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)

	hash := uint64(offset64)
	hash ^= uint64(k.Sign())
	hash *= prime64
	for _, v := range k.Bits() {
		hash ^= uint64(v)
		hash *= prime64
	}
	return int64(hash)
}

func cmp(a, b *big.Int) bool { return a.Cmp(b) == 0 }

func rnda(n int) []int64 {
	r, err := mathutil.NewFC32(math.MinInt32, math.MaxInt32, true)
	if err != nil {
		panic(err)
	}

	a := make([]int64, n)
	for i := range a {
		a[i] = int64(uint32(r.Next()))<<32 | int64(uint32(r.Next()))
	}
	return a
}

func (m *Map) insert(k, v int64) { m.Insert(big.NewInt(k), big.NewInt(v)) }

func (m *Map) get(k int64) (int64, bool) {
	if v, ok := m.Get(big.NewInt(k)); ok {
		return v.Int64(), true
	}

	return 0, false
}

func test0(t *testing.T, initialCap, sz int) {
	mp := New(fnv, cmp, initialCap)
	n := 2 * initialCap
	for i := 0; i < n; i++ {
		mp.insert(int64(i), int64(10*i))
		if g, e := mp.Len(), i+1; g != e {
			t.Fatal(g, e)
		}
	}

	for i := 0; i < n; i++ {
		mp.insert(int64(i), int64(10*i))
		if g, e := mp.Len(), n; g != e {
			t.Fatal(g, e)
		}
	}

	a := rnda(sz)
	mp = New(fnv, cmp, initialCap)
	for v, key := range a {
		mp.insert(int64(key), int64(v))
		if g, e := mp.Len(), v+1; g != e {
			t.Fatal(g, e)
		}
	}

	for i, key := range a {
		v, ok := mp.get(int64(key))
		if g, e := ok, true; g != e {
			t.Logf(
				"initialCap %d, threshold %d, i %d, key %d",
				initialCap, threshold, i, key,
			)
			t.Fatal(g, e)
		}

		if g, e := v, int64(i); g != e {
			t.Logf(
				"initialCap %d, threshold %d, i %d, key %d",
				initialCap, threshold, i, key,
			)
			t.Fatal(g, e)
		}
	}
}

func Test0(t *testing.T) {
	for initialCap := 1; initialCap <= 16; initialCap <<= 1 {
		test0(t, initialCap, 51000)
	}
}

func (m *Map) delete(k int64) { m.Delete(big.NewInt(k)) }

func testDelete(t *testing.T, initialCap, sz int) {
	a := rnda(sz)
	mp := New(fnv, cmp, initialCap)
	for v, key := range a {
		mp.insert(int64(key), int64(v))
	}

	for i := len(a) - 1; i >= 0; i-- {
		key := a[i]
		v, ok := mp.get(int64(key))
		if g, e := ok, true; g != e {
			t.Logf(
				"initialCap %d, threshold %d, i %d, key %d",
				initialCap, threshold, i, key,
			)
			t.Fatal(g, e)
		}

		if g, e := v, int64(i); g != e {
			t.Logf(
				"initialCap %d, threshold %d, i %d, key %d",
				initialCap, threshold, i, key,
			)
			t.Fatal(g, e)
		}

		mp.delete(int64(key))
		if g, e := mp.Len(), i; g != e {
			t.Fatal(g, e)
		}

		_, ok = mp.get(int64(key))
		if g, e := ok, false; g != e {
			t.Logf(
				"initialCap %d, threshold %d, i %d, key %d",
				initialCap, threshold, i, key,
			)
			t.Fatal(g, e)
		}

		for j := 0; j < i; j++ {
			v, ok := mp.get(int64(a[j]))
			if g, e := ok, true; g != e {
				t.Logf(
					"i %v, j %v, initialCap %d, threshold %d, key %d",
					i, j, initialCap, threshold, a[j],
				)
				t.Fatal(g, e)
			}

			if g, e := v, int64(j); g != e {
				t.Logf(
					"i %v, j %v, initialCap %d, threshold %d, key %d",
					i, j, initialCap, threshold, a[j],
				)
				t.Fatal(g, e)
			}
		}
	}
}

func TestDelete(t *testing.T) {
	for initialCap := 1; initialCap <= 16; initialCap <<= 1 {
		testDelete(t, initialCap, 1100)
	}
}

func TestMap(t *testing.T) {
	a := rnda(560000)
	m := make(map[int64]int64, len(a))
	mp := New(fnv, cmp, 16)
	for v, key := range a {
		m[key] = int64(v)
		mp.insert(key, int64(v))
	}

	for v, key := range a {
		switch v % 3 {
		case 0:
			mp.delete(key)
			delete(m, key)
		case 1:
			mp.insert(key, -int64(v))
			m[key] = -int64(v)
		}
	}

	for v, key := range a {
		if v%3 == 0 {
			_, ok := mp.get(key)
			if g, e := ok, false; g != e {
				t.Fatal(g, e)
			}
		}
	}

	for key, v := range m {
		g, ok := mp.get(key)
		if !ok {
			t.Fatal(false)
		}

		if e := v; g != e {
			t.Fatal(g, e)
		}
	}

	for c := mp.Cursor(); c.Next(); {
		if len(m) == 0 {
			t.Fatal("Cursor fail")
		}

		k := c.K.Int64()
		e, ok := m[k]
		if !ok {
			t.Fatal("Cursor fail")
		}

		if g, e := c.V.Int64(), e; g != e {
			t.Fatal("Cursor fail")
		}

		delete(m, k)
	}
	if len(m) != 0 {
		t.Fatal("Cursor fail")
	}
}

func benchmarkGet(b *testing.B, sz int) {
	a := rnda(sz)
	m := New(fnv, cmp, 0)
	for v, k := range a {
		m.insert(k, int64(v))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for v, k := range a {
			g, ok := m.get(k)
			if !ok || g != int64(v) {
				b.Fatal(ok, g, v)
			}
		}
	}
	b.StopTimer()
}

func BenchmarkGet(b *testing.B) {
	var n int
	for _, e := range []int{3, 4, 5, 6} {
		if *exp > 0 && *exp != e {
			continue
		}

		n = 1
		for i := 0; i < e; i++ {
			n *= 10
		}
		b.Run(fmt.Sprintf("1e%d", e), func(b *testing.B) { benchmarkGet(b, n) })
	}
}

func benchmarkInsert(b *testing.B, sz int) {
	a := rnda(sz)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := New(fnv, cmp, 0)
		for v, k := range a {
			m.insert(k, int64(v))
		}
	}
	b.StopTimer()
}

func BenchmarkInsert(b *testing.B) {
	var n int
	for _, e := range []int{3, 4, 5, 6} {
		if *exp > 0 && *exp != e {
			continue
		}

		n = 1
		for i := 0; i < e; i++ {
			n *= 10
		}
		b.Run(fmt.Sprintf("1e%d", e), func(b *testing.B) { benchmarkInsert(b, n) })
	}
}

func benchmarkDelete(b *testing.B, sz int) {
	a := rnda(sz)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m := New(fnv, cmp, 0)
		b.StartTimer()
		for _, k := range a {
			m.delete(k)
		}
	}
	b.StopTimer()
}

func BenchmarkDelete(b *testing.B) {
	var n int
	for _, e := range []int{3, 4, 5, 6} {
		if *exp > 0 && *exp != e {
			continue
		}

		n = 1
		for i := 0; i < e; i++ {
			n *= 10
		}
		b.Run(fmt.Sprintf("1e%d", e), func(b *testing.B) { benchmarkDelete(b, n) })
	}
}
