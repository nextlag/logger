# logger

Глобальный `slog.Logger` синглтон для Go-сервисов. JSON/text вывод, несколько writers, цветные уровни, не нужно прокидывать логгер через аргументы.

## Установка

```bash
go get github.com/nextlag/logger
```

## Использование

```go
logger.SetLevel("DEBUG")
logger.SetServiceName("myapp")

log := logger.GetInstance()
log.Info("started", "port", 8080)
```

### Text-вывод с цветами

```go
logger.WithJSON(false)
logger.SetLevel("DEBUG")

log := logger.GetInstance()
log.Debug("connecting to database", "host", "localhost")
log.Info("server started", "port", 8080)
log.Warn("high memory usage", "percent", 92)
log.Error("connection lost", "err", "timeout")
```

```
2025-01-15 12:30:00 DEBUG "connecting to database" myapp.host=localhost
2025-01-15 12:30:00 INFO  "server started" myapp.port=8080
2025-01-15 12:30:00 WARN  "high memory usage" myapp.percent=92
2025-01-15 12:30:00 ERROR "connection lost" myapp.err=timeout
```

### JSON-вывод (по умолчанию)

```json
{"time":"2025-01-15T12:30:00","level":"INFO","msg":"server started","myapp":{"port":8080}}
```

### Несколько writers

```go
f, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
defer f.Close()

logger.AddWriter(f) // stdout + файл
```

### Source-информация

```go
logger.WithSource(true)
```

```json
{"time":"2025-01-15T12:30:00","level":"INFO","msg":"started","source":"cmd/main.go:15"}
```

### Custom handler

```go
logger.WithHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
```

## API

| Функция                        | Описание                                                       |
|--------------------------------|----------------------------------------------------------------|
| `GetInstance() *slog.Logger`   | Возвращает закешированный экземпляр, создаёт при первом вызове |
| `SetLevel(level string)`       | Устанавливает уровень (`DEBUG`, `INFO`, `WARN`, `ERROR`)       |
| `SetServiceName(name string)`  | Группирует все атрибуты под именем сервиса                     |
| `AddWriter(w io.Writer)`       | Добавляет writer (по умолчанию `os.Stdout`)                    |
| `WithAttr(attrs ...slog.Attr)` | Добавляет глобальные атрибуты ко всем записям                  |
| `WithSource(bool)`             | Включает вывод файла и строки в логах                          |
| `WithJSON(bool)`               | Переключает JSON (`true`) / text (`false`) формат              |
| `WithHandler(slog.Handler)`    | Устанавливает произвольный handler                             |

Все функции потокобезопасны. Изменение конфигурации сбрасывает кеш - следующий `GetInstance()` создаст новый экземпляр.

## Производительность

```
goos: darwin, goarch: amd64, cpu: Intel Core i7-9750H @ 2.60GHz

BenchmarkGetInstance            576 151 482    2.0 ns/op      0 B/op   0 allocs/op
BenchmarkLogJSON                    715 207   1700 ns/op     56 B/op   3 allocs/op
BenchmarkLogText                  1 519 186    791 ns/op     40 B/op   3 allocs/op
BenchmarkLogJSONWithSource          558 013   2150 ns/op    369 B/op   7 allocs/op
BenchmarkLogTextWithSource        1 361 030    889 ns/op     40 B/op   3 allocs/op
BenchmarkLogJSONParallel          3 107 156    381 ns/op     56 B/op   3 allocs/op
BenchmarkLogTextParallel          6 342 124    197 ns/op     40 B/op   3 allocs/op
```

```bash
go test -bench=. -benchmem ./...
```
