# FunTime

Игры для компании в браузере. Первый этап: каталог с «Мировым господством», адаптивный интерфейс, поиск и смена темы.

**Комнаты и правила игры пока не реализованы.** Карточка и окно входа явно показывают этот статус. В каталоге нет выдуманных ограничений по игрокам, длительности и формул очков.

## Запуск без Docker

Требования: Node.js 24 LTS (минимум 22.12), npm, Go 1.26 или новее.

Из корня проекта:

```sh
npm ci
npm run dev
```

В Windows PowerShell, если выполнение `npm.ps1` запрещено, используйте `npm.cmd ci` и `npm.cmd run dev`.

Откройте **http://127.0.0.1:5173**. Команда собирает и запускает оба Go-сервиса и Vite. Остановка — `Ctrl+C`. При первом запуске нужны интернет и время на загрузку Go-зависимостей. Изменения frontend обновляются автоматически; после изменений Go перезапустите команду.

| Адрес | Назначение |
|---|---|
| `127.0.0.1:5173` | React / Vite |
| `127.0.0.1:8080/api/v1/games` | Публичный каталог через Gateway |
| `127.0.0.1:8080/healthz` | Проверка процесса Gateway |
| `127.0.0.1:8080/readyz` | Готовность Gateway с проверкой каталога |
| `127.0.0.1:9001` | Внутренний gRPC Catalog Service |
| `127.0.0.1:9002/healthz` | Проверка процесса Catalog Service |

## Запуск в Docker

Нужны Docker Engine / Docker Desktop и Compose v2.

```sh
docker compose -f deploy/compose.yaml -f deploy/compose.dev.yaml up --build -d
```

Откройте **http://127.0.0.1:8088**. Собираются четыре контейнера: Caddy, frontend (Nginx), Gateway, Catalog Service. Go-сервисы и frontend не публикуют свои порты наружу. Остановка:

```sh
docker compose -f deploy/compose.yaml -f deploy/compose.dev.yaml down
```

Для production соберите и опубликуйте три образа, скопируйте `deploy/production.env.example` в локальный `deploy/production.env`, укажите домен и реальные digest образов. Домен должен указывать на сервер; порты 80/443 должны быть доступны. Caddy хранит сертификаты в постоянном volume.

```sh
docker compose --env-file deploy/production.env -f deploy/compose.yaml -f deploy/compose.prod.yaml pull
docker compose --env-file deploy/production.env -f deploy/compose.yaml -f deploy/compose.prod.yaml up -d --wait
```

Автоматическая публикация образов и выкладка на реальный сервер ещё не настроены: для этого нужны репозиторий, реестр и целевой сервер. Workflow `.github/workflows/check.yml` проверяет код, генерируемые контракты, браузерные сценарии и сборку контейнеров.

## Проверки

```sh
npm run check
npm run build
npm run test:e2e
```

`check` проверяет TypeScript, ESLint, импорты FSD и Go-тесты. `build` собирает frontend в `frontend/dist` и Go-бинарники в `.cache/bin`.

Браузерные тесты в Windows используют установленный Microsoft Edge. В Linux/macOS перед первым запуском выполните `npx playwright install chromium` (в Linux могут потребоваться системные зависимости: `npx playwright install --with-deps chromium`). Тесты самостоятельно запускают приложение, если оно ещё не работает. Есть desktop- и mobile-сценарии.

## Контракты

- `contracts/proto/catalog/v1/catalog.proto` — gRPC между Go-сервисами.
- `contracts/http/openapi.yaml` — HTTP API для браузера.
- `contracts/gen/go` — сгенерированный Go-код, отдельный модуль.
- `frontend/src/shared/api/generated` — сгенерированные TypeScript-типы.

После изменения контрактов:

```sh
npm run generate
node scripts/go.mjs tidy
npm run check
```

Генерация использует Buf и закреплённые Go-плагины. Сгенерированные файлы хранятся в Git, поэтому обычному запуску генератор не требуется. Кеши инструментов и бинарники находятся в `.cache`, а версии npm-зависимостей закреплены `package-lock.json`.

Подробности границ и дальнейшего развития — в [docs/architecture.md](docs/architecture.md).
