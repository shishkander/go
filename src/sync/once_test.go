// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync_test

import (
	. "sync"
	"sync/atomic"
	"testing"
)

type one int

func (o *one) Increment() {
	*o++
}

func run(t *testing.T, once *Once, o *one, c chan bool) {
	once.Do(func() { o.Increment() })
	if v := *o; v != 1 {
		t.Errorf("once failed inside run: %d is not 1", v)
	}
	c <- true
}

func TestOnce(t *testing.T) {
	o := new(one)
	once := new(Once)
	c := make(chan bool)
	const N = 10
	for i := 0; i < N; i++ {
		go run(t, once, o, c)
	}
	for i := 0; i < N; i++ {
		<-c
	}
	if *o != 1 {
		t.Errorf("once failed outside run: %d is not 1", *o)
	}
}

func TestOncePanic(t *testing.T) {
	var once Once
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("Once.Do did not panic")
			}
		}()
		once.Do(func() {
			panic("failed")
		})
	}()

	once.Do(func() {
		t.Fatalf("Once.Do called twice")
	})
}

func BenchmarkOnce(b *testing.B) {
	var once Once
	f := func() {}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			once.Do(f)
		}
	})
}

const slowN = int(1)

type slow struct {
	a [slowN]byte
}

func (s *slow) write() {
	for i := 0; i < slowN; i++ {
		s.a[i] = 1
	}
}
func (s *slow) test(t *testing.T) bool {
	for i := slowN - 1; i >= 0; i-- {
		b := s.a[i]
		if b != 1 {
			if t != nil {
				t.Errorf("fuck up at %d %d", i, b)
			}
			return false
		}
	}
	return true
}

func BenchmarkOnceSlow(b *testing.B) {
	var once Once
	s := slow{}
	oncef := s.write
	ok := int32(1)
	f := func() {
		once.Do(oncef)
		if !s.test(nil) {
			atomic.StoreInt32(&ok, 0)
		}
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f()
		}
	})
	if atomic.LoadInt32(&ok) == 0 {
		b.Fatal("failed test.")
	}
}
