# Тесты каналов

Запустим тесты каналов (код находится в `loadgen/cmd/tests/chans_test.go`):

```bash
cd loadgen

go test -run "^TestReadFromClosed$" -timeout 5s -v -count=1 ./...
go test -run "^TestWriteToClosed$" -timeout 5s -v -count=1 ./...
go test -run "^TestCloseClosed$" -timeout 5s -v -count=1 ./...
go test -run "^TestReadFromNil$" -timeout 5s -v -count=1 ./...
go test -run "^TestWriteToNil$" -timeout 5s -v -count=1 ./...
go test -run "^TestCloseNil$" -timeout 5s -v -count=1 ./...
```

# Инициализация БД

Запустим Postgres:

```bash
make pg
```

и накатим миграции:

```bash
docker run --rm \
    -v $(realpath ./app/internal/store/migrations):/migrations \
    migrate/migrate:v4.16.2 \
        -path=/migrations \
        -database postgres://gopher:gopher@172.17.0.2:5432/gopher_corp?sslmode=disable \
        up
```

# Генерация данных

Скомпилируем утилиту `datagen`:

```bash
make build-datagen
```

и сгенирируем данные для 100000 сотрудников:

```bash
./cmd/datagen/datagen -d postgres://gopher:gopher@localhost:5432/gopher_corp?sslmode=disable -n 1000000
```

Для остановки:

```bash
make stop-pg
```

Для удаления данных из БД:

```bash
make clean-data
```

# Генерация нагрузки

Для реализации генератор наргрузки необходимо реализовать функцию `loadgen/internal/loadgen/generator.go:runGenerator`.

Вот спецификация генератора:

1. Диспетчер генерирует набор имен (длину можно выбрать произвольно, но она должна быть больше 200) при помощи `common.NamesFetcher` и рассылает их рабочим
2. Рабочие отправляют http-запросы сервису, получают ответы. Если ответ содержит код отличный от `200` и `404`, то ошибка логируется
3. Длительность теста передается через конфиурацию. Как только время вышло, рабочие должны корректно завершиться
4. Результат работы функции - `loadTestResult`, который содержит реальную длительность тестов и количество успешно выполненных операций (при которых код овтета `200` или `404`)

Для запуска генератора скомпилируйте его:

```bash
make build-loadgen
```

И запустите с необходимыми параметрами:

```bash
./cmd/loadgen/loadgen -dur "5s" -workers 10
```

Номинальная реализация этой функции здесь: `loadgen/internal/loadgen/generator_nominal.go:runGeneratorNominal`

# Контекст и graceful shutdown

См. пример в файле `app/cmd/migrations/main.go`

Запустим приложение:

```bash
./cmd/migrations/migrations -dsn postgres://gopher:gopher@localhost:5432/gopher_corp?sslmode=disable
```

# Race-детектор

Запустим load-генератор с race-детектором. Для этого скомпилируем его с подключенным детектором:

```bash
make race-build-loadgen
```

И запустим:

```bash
./cmd/loadgen/loadgen -dur "5s" -workers 10
```
