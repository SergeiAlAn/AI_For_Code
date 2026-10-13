import hashlib
import hmac
import json
import os
import sqlite3
import threading
from collections import deque
from contextlib import closing
from typing import Iterable, Optional

DB_PATH = "users.db"
PASSWORDS_PATH = "passwords.txt"

PBKDF2_ITERATIONS = 600_000
SALT_SIZE = 16

MAX_ACTIVE_USERS = 5
_active_users: deque = deque(maxlen=MAX_ACTIVE_USERS)
_active_users_lock = threading.Lock()


def _connect() -> sqlite3.Connection:
    conn = sqlite3.connect(DB_PATH)
    conn.execute(
        "CREATE TABLE IF NOT EXISTS users ("
        "id INTEGER PRIMARY KEY AUTOINCREMENT, "
        "name TEXT NOT NULL, "
        "tags TEXT NOT NULL DEFAULT '[]')"
    )
    return conn


def add_user(name: str, tags: Optional[Iterable[str]] = None) -> int:
    """
    Добавляет пользователя в локальную базу.
    К тегам автоматически добавляется "new"; переданная коллекция не изменяется.
    Возвращает id созданной записи.
    """
    user_tags = list(tags or [])
    user_tags.append("new")
    with closing(_connect()) as conn, conn:
        cur = conn.execute(
            "INSERT INTO users (name, tags) VALUES (?, ?)",
            (name, json.dumps(user_tags)),
        )
        return cur.lastrowid


def _hash_password(password: str, salt: bytes) -> bytes:
    return hashlib.pbkdf2_hmac("sha256", password.encode("utf-8"), salt, PBKDF2_ITERATIONS)


def store_password(user_id: int, password: str) -> None:
    """Сохраняет солёный PBKDF2-хеш пароля в формате `user_id:salt_hex:hash_hex`."""
    salt = os.urandom(SALT_SIZE)
    hashed = _hash_password(password, salt)
    with open(PASSWORDS_PATH, "a", encoding="utf-8") as f:
        f.write(f"{user_id}:{salt.hex()}:{hashed.hex()}\n")


def verify_password(user_id: int, password: str) -> bool:
    """Проверяет пароль по последней сохранённой для пользователя записи."""
    record = None
    try:
        with open(PASSWORDS_PATH, encoding="utf-8") as f:
            for line in f:
                uid, salt_hex, hash_hex = line.rstrip("\n").split(":")
                if uid == str(user_id):
                    record = (bytes.fromhex(salt_hex), bytes.fromhex(hash_hex))
    except FileNotFoundError:
        return False
    if record is None:
        return False
    salt, expected = record
    return hmac.compare_digest(_hash_password(password, salt), expected)


def set_active(user_id: int) -> None:
    """Помечает пользователя активным; хранятся только последние MAX_ACTIVE_USERS."""
    with _active_users_lock:
        _active_users.append(user_id)


def get_active_users() -> list:
    with _active_users_lock:
        return list(_active_users)


def get_user_by_name(name: str) -> Optional[dict]:
    """Возвращает пользователя как dict или None, если он не найден."""
    with closing(_connect()) as conn:
        row = conn.execute(
            "SELECT id, name, tags FROM users WHERE name = ?", (name,)
        ).fetchone()
    if row is None:
        return None
    uid, uname, tags_json = row
    try:
        tags = json.loads(tags_json) if tags_json else []
    except json.JSONDecodeError:
        tags = []
    return {"id": uid, "name": uname, "tags": tags}
