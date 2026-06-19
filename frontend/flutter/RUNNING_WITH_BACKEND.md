# Running the Flutter app against the backend

How to run the Go backend and the Flutter client together to test the REST
auth integration end-to-end (`USE_API_AUTH=true`).

## Prerequisites

- Docker + Docker Compose
- Flutter SDK
- `just` (optional, recommended) — or use the raw `docker compose` commands

## 1. Start the backend stack

From the repo root:

```bash
just up
```

This builds and starts all services, waits for the gateway to be ready, and
**seeds dev accounts + languages**. Without `just`:

```bash
cp -n .env.example .env        # first run only
docker compose up --build -d
# wait until ready:
curl -sf http://localhost:8080/api/v1/ready
```

Verify the gateway is up:

```bash
curl http://localhost:8080/api/v1/health      # {"status":"ok",...}
```

Gateway base URL: **`http://localhost:8080/api/v1`**.

## 2. Seeded dev accounts

All with password **`password123`**:

| Role    | Email                |
|---------|----------------------|
| Teacher | `teacher@even.local` |
| Student | `student@even.local` |
| Admin   | `admin@even.local`   |

## 3. Run the Flutter app with the API-auth flag

From `frontend/flutter`:

```bash
flutter pub get        # first run in this folder

flutter run -d chrome \
  --dart-define=USE_API_AUTH=true \
  --dart-define=API_BASE_URL=http://localhost:8080/api/v1
```

- `USE_API_AUTH=true` switches auth from Firebase to the REST backend.
- `API_BASE_URL` is optional (this is the default). CORS allows all origins in
  dev, so any Flutter web port works.
- If the Chrome debugger connection hangs, use a plain web server instead:
  `flutter run -d web-server --web-port 8095 ...` and open the printed URL.

## 4. What to test

- **Login** with `teacher@even.local` / `password123` → Teacher dashboard.
- **Login** with `student@even.local` → Student app (role routing).
- **Register** a new account → lands signed-in as a student.
- **Sign out** (Profile → Sign Out) → returns to login, tokens cleared.
- **Session restore**: refresh the page while logged in → stays logged in
  (token in secure storage, `GET /auth/me`).
- **Token refresh**: handled automatically on 401.

## Known limitations (transitional state)

Only **auth** is wired to REST. Screens still backed by Firebase (Home/courses,
profile details, teacher classes) will be empty or error out under
`USE_API_AUTH=true` — that's expected until those screens are migrated to
`/courses`, `/teacher/*`, etc. Run **without** the flag for the current
Firebase behaviour.
