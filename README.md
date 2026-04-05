# logger

Глобальный `slog.Logger` синглтон для Go-сервисов. JSON-вывод, несколько writers, не нужно прокидывать логгер через
аргументы.

## Установка

```bash
go get github.com/nextlag/logger
```

## Использование

```go
logger.SetLevel(os.Getenv("LOG_LEVEL"))
logger.SetServiceName("myapp")

log := logger.GetInstance()
log.Info("started", "port", 8080)
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