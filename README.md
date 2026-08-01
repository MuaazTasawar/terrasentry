# TerraSentry

> AI-powered infrastructure governance with a mobile-first approval loop â€” catch risky Terraform changes and live cluster drift before they bite you.

## Overview

TerraSentry is an infrastructure governance copilot built for platform and DevOps teams who need risk visibility without living in a dashboard. It scans Terraform plans pre-apply for cost, security, and drift risk using an LLM, runs a lightweight Kubernetes controller that continuously watches live cluster state for configuration drift, and â€” instead of routing everything through a web UI â€” pushes risky changes straight to an on-call engineer's phone as a one-tap approve/reject decision. It exists because the existing category (Spacelift, env0, Firefly) assumes the reviewer is sitting at a laptop; TerraSentry assumes they're not, and closes that gap with a genuinely mobile-first approval loop.

v2 adds a read-focused Next.js dashboard for the rest of the team, real JWT authentication in front of the endpoints that can change state, drift detection for StatefulSets and ConfigMaps (not just Deployments), and a declarative policy layer that lets a team encode hard rules ("IAM changes are always at least medium risk") on top of the LLM's judgment.

## Features

- **Pre-apply Terraform risk scoring** â€” submit a `terraform show -json` plan and get back an LLM-generated risk score (0â€“100), a risk tier (low/medium/high), human-readable reasoning, and a list of specifically flagged resources.
- **Deterministic threshold safety net** â€” the LLM's risk score is re-checked against configurable numeric thresholds server-side, so the final low/medium/high classification never depends on the model's own labeling alone.
- **Policy-as-code floor** â€” a version-controlled `policy.yaml` layers hard rules on top of the LLM/threshold result (e.g. "any `aws_iam_*` change is at least medium", "deletes on `env=prod`-tagged resources are always high"). Rules can only raise the risk level, never lower it.
- **Live Kubernetes drift detection** â€” controller-runtime reconcilers watch Deployments, StatefulSets, and ConfigMaps, snapshot their key spec fields, and detect when live state diverges from the last known-good baseline â€” catching unauthorized `kubectl edit`/`kubectl scale` changes that bypass Terraform entirely.
- **Mobile-first approval flow** â€” medium/high risk scans trigger a push notification to every registered on-call device; the Flutter app shows a home feed of pending approvals and a detail screen with full reasoning, plan summary, and one-tap Approve/Reject actions.
- **Team-wide dashboard** â€” a Next.js dashboard gives the rest of the team a read-only view of scan history (filterable by risk/status), drift events, and the full approval audit trail, without needing database access.
- **JWT authentication** â€” state-changing endpoints (approve/reject a scan, register a device) require a valid Bearer token; the dashboard and mobile app both authenticate against the same `/api/v1/auth/login` endpoint.
- **Full audit trail** â€” every approval or rejection is recorded with who decided, what they decided, and when, stored independently of the scan record itself.
- **Fail-safe LLM parsing** â€” if the LLM ever returns a response that can't be parsed as valid JSON, the system defaults to flagging the change as high-risk for manual review rather than silently letting it through.
- **Zero-cost local development** â€” the entire stack (API, risk-scoring service, dashboard, Postgres, Kubernetes cluster) runs locally via Docker Compose and k3d, with no managed cloud services required to build, test, or demo the project.

## Tech Stack

| Layer | Technology |
|-------|------------|
| API, Auth & K8s Controllers | Go, Gin, pgx, controller-runtime, client-go, golang-jwt, bcrypt |
| AI Risk Scoring & Policy Engine | Python, FastAPI, Anthropic API, python-hcl2, httpx, PyYAML |
| Dashboard | Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS 4 |
| Mobile Approvals | Flutter, Dart, http, shared_preferences |
| Database | PostgreSQL 16 |
| Local Kubernetes | k3d (k3s in Docker) |
| Containerization | Docker, Docker Compose |
| Push Notifications | Firebase Cloud Messaging (FCM) |
| Testing | Go's built-in `testing` package, pytest + pytest-asyncio |

## Project Structure

```
terrasentry/
â”œâ”€â”€ api/
â”‚   â”œâ”€â”€ cmd/
â”‚   â”‚   â”œâ”€â”€ server/
â”‚   â”‚   â”‚   â””â”€â”€ main.go
â”‚   â”‚   â”œâ”€â”€ operator/
â”‚   â”‚   â”‚   â””â”€â”€ main.go
â”‚   â”‚   â””â”€â”€ seed-user/
â”‚   â”‚       â””â”€â”€ main.go
â”‚   â”œâ”€â”€ internal/
â”‚   â”‚   â”œâ”€â”€ api/
â”‚   â”‚   â”‚   â”œâ”€â”€ handlers.go
â”‚   â”‚   â”‚   â”œâ”€â”€ router.go
â”‚   â”‚   â”‚   â””â”€â”€ middleware.go
â”‚   â”‚   â”œâ”€â”€ auth/
â”‚   â”‚   â”‚   â”œâ”€â”€ jwt.go
â”‚   â”‚   â”‚   â””â”€â”€ jwt_test.go
â”‚   â”‚   â”œâ”€â”€ db/
â”‚   â”‚   â”‚   â”œâ”€â”€ db.go
â”‚   â”‚   â”‚   â””â”€â”€ models.go
â”‚   â”‚   â”œâ”€â”€ controller/
â”‚   â”‚   â”‚   â”œâ”€â”€ drift_controller.go
â”‚   â”‚   â”‚   â”œâ”€â”€ reconciler.go
â”‚   â”‚   â”‚   â”œâ”€â”€ reconciler_test.go
â”‚   â”‚   â”‚   â”œâ”€â”€ statefulset_reconciler.go
â”‚   â”‚   â”‚   â”œâ”€â”€ statefulset_reconciler_test.go
â”‚   â”‚   â”‚   â”œâ”€â”€ configmap_reconciler.go
â”‚   â”‚   â”‚   â””â”€â”€ configmap_reconciler_test.go
â”‚   â”‚   â”œâ”€â”€ notify/
â”‚   â”‚   â”‚   â””â”€â”€ push.go
â”‚   â”‚   â””â”€â”€ config/
â”‚   â”‚       â””â”€â”€ config.go
â”‚   â”œâ”€â”€ migrations/
â”‚   â”‚   â”œâ”€â”€ 001_init.sql
â”‚   â”‚   â”œâ”€â”€ 002_approvals.sql
â”‚   â”‚   â””â”€â”€ 003_users.sql
â”‚   â”œâ”€â”€ go.mod
â”‚   â”œâ”€â”€ go.sum
â”‚   â”œâ”€â”€ Dockerfile
â”‚   â””â”€â”€ .env.example
â”œâ”€â”€ risk-scoring/
â”‚   â”œâ”€â”€ app/
â”‚   â”‚   â”œâ”€â”€ main.py
â”‚   â”‚   â”œâ”€â”€ policy.yaml
â”‚   â”‚   â”œâ”€â”€ routers/
â”‚   â”‚   â”‚   â””â”€â”€ scan.py
â”‚   â”‚   â”œâ”€â”€ services/
â”‚   â”‚   â”‚   â”œâ”€â”€ llm_client.py
â”‚   â”‚   â”‚   â”œâ”€â”€ terraform_parser.py
â”‚   â”‚   â”‚   â”œâ”€â”€ risk_engine.py
â”‚   â”‚   â”‚   â””â”€â”€ policy_engine.py
â”‚   â”‚   â”œâ”€â”€ models/
â”‚   â”‚   â”‚   â””â”€â”€ schemas.py
â”‚   â”‚   â””â”€â”€ config.py
â”‚   â”œâ”€â”€ tests/
â”‚   â”‚   â”œâ”€â”€ __init__.py
â”‚   â”‚   â”œâ”€â”€ test_risk_engine.py
â”‚   â”‚   â”œâ”€â”€ test_terraform_parser.py
â”‚   â”‚   â””â”€â”€ test_policy_engine.py
â”‚   â”œâ”€â”€ requirements.txt
â”‚   â”œâ”€â”€ pytest.ini
â”‚   â”œâ”€â”€ Dockerfile
â”‚   â””â”€â”€ .env.example
â”œâ”€â”€ dashboard/
â”‚   â”œâ”€â”€ src/
â”‚   â”‚   â”œâ”€â”€ app/
â”‚   â”‚   â”‚   â”œâ”€â”€ layout.tsx
â”‚   â”‚   â”‚   â”œâ”€â”€ page.tsx
â”‚   â”‚   â”‚   â”œâ”€â”€ globals.css
â”‚   â”‚   â”‚   â”œâ”€â”€ login/
â”‚   â”‚   â”‚   â”‚   â””â”€â”€ page.tsx
â”‚   â”‚   â”‚   â”œâ”€â”€ scans/
â”‚   â”‚   â”‚   â”‚   â””â”€â”€ page.tsx
â”‚   â”‚   â”‚   â”œâ”€â”€ drift/
â”‚   â”‚   â”‚   â”‚   â””â”€â”€ page.tsx
â”‚   â”‚   â”‚   â””â”€â”€ audit/
â”‚   â”‚   â”‚       â””â”€â”€ page.tsx
â”‚   â”‚   â”œâ”€â”€ components/
â”‚   â”‚   â”‚   â”œâ”€â”€ TopNav.tsx
â”‚   â”‚   â”‚   â”œâ”€â”€ RiskBadge.tsx
â”‚   â”‚   â”‚   â”œâ”€â”€ ScanTable.tsx
â”‚   â”‚   â”‚   â”œâ”€â”€ DriftTable.tsx
â”‚   â”‚   â”‚   â””â”€â”€ AuditTable.tsx
â”‚   â”‚   â””â”€â”€ lib/
â”‚   â”‚       â””â”€â”€ apiClient.ts
â”‚   â”œâ”€â”€ package.json
â”‚   â”œâ”€â”€ next.config.ts
â”‚   â”œâ”€â”€ Dockerfile
â”‚   â””â”€â”€ .env.local.example
â”œâ”€â”€ mobile/
â”‚   â”œâ”€â”€ lib/
â”‚   â”‚   â”œâ”€â”€ main.dart
â”‚   â”‚   â”œâ”€â”€ models/
â”‚   â”‚   â”‚   â””â”€â”€ approval_request.dart
â”‚   â”‚   â”œâ”€â”€ screens/
â”‚   â”‚   â”‚   â”œâ”€â”€ login_screen.dart
â”‚   â”‚   â”‚   â”œâ”€â”€ home_screen.dart
â”‚   â”‚   â”‚   â””â”€â”€ approval_detail_screen.dart
â”‚   â”‚   â”œâ”€â”€ services/
â”‚   â”‚   â”‚   â””â”€â”€ api_service.dart
â”‚   â”‚   â””â”€â”€ widgets/
â”‚   â”‚       â””â”€â”€ approval_card.dart
â”‚   â”œâ”€â”€ pubspec.yaml
â”‚   â””â”€â”€ .env.example
â”œâ”€â”€ infra/
â”‚   â”œâ”€â”€ k3d-config.yaml
â”‚   â””â”€â”€ docker-compose.yml
â”œâ”€â”€ .gitignore
â””â”€â”€ README.md
```

## Architecture

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”     scan plan      â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  Terraform   â”‚ â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â–¶â”‚  Risk Scoring (Py)    â”‚
â”‚  CI Pipeline â”‚                    â”‚  FastAPI + LLM        â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                    â”‚  + Policy Engine       â”‚
                                    â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                                               â”‚ risk score (post-policy)
                                               â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   drift events   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   push notify    â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  K8s Operator  â”‚ â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â–¶â”‚   Go API + Postgres   â”‚ â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â–¶â”‚ Flutter App  â”‚
â”‚  (Deployment,  â”‚                  â”‚   (JWT-protected       â”‚â—€â”€â”€approve/rejectâ”€â”‚ (on-call,    â”‚
â”‚  StatefulSet,  â”‚                  â”‚    write endpoints)    â”‚                  â”‚  authed)     â”‚
â”‚  ConfigMap)    â”‚                  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                              â”‚ read-only, JWT-protected
                                                â–¼
                                    â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
                                    â”‚  Next.js Dashboard     â”‚
                                    â”‚  scans / drift / audit â”‚
                                    â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

**How a change flows through the system:**
1. A CI pipeline runs `terraform plan`, converts it to JSON (`terraform show -json`), and POSTs it to the risk-scoring service.
2. The risk-scoring service parses the plan into a compact resource-change summary, sends it to the LLM with a risk-analysis system prompt, and gets back a structured score. That score is re-derived against numeric thresholds server-side, then run through the policy engine (`policy.yaml`), which can only raise the final level, never lower it.
3. The Go API stores the scored scan in Postgres. If the risk is medium or high, it pushes an FCM notification to every registered on-call device.
4. Independently, the K8s operator continuously reconciles Deployments, StatefulSets, and ConfigMaps in the cluster, comparing live spec against a stored baseline snapshot, and writes any detected drift to the same database.
5. The Flutter app (now behind a login screen) polls the API for pending approvals, shows them in a risk-colored list, and lets the on-call engineer approve or reject with two taps â€” both of which require a valid JWT.
6. The Next.js dashboard gives the rest of the team a read-only, filterable view of the same scan history, drift events, and approval audit trail, also behind login.
7. Every decision is written to an audit table, independent of the scan record, for traceability.

## Getting Started

### Prerequisites

| Tool | Minimum Version | Notes |
|---|---|---|
| Go | 1.22+ | Builds the API server, K8s operator, and seed-user CLI |
| Python | 3.12+ | Runs the FastAPI risk-scoring service |
| Node.js | 20+ | Builds and runs the Next.js dashboard |
| Flutter | 3.3+ (Dart 3.3+) | Builds and runs the mobile approval app |
| Docker & Docker Compose | Recent stable | Runs Postgres and (optionally) containerized services |
| k3d | v5+ | Spins up a local Kubernetes cluster for drift detection (optional but needed to demo the operator) |
| kubectl | Any recent version | Interacts with the local k3d cluster |
| An Anthropic API key | â€” | Powers the LLM risk-scoring calls; free-tier alternatives (Groq, Gemini Flash, local Ollama) can be swapped in via `risk-scoring/app/services/llm_client.py` |

### Clone the Repo

```bash
git clone https://github.com/MuaazTasawar/terrasentry.git
cd terrasentry
```

### Installation

**1. Set up environment files**
```bash
cp api/.env.example api/.env
cp risk-scoring/.env.example risk-scoring/.env
cp dashboard/.env.local.example dashboard/.env.local
```
Then edit `risk-scoring/.env` and add a real `LLM_API_KEY`, and edit `api/.env` to set a real `JWT_SECRET` (any long random string â€” required, the server won't start without one). All other defaults work out of the box for local development.

**2. Install Go dependencies**
```bash
cd api
go mod tidy
cd ..
```

**3. Install Python dependencies**
```bash
cd risk-scoring
python -m venv venv
venv\Scripts\activate   # Windows
# source venv/bin/activate   # macOS/Linux
pip install -r requirements.txt
cd ..
```

**4. Install dashboard dependencies**
```bash
cd dashboard
npm install
cd ..
```

**5. Install Flutter dependencies**
```bash
cd mobile
flutter pub get
cd ..
```

### Running the App

**Start Postgres (and optionally the containerized API/risk-scoring/dashboard services):**
```bash
cd infra
docker compose up --build
```

**Run database migrations** (only needed once, or after pulling new migration files):
```bash
psql "postgres://terrasentry:terrasentry@localhost:5432/terrasentry" -f api/migrations/001_init.sql
psql "postgres://terrasentry:terrasentry@localhost:5432/terrasentry" -f api/migrations/002_approvals.sql
psql "postgres://terrasentry:terrasentry@localhost:5432/terrasentry" -f api/migrations/003_users.sql
```

**Create a login for yourself** (there's no signup UI on purpose â€” this is a single-command seed):
```bash
cd api
SEED_EMAIL=oncall@example.com SEED_PASSWORD=changeme go run ./cmd/seed-user
cd ..
```

**Run the Go API server directly (alternative to Docker, useful for active development):**
```bash
cd api
go run cmd/server/main.go
```
Server starts on `http://localhost:8080`. Verify with:
```bash
curl http://localhost:8080/health
```

**Run the risk-scoring service directly:**
```bash
cd risk-scoring
venv\Scripts\activate
uvicorn app.main:app --reload --port 8000
```
Verify with:
```bash
curl http://localhost:8000/health
```

**Run the dashboard directly:**
```bash
cd dashboard
npm run dev
```
Dashboard starts on `http://localhost:3000`. Log in with the credentials you created via `seed-user`.

**Spin up a local Kubernetes cluster and run the drift operator:**
```bash
k3d cluster create --config infra/k3d-config.yaml
cd api
go run cmd/operator/main.go
```
This now watches Deployments, StatefulSets, and ConfigMaps. Trigger drift on any of them by bypassing Terraform directly:
```bash
kubectl create deployment demo --image=nginx --replicas=2
kubectl scale deployment demo --replicas=5
```
Check for the resulting drift event:
```bash
curl http://localhost:8080/api/v1/drift-events
```

**Run the mobile app:**
```bash
cd mobile
flutter run
```
You'll land on a login screen first â€” use the same credentials created via `seed-user`. On an Android emulator, the API base URL in `mobile/lib/services/api_service.dart` should point to `http://10.0.2.2:8080` instead of `localhost`, since the emulator can't resolve the host machine's localhost directly. On a physical device on the same network, use your machine's LAN IP.

### Running Tests

**Go tests** (drift reconciler diff/snapshot logic across all three resource kinds, JWT generation/validation, edge cases like nil replicas and empty containers):
```bash
cd api
go test ./... -v
```

**Python tests** (risk-scoring threshold boundaries, Terraform plan parsing, policy engine rule matching, and LLM-vs-deterministic-threshold override logic â€” all run with the real LLM call mocked out, so they're free and fast):
```bash
cd risk-scoring
venv\Scripts\activate   # Windows
# source venv/bin/activate   # macOS/Linux
pytest -v
```

**Dashboard type-check and lint:**
```bash
cd dashboard
npx tsc --noEmit
npm run lint
```

All suites are pure unit/static-analysis checks â€” no live database, cluster, or API keys required to run them.

## Environment Variables

### `api/.env`

| Variable | Description | Where to get it |
|---|---|---|
| `PORT` | Port the Go API listens on | Default `8080`, change if needed |
| `ENV` | Environment name (`development`/`production`) | Set manually |
| `DATABASE_URL` | Postgres connection string | Matches `docker-compose.yml` credentials by default |
| `RISK_SCORING_URL` | Base URL of the Python risk-scoring service | Default `http://localhost:8000` |
| `JWT_SECRET` | Secret used to sign/verify auth tokens â€” **required**, server exits at startup if unset | Generate any long random string |
| `JWT_EXPIRY_HOURS` | How long an issued login token stays valid | Default `24` |
| `FCM_SERVER_KEY` | Firebase Cloud Messaging server key for push notifications | Firebase Console â†’ Project Settings â†’ Cloud Messaging |
| `KUBECONFIG_PATH` | Path to your local kubeconfig | Defaults to `~/.kube/config`, auto-detected by k3d |

### `risk-scoring/.env`

| Variable | Description | Where to get it |
|---|---|---|
| `PORT` | Port the FastAPI service listens on | Default `8000` |
| `ENV` | Environment name | Set manually |
| `LLM_API_KEY` | API key for the LLM provider | console.anthropic.com (or your chosen provider) |
| `LLM_MODEL` | Model identifier used for scoring calls | Default `claude-sonnet-4-6` |
| `RISK_SCORE_HIGH_THRESHOLD` | Numeric score at/above which a plan is classified high risk | Default `75`, tune to taste |
| `RISK_SCORE_MEDIUM_THRESHOLD` | Numeric score at/above which a plan is classified medium risk | Default `40`, tune to taste |
| `POLICY_FILE_PATH` | Path to the declarative policy rules file | Default `app/policy.yaml` |

### `dashboard/.env.local`

| Variable | Description | Where to get it |
|---|---|---|
| `NEXT_PUBLIC_API_BASE_URL` | Base URL of the Go API the dashboard calls | `http://localhost:8080` locally |
| `NEXT_PUBLIC_APP_NAME` | Display name shown in the dashboard header | Cosmetic, default `TerraSentry Dashboard` |

### Mobile app config

The Flutter app doesn't read a `.env` file -- its API base URL is a hardcoded
constructor default in `mobile/lib/services/api_service.dart`
(`ApiService({this.baseUrl = 'http://localhost:8080'})`). Edit that line
directly to point at a different host -- e.g. `http://10.0.2.2:8080` for an
Android emulator, or your machine's LAN IP for a physical device.

## API Reference

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | â€” | API server health check |
| `POST` | `/api/v1/auth/login` | â€” | Exchange email/password for a JWT (`{"token": "...", "expires_in_hours": 24}`) |
| `POST` | `/api/v1/scans` | â€” | Store a scored plan; triggers a push notification if risk is medium/high |
| `GET` | `/api/v1/scans/pending` | â€” | List scans currently awaiting a decision (used by the mobile home screen) |
| `GET` | `/api/v1/scans` | Bearer | Dashboard scan history, filterable via `?status=` and `?risk_level=` query params |
| `POST` | `/api/v1/scans/:id/decision` | Bearer | Approve or reject a scan (`{"decision": "approved" \| "rejected"}`) |
| `POST` | `/api/v1/devices` | Bearer | Register a device token for push notifications |
| `GET` | `/api/v1/drift-events` | â€” | List the 50 most recent Kubernetes drift events |
| `GET` | `/api/v1/audit` | Bearer | Full approve/reject decision trail, joined with the scan it belongs to |
| `GET` | `/health` (risk-scoring service, port 8000) | â€” | Risk-scoring service health check |
| `POST` | `/scan` (risk-scoring service, port 8000) | â€” | Score a raw `terraform show -json` plan, apply the policy floor, and return a risk assessment |

## Verified Working

This isn't just scaffolded code â€” the full loop has been manually tested end-to-end on a local environment:

- **Risk-scored approval flow**: a test scan was inserted with a high-risk score, correctly rendered in the Flutter app with the right risk-color coding, opened into the detail view with full reasoning and plan summary, and approved via the mobile UI. The decision was confirmed written back to the `approval_audit` table with the correct scan reference, decision, and timestamp.
- **Live Kubernetes drift detection**: a local k3d cluster was stood up, the operator was run against it, and a Deployment was scaled directly via `kubectl` (bypassing any Terraform-managed path) to simulate an unauthorized change. The operator correctly detected the drift and wrote a single, accurate event to `drift_events`.
- **Reconciler correctness fixes**: initial testing surfaced two real bugs in the drift controller â€” duplicate drift events caused by status-only reconciles, and a resource-version conflict causing retry-induced duplicate writes â€” both were root-caused and fixed (via a `GenerationChangedPredicate` filter and `retry.RetryOnConflict`, respectively), and re-verified to produce exactly one clean event per real spec change.
- **CORS and networking**: the Flutter web client, Go API, Postgres (in Docker), and the k3d cluster were all connected and validated communicating correctly across their actual local network boundaries â€” not just tested in isolation.
- **Automated test coverage**: the risk-scoring threshold classifier, the policy engine, the Terraform tag extraction, and the drift reconciler's diff/snapshot logic (across all three watched resource kinds) all have unit tests covering boundary conditions, defensive edge cases (nil pointers, empty/missing fields), and regression tests â€” e.g. asserting the deterministic threshold always overrides a mismatched LLM-claimed risk label, and that a policy rule can raise but never lower the final risk level. All 33 Python tests pass (verified in CI-equivalent conditions); the dashboard's TypeScript compiles clean (`tsc --noEmit`) and lints clean under the project's ESLint config.
- **Go/Flutter verification caveat**: the Go test suite and the Flutter app were verified by the same methodology as v1 (direct local runs) but could not be re-run inside the sandboxed environment used to build v2, since it has no Go or Flutter toolchain installed â€” verify `go test ./... -v` and `flutter analyze` locally before treating v2 as fully green.

## v1 Phase Build History

| Phase | Name | What Was Built |
|-------|------|----------------|
| 0 | Project Init & Config | Repo scaffolding, `.gitignore`, Docker Compose, k3d config, Go module init, Python requirements, Flutter `pubspec.yaml`, all `.env.example` files |
| 1 | Core Structure | Go config loader, Postgres connection pool, DB models, first migration, minimal Gin server with `/health`, FastAPI app bootstrap with `/health`, Flutter app shell |
| 2 | AI Risk Scoring Engine | Terraform plan parser, LLM client with a risk-analysis system prompt, deterministic threshold-based risk engine, `/scan` endpoint |
| 3 | K8s Drift Controller | Deployment snapshot/diff logic, controller-runtime reconciler, operator entrypoint watching live cluster state for drift |
| 4 | Approval Flow (API + Mobile) | Approval/device-token migrations, FCM push notifier, full REST handler set (scans, decisions, devices, drift events), Flutter models, API service, approval card widget, home screen, and detail screen |
| 5 | Polish & Finalize | Multi-stage Go Dockerfile (server + operator), Python Dockerfile, complete project README |
| 6 | Hardening & Test Coverage | Fixed two reconciler bugs found during live testing (duplicate reconciles, resource-version conflicts); added unit tests for the risk-scoring threshold engine and the drift reconciler's diff/snapshot logic |

## v2 Phase Build History

> v2 restarts its own phase count from the scaffolding step, since it's an additive build on top of the already-complete v1 above rather than a continuation of the same numbering.

| Phase | Name | What Was Built |
|-------|------|----------------|
| 6 | v2 Scaffolding & Config | `dashboard/` Next.js app init, empty stub files for auth/new reconcilers/users migration, `.env.example`/`.env.local.example` additions for the new config vars |
| 7 | Authentication Layer | JWT generate/validate (`internal/auth`), `users` migration, `AuthRequired` middleware, `/api/v1/auth/login`, `Handler`/`NewRouter` threaded with JWT config, `POST /scans/:id/decision` and `POST /devices` now require a Bearer token |
| 8 | Expanded Drift Detection | `StatefulSetDriftReconciler` and `ConfigMapDriftReconciler` (same snapshot/diff/annotation pattern as the Deployment reconciler), both registered with the operator's manager, unit tests for each including nil/empty-field edge cases |
| 9 | Policy-as-Code Layer | `policy.yaml` + `PolicyEngine` (loaded once at startup, applied as a floor over the LLM/threshold result), `tags` threaded through `ResourceChange`/the Terraform parser so tag-based rules (e.g. `env=prod` deletes) are possible, full unit + integration test coverage |
| 10 | Dashboard Read Views & API Wiring | Next.js pages for scan history (with risk/status filters), drift events, and the approval audit trail; `apiClient.ts`; `GET /api/v1/scans` (filtered) and `GET /api/v1/audit` added to the Go API, both auth-protected |
| 11 | Auth Wiring Across Clients | Flutter login screen, on-device JWT storage (`shared_preferences`), `AuthGate` startup check, `Authorization` header on the two protected mobile calls, redirect-to-login on a 401 from either client |
| 12 | Polish & Finalize | Dashboard `Dockerfile` + `docker-compose.yml` entry, `seed-user` CLI (there was no way to create a login before this), corrected `next.config.ts`/`package.json` drift from the scaffolding step, full README update |

## Roadmap

- [ ] Cost-impact analytics on flagged changes (originally scoped as a separate .NET billing service â€” per the v2 constraints, this belongs in the Python service or the dashboard if built, not a new backend)
- [ ] Slack/Teams notification channel as an alternative to mobile push
- [ ] Per-user device registration and role-based permissions (currently any authenticated user can approve/reject â€” there's no reviewer/admin distinction yet)
- [ ] A real signup/user-management flow (currently a single `seed-user` CLI command, intentionally minimal for a portfolio project)
- [ ] Support for additional Kubernetes resource kinds beyond Deployment/StatefulSet/ConfigMap (Services, Ingresses, HPAs)

## Why This Exists

Existing infrastructure governance tools â€” Spacelift, env0, Firefly â€” are built dashboard-first, which assumes the person reviewing a change is at their laptop when it matters. In practice, the person who needs to approve a risky Terraform change is often paged while away from a screen. TerraSentry's wedge is closing that specific gap: risky changes reach the on-call engineer wherever they are, and a decision takes two taps instead of a context switch into a web UI â€” while still giving the rest of the team a proper dashboard for everything that isn't time-critical.

## Contributing

This is currently a solo portfolio project, but issues and pull requests are welcome if you'd like to extend it â€” particularly around role-based permissions, additional Kubernetes resource support, or alternative LLM providers.

## License

MIT
