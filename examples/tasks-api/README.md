# tasks-api

ตัวอย่าง REST API สำหรับจัดการ task ของผู้ใช้ ประกอบบท "สร้าง Web API" ของเว็บ Go ภาษาไทย

- Go `net/http` (routing แบบ Go 1.22+) ไม่ใช้ framework
- PostgreSQL 18 + [pgx](https://github.com/jackc/pgx) + [sqlc](https://sqlc.dev) + [goose](https://github.com/pressly/goose) migrations
- JWT (HS256) + bcrypt, rate limit ที่ endpoint auth
- Response envelope, error format, cursor pagination, request ID, structured logging (`log/slog`)
- Graceful shutdown, server timeouts, health/readiness check
- Unit tests + integration test กับ PostgreSQL จริง
- OpenAPI 3.1 (`api/openapi.yaml`, เสิร์ฟที่ `GET /openapi.yaml`)

## รันด้วย Docker

```bash
docker compose up --build -d
curl localhost:8080/healthz
```

PostgreSQL เปิดที่ host port `5433` และ API ที่ `8080` เปลี่ยนได้ด้วย `DB_HOST_PORT` / `API_HOST_PORT`

## รันบนเครื่อง

```bash
cp .env.example .env
make db-up      # PostgreSQL ใน Docker
make run
```

## Endpoints

| Method | Path | Auth | คำอธิบาย |
|---|---|---|---|
| GET | `/healthz` | – | liveness |
| GET | `/readyz` | – | readiness (ping DB) |
| GET | `/openapi.yaml` | – | OpenAPI spec |
| POST | `/v1/auth/register` | – | สมัครสมาชิก |
| POST | `/v1/auth/login` | – | รับ access token |
| GET | `/v1/me` | Bearer | ข้อมูลผู้ใช้ |
| GET | `/v1/tasks` | Bearer | รายการ task (`status`, `limit`, `cursor`) |
| POST | `/v1/tasks` | Bearer | สร้าง task |
| GET | `/v1/tasks/{id}` | Bearer | ดู task |
| PATCH | `/v1/tasks/{id}` | Bearer | แก้บาง field |
| DELETE | `/v1/tasks/{id}` | Bearer | ลบ task |

## ลองเรียก

```bash
B=http://localhost:8080
curl -s -X POST $B/v1/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"somchai@example.com","name":"สมชาย","password":"correct-horse"}'

TOKEN=$(curl -s -X POST $B/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"somchai@example.com","password":"correct-horse"}' | jq -r .data.access_token)

curl -s -X POST $B/v1/tasks -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"title":"เรียน Go"}'
curl -s "$B/v1/tasks?limit=10" -H "Authorization: Bearer $TOKEN"
```

## Tests

```bash
make test               # unit tests (-race)
make db-up
make test-integration   # รวม integration test กับ PostgreSQL
```

## แก้ query

แก้ไฟล์ใน `internal/postgres/queries/` แล้วรัน `make sqlc` (ใช้ sqlc ผ่าน Docker) เพื่อ generate code ใน `internal/postgres/sqlc/` ใหม่
