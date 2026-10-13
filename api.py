import sqlite3
import threading
from collections import deque
from pathlib import Path

from flask import Flask, g, jsonify, redirect, request, send_file

BASE_DIR = Path(__file__).resolve().parent
DB_PATH = "test.db"
MAX_NAME_LENGTH = 100
MAX_ACTIVE_USERS = 5

app = Flask(__name__)

_active_users: deque = deque(maxlen=MAX_ACTIVE_USERS)
_active_users_lock = threading.Lock()


def get_db() -> sqlite3.Connection:
    if "db" not in g:
        g.db = sqlite3.connect(DB_PATH)
    return g.db


@app.teardown_appcontext
def close_db(exc) -> None:
    db = g.pop("db", None)
    if db is not None:
        db.close()


def init_db() -> None:
    with app.app_context():
        db = get_db()
        db.execute("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)")
        db.commit()


def error(message: str, status: int):
    return jsonify({"error": message}), status


@app.post("/adduser")
def add_user():
    data = request.get_json(silent=True)
    if not isinstance(data, dict):
        return error("request body must be a JSON object", 400)

    name = data.get("name")
    if not isinstance(name, str) or not name.strip():
        return error("'name' must be a non-empty string", 400)
    name = name.strip()
    if len(name) > MAX_NAME_LENGTH:
        return error(f"'name' must be at most {MAX_NAME_LENGTH} characters", 400)

    db = get_db()
    with db:
        cur = db.execute("INSERT INTO users (name) VALUES (?)", (name,))
    return jsonify({"id": cur.lastrowid, "name": name}), 201


@app.get("/user/<int:uid>")
def get_user(uid: int):
    row = get_db().execute("SELECT id, name FROM users WHERE id = ?", (uid,)).fetchone()
    if row is None:
        return error("user not found", 404)
    return jsonify({"id": row[0], "name": row[1]}), 200


@app.post("/user/<int:uid>/active")
def set_active(uid: int):
    exists = get_db().execute("SELECT 1 FROM users WHERE id = ?", (uid,)).fetchone()
    if exists is None:
        return error("user not found", 404)
    with _active_users_lock:
        _active_users.append(uid)
        active = list(_active_users)
    return jsonify({"active": active}), 200


@app.get("/active")
def get_active():
    with _active_users_lock:
        active = list(_active_users)
    return jsonify({"active": active}), 200


@app.get("/")
def index():
    return redirect("/docs")


@app.get("/docs")
def docs():
    return send_file(BASE_DIR / "docs.html", mimetype="text/html")


@app.get("/openapi.yaml")
def openapi_spec():
    return send_file(BASE_DIR / "openapi.yaml", mimetype="application/yaml")


init_db()

if __name__ == "__main__":
    app.run()
