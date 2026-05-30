package main

import (
	"net/http"
	"testing"
)

// Симулируем текущее поведение: новый transport на каждый вызов
func BenchmarkNewTransportEachTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &http.Transport{}
		_ = &http.Client{Transport: &http.Transport{}}
	}
}

// Правильный вариант: transport создаётся один раз
var sharedTransport = &http.Transport{}

func BenchmarkSharedTransport(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &http.Client{Transport: sharedTransport}
	}
}

// Симулируем полный createUploaders для 1 URL типа https+pow
func BenchmarkCreateUploadersFull(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Это то, что происходит сейчас при каждом sendMsgToPublicUploaders
		u := &struct {
			baseURL   string
			transport *http.Transport
		}{
			baseURL:   "https://pow.west.albion-online-data.com",
			transport: &http.Transport{},
		}
		_ = u
	}
}
