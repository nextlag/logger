package logger

import (
	"io"
	"log/slog"
	"testing"
)

func setupBench(b *testing.B, json bool) {
	b.Helper()

	global.mx.Lock()
	defer global.mx.Unlock()

	global.writerList = []io.Writer{io.Discard}
	global.attrList = nil
	global.serviceName = ""
	global.envName = defaultEnvName
	global.level = new(slog.LevelInfo)
	global.withSource = false
	global.useJSON = json
	global.handler = nil
	global.instance.Store(nil)
}

func BenchmarkGetInstance(b *testing.B) {
	setupBench(b, true)
	_ = GetInstance()

	b.ResetTimer()

	for b.Loop() {
		GetInstance()
	}
}

func BenchmarkLogJSON(b *testing.B) {
	setupBench(b, true)
	log := GetInstance()

	b.ResetTimer()

	for b.Loop() {
		log.Info("request handled", "method", "GET", "path", "/api/v1/users", "status", 200)
	}
}

func BenchmarkLogText(b *testing.B) {
	setupBench(b, false)
	log := GetInstance()

	b.ResetTimer()

	for b.Loop() {
		log.Info("request handled", "method", "GET", "path", "/api/v1/users", "status", 200)
	}
}

func BenchmarkLogJSONWithSource(b *testing.B) {
	setupBench(b, true)
	WithSource(true)

	log := GetInstance()

	b.ResetTimer()

	for b.Loop() {
		log.Info("request handled", "method", "GET", "path", "/api/v1/users", "status", 200)
	}
}

func BenchmarkLogTextWithSource(b *testing.B) {
	setupBench(b, false)
	WithSource(true)

	log := GetInstance()

	b.ResetTimer()

	for b.Loop() {
		log.Info("request handled", "method", "GET", "path", "/api/v1/users", "status", 200)
	}
}

func BenchmarkLogJSONParallel(b *testing.B) {
	setupBench(b, true)
	log := GetInstance()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info("request handled", "method", "GET", "path", "/api/v1/users", "status", 200)
		}
	})
}

func BenchmarkLogTextParallel(b *testing.B) {
	setupBench(b, false)
	log := GetInstance()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info("request handled", "method", "GET", "path", "/api/v1/users", "status", 200)
		}
	})
}
