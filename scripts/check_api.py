"""
Проверка всех эндпоинтов Users API (Flask- или Go-версии).

Запуск:
    py scripts/check_api.py                              # Windows, http://127.0.0.1:5000
    python3 scripts/check_api.py --base-url http://my-server:5001

Код выхода: 0 — все проверки пройдены, 1 — есть ошибки, 2 — сервер недоступен.
Скрипт создаёт тестовых пользователей и добавляет отметки активности.
"""

import argparse
import json
import sys
import time
import urllib.error
import urllib.request
import uuid

MISSING_ID = 9_223_372_036_854_775_807
MAX_NAME_LENGTH = 100
MAX_ACTIVE = 5


class Client:
    def __init__(self, base_url: str, timeout: float):
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    def request(self, method, path, body=None, raw=None):
        """Возвращает (статус, JSON или None). body сериализуется в JSON, raw отправляется как есть."""
        data = raw.encode("utf-8") if raw is not None else None
        if body is not None:
            data = json.dumps(body).encode("utf-8")
        headers = {"Content-Type": "application/json; charset=utf-8"} if data is not None else {}
        req = urllib.request.Request(self.base_url + path, data=data, method=method, headers=headers)
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                return resp.status, _parse_json(resp.read(), resp.headers)
        except urllib.error.HTTPError as e:
            return e.code, _parse_json(e.read(), e.headers)

    def fetch(self, path):
        """GET с переходом по редиректам. Возвращает (статус, Content-Type, текст, итоговый URL)."""
        try:
            with urllib.request.urlopen(self.base_url + path, timeout=self.timeout) as resp:
                text = resp.read().decode("utf-8")
                return resp.status, resp.headers.get("Content-Type", ""), text, resp.url
        except urllib.error.HTTPError as e:
            return e.code, e.headers.get("Content-Type", ""), "", e.url

    def wait_until_ready(self, seconds: float) -> bool:
        deadline = time.monotonic() + seconds
        while True:
            try:
                self.request("GET", "/active")
                return True
            except (urllib.error.URLError, ConnectionError, TimeoutError):
                if time.monotonic() >= deadline:
                    return False
                time.sleep(0.5)


def _parse_json(payload: bytes, headers):
    if "application/json" not in (headers.get("Content-Type") or ""):
        return None
    return json.loads(payload.decode("utf-8"))


def expect(cond, message):
    if not cond:
        raise AssertionError(message)


def expect_status(got, want, body):
    expect(got == want, f"ожидался статус {want}, получен {got}, тело: {body}")


def expect_error(status, body, want_status):
    expect_status(status, want_status, body)
    expect(isinstance(body, dict) and isinstance(body.get("error"), str),
           f"ожидалось тело {{\"error\": \"...\"}}, получено: {body}")


def unique(prefix):
    return f"{prefix}-{uuid.uuid4().hex[:8]}"


def run_checks(api: Client):
    state = {}

    def add_user():
        name = unique("check")
        status, body = api.request("POST", "/adduser", {"name": name})
        expect_status(status, 201, body)
        expect(isinstance(body.get("id"), int), f"'id' должен быть целым числом: {body}")
        expect(body.get("name") == name, f"ожидалось имя {name!r}: {body}")
        state["user"] = body

    def add_user_trims_and_unicode():
        name = unique("Алиса")
        status, body = api.request("POST", "/adduser", {"name": f"  {name}\t "})
        expect_status(status, 201, body)
        expect(body.get("name") == name, f"пробелы по краям не обрезаны или искажена кириллица: {body}")

    def add_user_max_length():
        name = "я" * MAX_NAME_LENGTH
        status, body = api.request("POST", "/adduser", {"name": name})
        expect_status(status, 201, body)
        expect(body.get("name") == name, "имя максимальной длины сохранено неверно")

    def add_user_sql_injection():
        name = unique("x'); DROP TABLE users;--")
        status, body = api.request("POST", "/adduser", {"name": name})
        expect_status(status, 201, body)
        status, body = api.request("GET", f"/user/{body['id']}")
        expect_status(status, 200, body)
        expect(body.get("name") == name, f"имя с SQL-кодом должно сохраниться как текст: {body}")

    def add_user_validation():
        cases = {
            "тело не JSON": dict(raw="junk"),
            "JSON-массив": dict(body=[]),
            "JSON null": dict(raw="null"),
            "нет поля name": dict(body={}),
            "name = null": dict(body={"name": None}),
            "пустое name": dict(body={"name": ""}),
            "name из пробелов": dict(body={"name": "   "}),
            "name — число": dict(body={"name": 5}),
            f"name длиннее {MAX_NAME_LENGTH}": dict(body={"name": "я" * (MAX_NAME_LENGTH + 1)}),
        }
        failed = []
        for label, kwargs in cases.items():
            status, body = api.request("POST", "/adduser", **kwargs)
            if status != 400 or not isinstance(body, dict) or "error" not in body:
                failed.append(f"{label}: {status} {body}")
        expect(not failed, "ожидался 400 с JSON-ошибкой:\n      " + "\n      ".join(failed))

    def add_user_wrong_method():
        status, _ = api.request("GET", "/adduser")
        expect_status(status, 405, None)

    def get_user():
        user = state["user"]
        status, body = api.request("GET", f"/user/{user['id']}")
        expect_status(status, 200, body)
        expect(body == user, f"ожидалось {user}, получено {body}")

    def get_user_missing():
        status, body = api.request("GET", f"/user/{MISSING_ID}")
        expect_error(status, body, 404)

    def get_user_bad_id():
        for bad in ("abc", "1%20OR%201=1"):
            status, body = api.request("GET", f"/user/{bad}")
            expect_status(status, 404, body)

    def set_active():
        uid = state["user"]["id"]
        status, body = api.request("POST", f"/user/{uid}/active")
        expect_status(status, 200, body)
        active = body.get("active")
        expect(isinstance(active, list) and 1 <= len(active) <= MAX_ACTIVE, f"неверный список: {body}")
        expect(active[-1] == uid, f"последней должна быть отметка {uid}: {active}")

    def set_active_missing():
        status, body = api.request("POST", f"/user/{MISSING_ID}/active")
        expect_error(status, body, 404)

    def set_active_evicts_oldest():
        ids = []
        for _ in range(2):
            status, body = api.request("POST", "/adduser", {"name": unique("active")})
            expect_status(status, 201, body)
            ids.append(body["id"])
        marks = [ids[i % 2] for i in range(MAX_ACTIVE + 1)]
        for uid in marks:
            status, body = api.request("POST", f"/user/{uid}/active")
            expect_status(status, 200, body)
        expect(body["active"] == marks[-MAX_ACTIVE:],
               f"ожидались последние {MAX_ACTIVE} отметок {marks[-MAX_ACTIVE:]}, получено {body['active']}")
        state["active"] = body["active"]

    def get_active():
        status, body = api.request("GET", "/active")
        expect_status(status, 200, body)
        active = body.get("active")
        expect(isinstance(active, list) and len(active) <= MAX_ACTIVE, f"неверный список: {body}")
        expect(all(isinstance(i, int) for i in active), f"элементы должны быть целыми: {active}")
        expect(active == state["active"], f"ожидалось {state['active']}, получено {active}")

    def docs_page():
        status, content_type, text, _ = api.fetch("/docs")
        expect_status(status, 200, None)
        expect(content_type.startswith("text/html"), f"ожидался text/html, получен {content_type!r}")
        expect("SwaggerUIBundle" in text and "/openapi.yaml" in text, "на странице нет Swagger UI")

    def openapi_spec():
        status, content_type, text, _ = api.fetch("/openapi.yaml")
        expect_status(status, 200, None)
        expect(content_type.startswith("application/yaml"), f"ожидался application/yaml, получен {content_type!r}")
        expect(text.lstrip().startswith("openapi: 3.1"), "файл не похож на спецификацию OpenAPI 3.1")
        for path in ("/adduser:", "/user/{id}:", "/user/{id}/active:", "/active:"):
            expect(path in text, f"в спецификации нет пути {path[:-1]}")

    def root_redirect():
        status, _, _, url = api.fetch("/")
        expect_status(status, 200, None)
        expect(url.rstrip("/").endswith("/docs"), f"ожидался переход на /docs, итоговый адрес {url}")

    return [
        ("GET  /docs — страница Swagger UI", docs_page),
        ("GET  /openapi.yaml — спецификация", openapi_spec),
        ("GET  / — перенаправление на /docs", root_redirect),
        ("POST /adduser — создание пользователя", add_user),
        ("POST /adduser — обрезка пробелов и кириллица", add_user_trims_and_unicode),
        (f"POST /adduser — имя длиной {MAX_NAME_LENGTH} символов", add_user_max_length),
        ("POST /adduser — защита от SQL-инъекции", add_user_sql_injection),
        ("POST /adduser — валидация (9 некорректных запросов)", add_user_validation),
        ("GET  /adduser — неподдерживаемый метод -> 405", add_user_wrong_method),
        ("GET  /user/{id} — существующий пользователь", get_user),
        ("GET  /user/{id} — несуществующий -> 404", get_user_missing),
        ("GET  /user/{id} — нецелый id -> 404", get_user_bad_id),
        ("POST /user/{id}/active — отметка активности", set_active),
        ("POST /user/{id}/active — несуществующий -> 404", set_active_missing),
        (f"POST /user/{{id}}/active — хранятся последние {MAX_ACTIVE}", set_active_evicts_oldest),
        ("GET  /active — список активных", get_active),
    ]


def main():
    parser = argparse.ArgumentParser(description="Проверка эндпоинтов Users API")
    parser.add_argument("--base-url", default="http://127.0.0.1:5000", help="адрес сервера (по умолчанию %(default)s)")
    parser.add_argument("--timeout", type=float, default=5, help="таймаут одного запроса, секунд")
    parser.add_argument("--wait", type=float, default=15, help="сколько секунд ждать запуска сервера")
    args = parser.parse_args()
    # A redirected stdout on Windows uses the ANSI code page, which lacks some characters.
    sys.stdout.reconfigure(errors="replace")

    api = Client(args.base_url, args.timeout)
    print(f"Сервер: {api.base_url}")
    if not api.wait_until_ready(args.wait):
        print(f"Сервер недоступен (ждали {args.wait:g} с). Он запущен? Верный ли адрес?")
        return 2

    passed = failed = 0
    for title, check in run_checks(api):
        try:
            check()
        except Exception as e:
            failed += 1
            print(f"  FAIL  {title}\n        {e}")
        else:
            passed += 1
            print(f"  OK    {title}")

    print(f"\nИтого: {passed} пройдено, {failed} с ошибками")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
