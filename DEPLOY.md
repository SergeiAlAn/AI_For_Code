# Публикация на Docker Hub и запуск на сервере

Пошаговая инструкция: от регистрации на Docker Hub до работающего API на сервере, обновлений и резервных копий.

## Как это устроено

```
 Ваш компьютер                    Docker Hub                       Сервер
┌──────────────────┐   push   ┌──────────────────┐   pull   ┌────────────────────┐
│ исходный код     │ ───────► │ yourname/        │ ───────► │ deploy/            │
│ + Dockerfile     │          │  users-api-go    │          │  docker-compose.yml│
│                  │          │  users-api-python│          │  .env              │
│ scripts/         │          │                  │          │                    │
│  publish.ps1     │          │ (хранилище       │          │ контейнеры         │
└──────────────────┘          │  образов)        │          └────────────────────┘
                              └──────────────────┘
```

1. На своём компьютере вы **собираете образы** — архивы, в которых уже лежат программа и всё, что нужно для её запуска.
2. Образы **отправляются на Docker Hub** — это как GitHub, только для образов.
3. Сервер **скачивает готовые образы** и запускает их. Исходный код, Python и Go на сервере не нужны — только Docker и два небольших файла из папки `deploy/`.

Файлы проекта, которые для этого используются:

| Файл | Где используется | Назначение |
|------|------------------|------------|
| `scripts/publish.ps1` | ваш компьютер | Собирает и публикует оба образа |
| `deploy/docker-compose.yml` | сервер | Описывает, какие образы запускать и с какими настройками |
| `deploy/.env.example` | сервер | Шаблон настроек: логин Docker Hub, версия, порты |

В инструкции `yourname` — ваш логин на Docker Hub, `203.0.113.10` — IP-адрес сервера. Подставляйте свои значения.

---

## Часть 1. Публикация на Docker Hub

### Шаг 1. Регистрация на Docker Hub

1. Откройте [hub.docker.com](https://hub.docker.com) и нажмите **Sign up**.
2. Придумайте логин (**Username**). Он станет частью имени образов: `yourname/users-api-go`. Поменять его потом сложно, поэтому выбирайте осознанно.
3. Подтвердите почту.

### Шаг 2. Токен доступа

Входить в Docker из командной строки лучше не паролем, а **токеном доступа**: его можно в любой момент отозвать, не меняя пароль.

1. На Docker Hub нажмите на аватар справа вверху → **Account settings**.
2. Слева выберите **Personal access tokens** → **Generate new token**.
3. Заполните:
   - **Description** — например, `my-laptop`;
   - **Expiration date** — срок действия;
   - **Access permissions** — **Read & Write** (нужна запись, чтобы публиковать образы).
4. Нажмите **Generate** и **сразу скопируйте токен** (начинается с `dckr_pat_`). Docker Hub покажет его только один раз.

### Шаг 3. Вход в Docker Hub на своём компьютере

Убедитесь, что Docker Desktop запущен (значок кита в трее), затем в PowerShell:

```powershell
docker login -u yourname
```

На запрос `Password:` вставьте токен (правой кнопкой мыши или Ctrl+V) и нажмите Enter. Символы при вводе не отображаются — это нормально. Должно появиться `Login Succeeded`.

### Шаг 4. Публикация образов

Перейдите в папку проекта и запустите скрипт:

```powershell
cd путь\к\проекту
powershell -ExecutionPolicy Bypass -File scripts\publish.ps1 -User yourname -Version 1.0.0
```

`-ExecutionPolicy Bypass` нужен, потому что Windows по умолчанию запрещает запуск `.ps1`-скриптов. Этот флаг разрешает запуск только этого скрипта и только сейчас.

Что делает скрипт:

1. Собирает образы Flask- и Go-версии сразу для двух типов процессоров:
   - `linux/amd64` — обычные серверы на Intel/AMD (подавляющее большинство VPS);
   - `linux/arm64` — ARM-серверы (Hetzner CAX, Oracle Ampere, AWS Graviton и т.п.).

   Docker на сервере сам скачает подходящий вариант.
2. Ставит каждому образу два тега:
   - `1.0.0` — конкретная версия, она никогда не меняется;
   - `latest` — «последняя опубликованная», переезжает на каждую новую версию.
3. Отправляет образы на Docker Hub.

Первая публикация занимает несколько минут, следующие — быстрее: неизменившиеся слои не пересобираются и не отправляются заново.

В конце скрипт выведет:

```
Готово. Опубликованы образы:
  yourname/users-api-python:1.0.0
  yourname/users-api-go:1.0.0
```

Дополнительные параметры:

```powershell
# Только Go-версия
powershell -ExecutionPolicy Bypass -File scripts\publish.ps1 -User yourname -Version 1.0.0 -Only go

# Только для обычных x86-серверов (быстрее)
powershell -ExecutionPolicy Bypass -File scripts\publish.ps1 -User yourname -Version 1.0.0 -Platforms linux/amd64

# Не трогать тег latest (например, для тестовой версии)
powershell -ExecutionPolicy Bypass -File scripts\publish.ps1 -User yourname -Version 1.1.0 -NoLatest
```

<details>
<summary>То же самое без скрипта (Linux, macOS или вручную)</summary>

```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -t yourname/users-api-python:1.0.0 -t yourname/users-api-python:latest \
  --push .

docker buildx build --platform linux/amd64,linux/arm64 \
  -t yourname/users-api-go:1.0.0 -t yourname/users-api-go:latest \
  --push ./go-server
```

Если при сборке для двух платформ появляется ошибка `Multi-platform build is not supported for the docker driver`, создайте сборщик один раз: `docker buildx create --use --name multi` и повторите команды.

</details>

### Шаг 5. Проверка на Docker Hub

Откройте [hub.docker.com/repositories](https://hub.docker.com/repositories) — там появятся `users-api-python` и `users-api-go`. Во вкладке **Tags** каждого видны `1.0.0` и `latest`, а в столбце **OS/ARCH** — `linux/amd64` и `linux/arm64`.

### Публичный или приватный репозиторий

По умолчанию новые репозитории на Docker Hub **публичные**: скачать образ может любой. Учтите:

- В образе Flask-версии лежит исходный код `api.py` в открытом виде. Go-образ содержит только скомпилированную программу, но и её можно исследовать.
- Секретов (паролей, ключей) в образах этого проекта нет, так что публичный вариант допустим для учебного проекта.

Чтобы сделать репозиторий приватным: откройте его на Docker Hub → **Settings** → **Visibility settings** → **Make private**. На бесплатном тарифе доступен один приватный репозиторий. Для приватных образов на сервере понадобится `docker login` (см. шаг 10).

---

## Часть 2. Подготовка сервера

### Шаг 6. Какой сервер нужен

Подойдёт любой VPS (виртуальный сервер) у любого провайдера:

| Параметр | Минимум | Комментарий |
|----------|---------|-------------|
| Процессор | 1 vCPU | x86 или ARM — образы есть для обоих |
| Память | 512 МБ | 1 ГБ комфортнее; сами сервисы занимают мало: Go-версия ~5 МБ, Flask ~30 МБ |
| Диск | 10 ГБ | Сам Docker с образами занимает ~1 ГБ |
| ОС | Ubuntu 24.04 LTS | Команды ниже — для Ubuntu/Debian |

После создания сервера провайдер выдаст **IP-адрес** и **пароль root** (или попросит загрузить SSH-ключ).

### Шаг 7. Подключение к серверу

В Windows 10/11 SSH-клиент уже встроен. В PowerShell:

```powershell
ssh root@203.0.113.10
```

При первом подключении появится вопрос `Are you sure you want to continue connecting (yes/no)?` — введите `yes`. Затем введите пароль (символы не отображаются).

Приглашение командной строки сменится на что-то вроде `root@my-server:~#` — теперь все команды выполняются **на сервере**.

<details>
<summary>Рекомендуется: вход по SSH-ключу вместо пароля</summary>

На **своём компьютере** (не на сервере):

```powershell
ssh-keygen -t ed25519            # на все вопросы — Enter
type $env:USERPROFILE\.ssh\id_ed25519.pub | ssh root@203.0.113.10 "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
```

После этого `ssh root@203.0.113.10` будет входить без пароля. Ключ надёжнее пароля: его невозможно подобрать перебором.

</details>

### Шаг 8. Обновление системы и файрвол

На сервере:

```bash
apt update && apt upgrade -y
```

Если появится окно с вопросом о перезапуске служб или конфигурации — нажмите Enter (вариант по умолчанию).

Включите файрвол, разрешив SSH (иначе потеряете доступ к серверу):

```bash
ufw allow OpenSSH
ufw enable          # на вопрос ответьте y
```

> **Важно: Docker обходит ufw.** Порты, которые публикует Docker (в нашем случае 5000 и 5001), доступны из интернета, даже если в ufw они не разрешены. Поэтому доступ к API регулируется не ufw, а параметром `BIND_ADDR` в файле `.env` (шаг 11). Кроме того, у многих провайдеров есть свой файрвол в панели управления — если API не открывается снаружи, проверьте и его.

### Шаг 9. Установка Docker

Официальный скрипт устанавливает Docker и плагин Compose:

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
```

Проверьте:

```bash
docker --version
docker compose version
docker run --rm hello-world
```

Последняя команда должна вывести `Hello from Docker!`. Docker запускается автоматически при загрузке сервера.

<details>
<summary>Если работаете не под root, а под обычным пользователем</summary>

Добавьте пользователя в группу `docker`, чтобы не писать `sudo` перед каждой командой, и переподключитесь по SSH:

```bash
sudo usermod -aG docker $USER
exit
```

</details>

---

## Часть 3. Запуск проекта на сервере

### Шаг 10. Копирование файлов на сервер

На сервер нужна только папка `deploy`. На **своём компьютере** в новом окне PowerShell:

```powershell
cd путь\к\проекту
scp -r deploy root@203.0.113.10:~/users-api
```

На сервере появится папка `~/users-api` с файлами `docker-compose.yml` и `.env.example`.

<details>
<summary>Без scp: создать файлы вручную на сервере</summary>

```bash
mkdir -p ~/users-api && cd ~/users-api
nano docker-compose.yml
```

Вставьте содержимое `deploy/docker-compose.yml` (правая кнопка мыши в окне PowerShell), сохраните: Ctrl+O, Enter, выйдите: Ctrl+X. Файл `.env` создайте так же (шаг 11).

</details>

Если репозиторий на Docker Hub **приватный**, войдите на сервере (токену достаточно прав **Read-only**; лучше создать для сервера отдельный токен):

```bash
docker login -u yourname
```

### Шаг 11. Настройка

На сервере:

```bash
cd ~/users-api
cp .env.example .env
nano .env
```

Поменяйте значения:

```bash
DOCKERHUB_USER=yourname   # ваш логин на Docker Hub
TAG=1.0.0                 # версия, которую опубликовали
BIND_ADDR=0.0.0.0         # 0.0.0.0 — доступ из интернета; 127.0.0.1 — только с сервера
PYTHON_PORT=5000
GO_PORT=5001
```

Сохраните (Ctrl+O, Enter) и выйдите (Ctrl+X).

Проверьте, что Compose правильно прочитал настройки:

```bash
docker compose config --images
```

Должно вывести `yourname/users-api-python:1.0.0` и `yourname/users-api-go:1.0.0`. Если видите ошибку `required variable DOCKERHUB_USER is missing` — файл `.env` не создан или в нём не заполнен логин.

### Шаг 12. Запуск

```bash
docker compose pull      # скачать образы с Docker Hub
docker compose up -d     # запустить в фоне
docker compose ps        # проверить статус
```

В выводе `docker compose ps` у обоих сервисов должен быть статус `Up` (у Flask-версии через полминуты появится `(healthy)`).

Нужна только одна версия? Укажите её имя: `docker compose up -d api-go`.

### Шаг 13. Проверка

**В браузере** откройте:

- `http://203.0.113.10:5000/docs` — Flask-версия;
- `http://203.0.113.10:5001/docs` — Go-версия.

Должна открыться страница Swagger UI, где каждый эндпоинт можно вызвать кнопкой **Execute**.

**Скриптом проверки** со своего компьютера:

```powershell
py scripts/check_api.py --base-url http://203.0.113.10:5001
py scripts/check_api.py --base-url http://203.0.113.10:5000
```

Все 16 проверок должны быть `OK`. Скрипт создаёт тестовых пользователей — на сервере с настоящими данными учитывайте это.

**Логи** на сервере:

```bash
docker compose logs -f api-go     # Ctrl+C — выйти из просмотра
docker compose logs --tail 50     # последние 50 строк обоих сервисов
```

Готово: API работает и будет автоматически перезапускаться после сбоев и перезагрузки сервера (`restart: unless-stopped`). Логи ограничены 30 МБ на сервис (3 файла по 10 МБ), так что диск ими не переполнится.

---

## Часть 4. Обновление, откат и резервные копии

### Выпуск новой версии

1. Внесите изменения в код и проверьте их локально (`docker compose up -d --build`, `py scripts/check_api.py`).
2. Опубликуйте новую версию, увеличив номер:

   ```powershell
   powershell -ExecutionPolicy Bypass -File scripts\publish.ps1 -User yourname -Version 1.0.1
   ```

   Правило нумерации `MAJOR.MINOR.PATCH`: исправление ошибки — `1.0.1`, новая возможность — `1.1.0`, несовместимое изменение API — `2.0.0`.
3. На сервере:

   ```bash
   cd ~/users-api
   sed -i 's/^TAG=.*/TAG=1.0.1/' .env    # или поменяйте TAG в nano .env
   docker compose pull
   docker compose up -d
   ```

   Compose пересоздаст только те контейнеры, у которых изменился образ. Данные в базе сохраняются. Перерыв в работе — несколько секунд.
4. Удалите старые образы, чтобы они не занимали диск:

   ```bash
   docker image prune -f
   ```

### Откат на предыдущую версию

Если новая версия работает плохо — верните старый номер:

```bash
sed -i 's/^TAG=.*/TAG=1.0.0/' .env
docker compose up -d
```

Именно поэтому в `.env` лучше указывать конкретную версию, а не `latest`: всегда понятно, что запущено, и откат занимает секунды.

### Резервная копия базы

Данные хранятся в Docker-томах `users-api_python-data` и `users-api_go-data`. Копия Go-версии (для Flask замените `go` на `python`):

```bash
cd ~/users-api
docker compose stop api-go
docker run --rm -v users-api_go-data:/data -v "$PWD":/backup alpine \
  tar czf /backup/go-data-$(date +%F).tar.gz -C /data .
docker compose start api-go
```

Сервис останавливается на пару секунд, чтобы база не изменилась во время копирования. В папке появится файл вроде `go-data-2026-10-13.tar.gz`. Скачайте его на свой компьютер:

```powershell
scp root@203.0.113.10:~/users-api/go-data-2026-10-13.tar.gz .
```

Восстановление из копии:

```bash
docker compose stop api-go
docker run --rm -v users-api_go-data:/data -v "$PWD":/backup alpine sh -c \
  "rm -rf /data/* && tar xzf /backup/go-data-2026-10-13.tar.gz -C /data && chown -R 65532:65532 /data"
docker compose start api-go
```

`chown` возвращает файлам владельца, от имени которого работает контейнер: `65532` — для Go-версии, `10001` — для Flask-версии.

---

## Часть 5. Домен и HTTPS (по желанию)

Открывать API по IP и порту удобно для проверки, но для постоянной работы лучше домен с HTTPS. Проще всего это сделать с веб-сервером [Caddy](https://caddyserver.com): он сам получает и продлевает бесплатные сертификаты Let's Encrypt.

1. **DNS.** У регистратора домена создайте A-запись: `api.example.com` → `203.0.113.10`. Обновление DNS занимает от минут до нескольких часов; проверить можно командой `ping api.example.com`.
2. **Закройте прямой доступ к портам 5000/5001.** В `~/users-api/.env` поставьте `BIND_ADDR=127.0.0.1` и примените: `docker compose up -d`.
3. **Установите Caddy** (команды из [официальной документации](https://caddyserver.com/docs/install#debian-ubuntu-raspbian)):

   ```bash
   apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
   apt update && apt install -y caddy
   ```

4. **Настройте Caddy.** Откройте `nano /etc/caddy/Caddyfile`, удалите всё содержимое и вставьте:

   ```
   api.example.com {
       reverse_proxy 127.0.0.1:5001
   }
   ```

   Для второй версии можно добавить ещё один поддомен (`api-python.example.com` → `127.0.0.1:5000`, ему тоже нужна A-запись).
5. **Откройте порты для HTTP/HTTPS и перезапустите Caddy:**

   ```bash
   ufw allow 80,443/tcp
   systemctl reload caddy
   ```

Через минуту API будет доступен по `https://api.example.com/docs`.

---

## Шпаргалка команд на сервере

Все команды выполняются из папки `~/users-api`.

| Команда | Что делает |
|---------|------------|
| `docker compose ps` | Статус контейнеров |
| `docker compose logs -f api-go` | Логи в реальном времени |
| `docker compose restart` | Перезапуск обоих сервисов |
| `docker compose stop` / `start` | Остановить / запустить без удаления |
| `docker compose pull && docker compose up -d` | Обновиться до версии из `TAG` |
| `docker compose down` | Остановить и удалить контейнеры (данные сохранятся) |
| `docker compose down -v` | **Удалить всё, включая базы данных** |
| `docker stats` | Потребление памяти и процессора |
| `docker system df` | Сколько места занимает Docker |
| `docker image prune -f` | Удалить неиспользуемые образы |

## Частые проблемы

| Симптом | Причина и решение |
|---------|-------------------|
| `denied: requested access to the resource is denied` при публикации | Не выполнен `docker login`, у токена нет прав Read & Write, или в `-User` указан чужой логин |
| `unauthorized` / `pull access denied` на сервере | Репозиторий приватный — выполните `docker login` на сервере; или опечатка в `DOCKERHUB_USER` |
| `manifest unknown` / `not found` при `docker compose pull` | Такой версии нет на Docker Hub: проверьте `TAG` в `.env` и вкладку Tags на Docker Hub |
| `toomanyrequests: You have reached your pull rate limit` | Лимит анонимных скачиваний Docker Hub. Выполните `docker login` на сервере — для авторизованных лимит выше |
| `exec format error` в логах контейнера | Образ собран не для того процессора. Проверьте архитектуру сервера (`uname -m`: `x86_64` — amd64, `aarch64` — arm64) и опубликуйте образ с нужной `-Platforms` |
| `Bind for 0.0.0.0:5000 failed: port is already allocated` | Порт занят другой программой. Поменяйте `PYTHON_PORT`/`GO_PORT` в `.env` |
| `required variable DOCKERHUB_USER is missing a value` | Нет файла `.env` в текущей папке или он не заполнен (шаг 11) |
| Страница не открывается по IP, но `curl http://localhost:5001/active` на сервере работает | Порт закрыт файрволом провайдера (в его панели управления) или `BIND_ADDR=127.0.0.1` |
| Страница `/docs` открывается, но пустая | Браузер не смог загрузить Swagger UI с CDN jsDelivr. API при этом работает |
| `Execution of scripts is disabled on this system` при запуске `publish.ps1` | Запускайте через `powershell -ExecutionPolicy Bypass -File ...`, как в шаге 4 |

## Безопасность

- **В API нет авторизации**: любой, кто знает адрес, может создавать пользователей. Для учебного проекта это допустимо, для реальных данных — нет. Минимум — закрыть прямой доступ (`BIND_ADDR=127.0.0.1`) и поставить перед API Caddy с [базовой авторизацией](https://caddyserver.com/docs/caddyfile/directives/basic_auth).
- Используйте вход на сервер по SSH-ключу и регулярно обновляйте систему: `apt update && apt upgrade -y`.
- Токены Docker Hub храните как пароли. Если токен утёк — удалите его в **Account settings → Personal access tokens**.
- Контейнеры работают не от root, а Go-образ не содержит даже командной оболочки — это уменьшает последствия возможного взлома.
