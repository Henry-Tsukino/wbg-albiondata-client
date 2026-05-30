package main

import (
	"regexp"
	"runtime"
	"testing"
)

var precompiled = regexp.MustCompile(`^[0-9]+$`)

func BenchmarkRegexpInside(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := regexp.MustCompile(`^[0-9]+$`)
		r.MatchString("3301")
	}
}

func BenchmarkRegexpPrecompiled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		precompiled.MatchString("3301")
	}
}

func BenchmarkGOMAXPROCSRepeat(b *testing.B) {
	procs := runtime.NumCPU() / 4
	if procs < 1 {
		procs = 1
	}
	runtime.GOMAXPROCS(procs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.GOMAXPROCS(procs)
	}
}

func BenchmarkNoGOMAXPROCS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = runtime.NumCPU()
	}
}
