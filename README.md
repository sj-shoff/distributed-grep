# Distributed-grep - Распределённая CLI-утилита grep с кворумом и concurrency

## Описание
Это распределённая версия утилиты `grep`, реализованная на Go. Поддерживает работу в многосерверном режиме: клиент делит данные на чанки, рассылает на серверы через RPC, собирает результаты с проверкой кворума (N/2 + 1 успешных серверов). Внутри серверов - concurrency с горутинами для обработки чанков.

### Фичи
- **Флаги grep (subset)**: -F (fixed string), -i (ignore case), -v (invert), -n (line nums), -A/-B/-C (context), -c (count).
- **Concurrency**: Горутины для параллельного поиска в чанках (кол-во настраивается).
- **Кворум**: Если меньше кворума успешных серверов - ошибка.
- **Конфиг**: Через .env/env vars (таймаут, горутины, default addr).
- **Логи**: zerolog.
- **Ошибки**: Кастомные в domain/errors.
- **Архитектура**: Domain (models/errors), Usecases (service), Adapters (rpc/cli), Config (env/flags), App (composition root).

### Требования
- Go 1.24+

### Запуск
```bash
go run ./cmd/dgrep/main.go
```

### Сборка
```bash
go build -o dgrep ./cmd/dgrep
```