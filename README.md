# Users API

Учебный проект курса «AI for Code»: небольшой HTTP API для учёта пользователей, реализованный в двух версиях — на **Python (Flask)** и на **Go** — с общим контрактом, документацией OpenAPI, Docker-образами и инструкцией по развёртыванию на сервере.

Проект вырос из намеренно сломанного кода: исходные `utils.py` и `api.py` не запускались и содержали SQL-инъекции, хранение паролей в открытом виде, гонки потоков и неверные HTTP-коды. Что было исправлено и добавлено — в разделе [«История проекта»](#история-проекта).

## Возможности

- Создание пользователей и получение их по `id`.
- Учёт активных пользователей: хранятся последние 5 отметок.
- Валидация входных данных и понятные JSON-ошибки.
- Интерактивная документация Swagger UI по адресу `/docs` — любой эндпоинт можно вызвать из браузера.
- Две взаимозаменяемые реализации с одинаковым API: Flask и Go.
- Запуск одной командой через Docker Compose; образы для x86 и ARM.
- Скрипт, проверяющий все эндпоинты на любом сервере.

## Технологии

| Компонент | Технологии |
|-----------|------------|
| Python-версия | Python 3.13, Flask 3, waitress, SQLite |
| Go-версия | Go 1.26, стандартная библиотека `net/http`, `modernc.org/sqlite` (без CGO) |
| Документация | OpenAPI 3.1, Swagger UI |
| Инфраструктура | Docker, Docker Compose, Docker Hub |

## Структура проекта

```
.
├── api.py                  Flask-версия API
├── utils.py                функции для работы с пользователями и паролями
├── requirements.txt        зависимости Python
├── openapi.yaml            спецификация API (общая для обеих версий)
├── docs.html               страница Swagger UI
├── Dockerfile              образ Flask-версии
├── docker-compose.yml      локальный запуск обеих версий
├── go-server/              Go-версия API (подробности — в go-server/README.md)
│   ├── cmd/server/         точка входа
│   ├── internal/           модели, хранилище, обработчики, роутер и т.д.
│   └── Dockerfile          образ Go-версии
├── deploy/                 файлы для запуска на сервере из образов Docker Hub
│   ├── docker-compose.yml
│   └── .env.example
├── scripts/
│   ├── check_api.py        проверка всех эндпоинтов
│   └── publish.ps1         публикация образов на Docker Hub
└── DEPLOY.md               пошаговая инструкция по развёртыванию
```

## Быстрый старт

Нужен только [Docker](https://docs.docker.com/get-docker/) (на Windows — запущенный Docker Desktop).

```bash
git clone https://github.com/SergeiAlAn/AI_For_Code.git
cd AI_For_Code
docker compose up -d --build
```

Откройте документацию в браузере:

| Версия | Адрес |
|--------|-------|
| Flask | [http://localhost:5000/docs](http://localhost:5000/docs) |
| Go | [http://localhost:5001/docs](http://localhost:5001/docs) |

Проверьте все эндпоинты (нужен Python, сторонние пакеты не требуются):

```bash
python scripts/check_api.py                                  # Flask, порт 5000
python scripts/check_api.py --base-url http://localhost:5001 # Go
```

На Windows вместо `python` используйте `py`.

Остановить: `docker compose down` (данные сохранятся) или `docker compose down -v` (вместе с данными).

## Запуск без Docker

### Python (Flask)

```powershell
py -m venv venv
.\venv\Scripts\Activate.ps1      # Linux/macOS: source venv/bin/activate
pip install -r requirements.txt
py api.py                         # Linux/macOS: python3 api.py
```

Сервер стартует на `http://127.0.0.1:5000`; база `test.db` создаётся в текущей папке. `py api.py` запускает отладочный сервер Flask — для постоянной работы используйте `waitress-serve --port=5000 api:app`.

### Go

```bash
cd go-server
go run ./cmd/server
```

Нужен Go 1.26+. Флаги, переменные окружения, сборка бинарника и устройство кода описаны в [go-server/README.md](go-server/README.md).

## HTTP API

| Метод и путь | Описание | Успех | Ошибки |
|--------------|----------|-------|--------|
| `POST /adduser` | Создать пользователя, тело `{"name": "Alice"}` | 201 `{"id": 1, "name": "Alice"}` | 400 |
| `GET /user/{id}` | Получить пользователя | 200 `{"id": 1, "name": "Alice"}` | 404 |
| `POST /user/{id}/active` | Отметить пользователя активным | 200 `{"active": [3, 1]}` | 404 |
| `GET /active` | Последние 5 отметок активности | 200 `{"active": [3, 1]}` | — |
| `GET /docs` | Интерактивная документация Swagger UI | 200 (HTML) | — |
| `GET /openapi.yaml` | Спецификация OpenAPI | 200 (YAML) | — |
| `GET /` | Перенаправление на `/docs` | 302 | — |

Правила и особенности:

- Ошибки возвращаются в виде `{"error": "описание"}`.
- `name` — непустая строка до 100 символов; пробелы по краям обрезаются, длина считается в символах (кириллица поддерживается).
- Список активных хранится в памяти процесса и очищается при перезапуске сервера.

Полное описание всех запросов, ответов и различий между версиями — в спецификации [`openapi.yaml`](openapi.yaml). Её удобнее всего смотреть на странице `/docs` запущенного сервера: там же каждый эндпоинт можно вызвать кнопкой **Execute**. Без запущенного сервера файл можно открыть в [Swagger Editor](https://editor.swagger.io) (File → Import file).

### Примеры

curl:

```bash
curl -X POST http://localhost:5000/adduser -H "Content-Type: application/json" -d '{"name": "Alice"}'
curl http://localhost:5000/user/1
curl -X POST http://localhost:5000/user/1/active
curl http://localhost:5000/active
```

PowerShell:

```powershell
$body = [Text.Encoding]::UTF8.GetBytes('{"name": "Алиса"}')
Invoke-RestMethod -Method Post http://localhost:5000/adduser -ContentType 'application/json; charset=utf-8' -Body $body
Invoke-RestMethod http://localhost:5000/user/1
Invoke-RestMethod -Method Post http://localhost:5000/user/1/active
Invoke-RestMethod http://localhost:5000/active
```

Тело передаётся байтами UTF-8, потому что Windows PowerShell 5 иначе искажает кириллицу.

## Тестирование

**Проверка работающего сервера** — `scripts/check_api.py`. Скрипт выполняет 16 проверок: документация и спецификация, создание пользователя, кириллица, ограничение длины, защита от SQL-инъекций, 9 видов некорректных запросов, ошибки 404 и 405, вытеснение старых отметок активности. Подходит для обеих версий и любого сервера.

| Параметр | По умолчанию | Описание |
|----------|--------------|----------|
| `--base-url` | `http://127.0.0.1:5000` | Адрес сервера |
| `--wait` | `15` | Сколько секунд ждать запуска сервера |
| `--timeout` | `5` | Таймаут одного запроса, секунд |

Код выхода: `0` — всё прошло, `1` — есть ошибки, `2` — сервер недоступен. Скрипт создаёт тестовых пользователей — на сервере с реальными данными учитывайте это.

**Модульные тесты Go-версии:**

```bash
cd go-server
go test ./...
```

Без установленного Go — в Docker:

```bash
docker run --rm -v "${PWD}:/w" -w /w/go-server golang:1.26-alpine go test ./...
```

## Развёртывание на сервере

Образы публикуются на Docker Hub скриптом `scripts/publish.ps1`, а на сервере запускаются из папки `deploy/` — исходный код и компиляторы на сервер не нужны:

```powershell
# на своём компьютере
docker login -u yourname
powershell -ExecutionPolicy Bypass -File scripts\publish.ps1 -User yourname -Version 1.0.0
scp -r deploy root@<IP-сервера>:~/users-api
```

```bash
# на сервере
cd ~/users-api && cp .env.example .env && nano .env   # указать DOCKERHUB_USER и TAG
docker compose pull && docker compose up -d
```

Пошаговая инструкция — регистрация на Docker Hub, подготовка сервера, обновление и откат версий, резервные копии, домен и HTTPS, решение типичных проблем — в [DEPLOY.md](DEPLOY.md).

## Модуль `utils.py`

Независимый от API набор функций: пользователи с тегами в SQLite и безопасное хранение паролей.

```python
import utils

uid = utils.add_user("Alice", ["admin"])   # -> 1
utils.get_user_by_name("Alice")            # -> {"id": 1, "name": "Alice", "tags": ["admin", "new"]}
utils.get_user_by_name("Bob")              # -> None

utils.store_password(uid, "s3cret")
utils.verify_password(uid, "s3cret")       # -> True
utils.verify_password(uid, "wrong")        # -> False

utils.set_active(uid)
utils.get_active_users()                   # -> [1]
```

| Функция | Описание |
|---------|----------|
| `add_user(name, tags=None) -> int` | Добавляет пользователя в `users.db`, дописывает к тегам `"new"` (переданный список не изменяется), возвращает id |
| `get_user_by_name(name) -> dict \| None` | Возвращает `{"id", "name", "tags"}` или `None` |
| `store_password(user_id, password) -> None` | Сохраняет солёный хеш пароля (PBKDF2-HMAC-SHA256, 600 000 итераций) в `passwords.txt` |
| `verify_password(user_id, password) -> bool` | Проверяет пароль; сравнение в постоянном времени |
| `set_active(user_id) -> None` | Потокобезопасно отмечает пользователя активным (хранятся последние 5) |
| `get_active_users() -> list` | Копия списка активных пользователей |

## Хранение данных

| Файл | Кем создаётся | Содержимое |
|------|---------------|------------|
| `test.db` | `api.py`, Go-версия | Таблица `users(id, name)` |
| `users.db` | `utils.py` | Таблица `users(id, name, tags)`, теги в JSON |
| `passwords.txt` | `utils.py` | Строки `user_id:salt_hex:hash_hex` |

Файлы создаются в текущей рабочей папке и исключены из git (`.gitignore`). В Docker база лежит в томе `/data` и переживает перезапуск и пересоздание контейнеров.

## Безопасность

- Все SQL-запросы параметризованы — SQL-инъекции исключены.
- Пароли хранятся только как солёные хеши PBKDF2.
- Входные данные валидируются; некорректные запросы получают 400, а не 500. Внутренние ошибки не раскрывают подробностей клиенту.
- Контейнеры работают не от root; Go-образ построен на distroless и не содержит даже командной оболочки.
- Секреты (токены Docker Hub, файл `.env`) в репозиторий не попадают — см. `.gitignore`.
- **В API нет авторизации.** Для публичного сервера закройте прямой доступ к портам и поставьте API за обратный прокси с авторизацией — см. раздел «Безопасность» в [DEPLOY.md](DEPLOY.md).

## История проекта

1. **Аудит и рефакторинг `utils.py`.** Файл не импортировался (синтаксические ошибки). Исправлены SQL-инъекции, «хеширование» пароля разворотом строки заменено на PBKDF2 с солью, устранены мутируемый аргумент по умолчанию, утечки соединений и файлов, гонка потоков, ошибки типов.
2. **Аудит и рефакторинг `api.py`.** Восстановлен нерабочий Flask-сервер: SQL-инъекции, глобальное соединение с базой, неверные HTTP-коды (200 вместо 404, 201 вместо 200), отсутствие валидации, недописанный эндпоинт активных пользователей.
3. **Документация** и файл зависимостей.
4. **Порт на Go** с тем же API, затем разделение на пакеты: модели, хранилище, валидация, обработчики, роутер, middleware.
5. **Спецификация OpenAPI 3.1**, проверенная валидатором и реальными ответами сервера.
6. **Docker**: образы для обеих версий, Docker Compose, скрипт проверки эндпоинтов.
7. **Swagger UI** на `/docs` в обеих версиях.
8. **Публикация на Docker Hub** (образы для x86 и ARM) и развёртывание на сервере.

## Документация

| Документ | Содержание |
|----------|------------|
| [openapi.yaml](openapi.yaml) | Полная спецификация API |
| [DEPLOY.md](DEPLOY.md) | Публикация на Docker Hub и развёртывание на сервере |
| [go-server/README.md](go-server/README.md) | Устройство Go-версии, её запуск и тесты |
