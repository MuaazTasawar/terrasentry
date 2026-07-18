# TerraSentry

> AI-powered infrastructure governance with a mobile-first approval loop — catch risky Terraform changes and live cluster drift before they bite you.

## Overview

TerraSentry is an infrastructure governance copilot built for platform and DevOps teams who need risk visibility without living in a dashboard. It scans Terraform plans pre-apply for cost, security, and drift risk using an LLM, runs a lightweight Kubernetes controller that continuously watches live cluster state for configuration drift, and — instead of routing everything through a web UI — pushes risky changes straight to an on-call engineer's phone as a one-tap approve/reject decision. It exists because the existing category (Spacelift, env0, Firefly) assumes the reviewer is sitting at a laptop; TerraSentry assumes they're not, and closes that gap with a genuinely mobile-first approval loop.

## Features

- **Pre-apply Terraform risk scoring** — submit a `terraform show -json` plan and get back an LLM-generated risk score (0–100), a risk tier (low/medium/high), human-readable reasoning, and a list of specifically flagged resources.
- **Deterministic threshold safety net** — the LLM's risk score is re-checked against configurable numeric thresholds server-side, so the final low/medium/high classification never depends on the model's own labeling alone.
- **Live Kubernetes drift detection** — a custom controller (built with controller-runtime) watches Deployments, snapshots their key spec fields (replicas, image, resource limits, labels), and detects when live state diverges from the last known-good baseline — catching unauthorized `kubectl edit`/`kubectl scale` changes that bypass Terraform entirely.
- **Mobile-first approval flow** — medium/high risk scans trigger a push notification to every registered on-call device; the Flutter app shows a home feed of pending approvals and a detail screen with full reasoning, plan summary, and one-tap Approve/Reject actions.
- **Full audit trail** — every approval or rejection is recorded with who decided, what they decided, and when, stored independently of the scan record itself.
- **Fail-safe LLM parsing** — if the LLM ever returns a response that can't be parsed as valid JSON, the system defaults to flagging the change as high-risk for manual review rather than silently letting it through.
- **Zero-cost local development** — the entire stack (API, risk-scoring service, Postgres, Kubernetes cluster) runs locally via Docker Compose and k3d, with no managed cloud services required to build, test, or demo the project.

## Tech Stack

| Layer | Technology |
|-------|------------|
| API & K8s Controller | Go, Gin, pgx, controller-runtime, client-go |
| AI Risk Scoring | Python, FastAPI, Anthropic API, python-hcl2, httpx |
| Mobile Approvals | Flutter, Dart, http, provider |
| Database | PostgreSQL 16 |
| Local Kubernetes | k3d (k3s in Docker) |
| Containerization | Docker, Docker Compose |
| Push Notifications | Firebase Cloud Messaging (FCM) |

## Project Structure

```
terrasentry/
├── api/
│   ├── cmd/
│   │   ├── server/
│   │   │   └── main.go
│   │   └── operator/
│   │       └── main.go
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handlers.go
│   │   │   ├── router.go
│   │   │   └── middleware.go
│   │   ├── db/
│   │   │   ├── db.go
│   │   │   └── models.go
│   │   ├── controller/
│   │   │   ├── drift_controller.go
│   │   │   └── reconciler.go
│   │   ├── notify/
│   │   │   └── push.go
│   │   └── config/
│   │       └── config.go
│   ├── migrations/
│   │   ├── 001_init.sql
│   │   └── 002_approvals.sql
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   └── .env.example
├── risk-scoring/
│   ├── app/
│   │   ├── main.py
│   │   ├── routers/
│   │   │   └── scan.py
│   │   ├── services/
│   │   │   ├── llm_client.py
│   │   │   ├── terraform_parser.py
│   │   │   └── risk_engine.py
│   │   ├── models/
│   │   │   └── schemas.py
│   │   └── config.py
│   ├── requirements.txt
│   ├── Dockerfile
│   └── .env.example
├── mobile/
│   ├── lib/
│   │   ├── main.dart
│   │   ├── models/
│   │   │   └── approval_request.dart
│   │   ├── screens/
│   │   │   ├── home_screen.dart
│   │   │   └── approval_detail_screen.dart
│   │   ├── services/
│   │   │   └── api_service.dart
│   │   └── widgets/
│   │       └── approval_card.dart
│   ├── pubspec.yaml
│   └── .env.example
├── infra/
│   ├── k3d-config.yaml
│   └── docker-compose.yml
├── .gitignore
└── README.md
```

## Architecture

```
┌─────────────┐     scan plan      ┌────────────────────┐
│  Terraform   │ ─────────────────▶│  Risk Scoring (Py)  │
│  CI Pipeline │                    │  FastAPI + LLM      │
└─────────────┘                    └──────────┬───────────┘
                                               │ risk score
                                               ▼
┌─────────────┐    drift events   ┌────────────────────┐    push notify   ┌──────────────┐
│  K8s Operator│ ─────────────────▶│   Go API + Postgres │ ────────────────▶│ Flutter App  │
│  (Go)        │                   │                      │◀──approve/reject─│ (on-call)    │
└─────────────┘                   └────────────────────┘                   └──────────────┘
```

**How a change flows through the system:**
1. A CI pipeline runs `terraform plan`, converts it to JSON (`terraform show -json`), and POSTs it to the risk-scoring service.
2. The risk-scoring service parses the plan into a compact resource-change summary, sends it to the LLM with a risk-analysis system prompt, and returns a structured score.
3. The Go API stores the scored scan in Postgres. If the risk is medium or high, it pushes an FCM notification to every registered on-call device.
4. Independently, the K8s operator continuously reconciles Deployments in the cluster, comparing live spec against a stored baseline snapshot, and writes any detected drift to the same database.
5. The Flutter app polls the API for pending approvals, shows them in a risk-colored list, and lets the on-call engineer approve or reject with two taps.
6. Every decision is written to an audit table, independent of the scan record, for traceability.

## Getting Started

### Prerequisites

| Tool | Minimum Version | Notes |
|---|---|---|
| Go | 1.22+ | Builds the API server and K8s operator |
| Python | 3.12+ | Runs the FastAPI risk-scoring service |
| Flutter | 3.3+ (Dart 3.3+) | Builds and runs the mobile approval app |
| Docker & Docker Compose | Recent stable | Runs Postgres and (optionally) containerized services |
| k3d | v5+ | Spins up a local Kubernetes cluster for drift detection (optional but needed to demo the operator) |
| kubectl | Any recent version | Interacts with the local k3d cluster |
| An Anthropic API key | — | Powers the LLM risk-scoring calls; free-tier alternatives (Groq, Gemini Flash, local Ollama) can be swapped in via `risk-scoring/app/services/llm_client.py` |

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
cp mobile/.env.example mobile/.env
```
Then edit `risk-scoring/.env` and add a real `LLM_API_KEY`. All other defaults work out of the box for local development.

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

**4. Install Flutter dependencies**
```bash
cd mobile
flutter pub get
cd ..
```

### Running the App

**Start Postgres (and optionally the containerized API/risk-scoring services):**
```bash
cd infra
docker compose up --build
```

**Run database migrations** (only needed once, or after pulling new migration files):
```bash
psql "postgres://terrasentry:terrasentry@localhost:5432/terrasentry" -f api/migrations/001_init.sql
psql "postgres://terrasentry:terrasentry@localhost:5432/terrasentry" -f api/migrations/002_approvals.sql
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

**Spin up a local Kubernetes cluster and run the drift operator:**
```bash
k3d cluster create --config infra/k3d-config.yaml
cd api
go run cmd/operator/main.go
```
Then deploy a sample Deployment and mutate it directly (bypassing Terraform) to trigger drift detection:
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
> Note: on an Android emulator, the API base URL in `mobile/lib/services/api_service.dart` should point to `http://10.0.2.2:8080` instead of `localhost`, since the emulator can't resolve the host machine's localhost directly. On a physical device on the same network, use your machine's LAN IP.

## Environment Variables

### `api/.env`

| Variable | Description | Where to get it |
|---|---|---|
| `PORT` | Port the Go API listens on | Default `8080`, change if needed |
| `ENV` | Environment name (`development`/`production`) | Set manually |
| `DATABASE_URL` | Postgres connection string | Matches `docker-compose.yml` credentials by default |
| `RISK_SCORING_URL` | Base URL of the Python risk-scoring service | Default `http://localhost:8000` |
| `JWT_SECRET` | Secret used for future auth/token signing | Generate any long random string |
| `FCM_SERVER_KEY` | Firebase Cloud Messaging server key for push notifications | Firebase Console → Project Settings → Cloud Messaging |
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

### `mobile/.env`

| Variable | Description | Where to get it |
|---|---|---|
| `API_BASE_URL` | Base URL the Flutter app calls for approvals | `http://localhost:8080` locally, or your deployed API URL |

## API Reference

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | API server health check |
| `POST` | `/api/v1/scans` | Store a scored plan; triggers a push notification if risk is medium/high |
| `GET` | `/api/v1/scans/pending` | List all scans currently awaiting a decision |
| `POST` | `/api/v1/scans/:id/decision` | Approve or reject a scan (`{"decision": "approved" \| "rejected"}`) |
| `POST` | `/api/v1/devices` | Register a device token for push notifications |
| `GET` | `/api/v1/drift-events` | List the 50 most recent Kubernetes drift events |
| `GET` | `/health` (risk-scoring service, port 8000) | Risk-scoring service health check |
| `POST` | `/scan` (risk-scoring service, port 8000) | Score a raw `terraform show -json` plan and return a risk assessment |

## Verified Working

This isn't just scaffolded code — the full loop has been manually tested end-to-end on a local environment:

- **Risk-scored approval flow**: a test scan was inserted with a high-risk score, correctly rendered in the Flutter app with the right risk-color coding, opened into the detail view with full reasoning and plan summary, and approved via the mobile UI. The decision was confirmed written back to the `approval_audit` table with the correct scan reference, decision, and timestamp.
- **Live Kubernetes drift detection**: a local k3d cluster was stood up, the operator was run against it, and a Deployment was scaled directly via `kubectl` (bypassing any Terraform-managed path) to simulate an unauthorized change. The operator correctly detected the drift and wrote a single, accurate event to `drift_events`.
- **Reconciler correctness fixes**: initial testing surfaced two real bugs in the drift controller — duplicate drift events caused by status-only reconciles, and a resource-version conflict causing retry-induced duplicate writes — both were root-caused and fixed (via a `GenerationChangedPredicate` filter and `retry.RetryOnConflict`, respectively), and re-verified to produce exactly one clean event per real spec change.
- **CORS and networking**: the Flutter web client, Go API, Postgres (in Docker), and the k3d cluster were all connected and validated communicating correctly across their actual local network boundaries — not just tested in isolation.

## Phase Build History

| Phase | Name | What Was Built |
|-------|------|----------------|
| 0 | Project Init & Config | Repo scaffolding, `.gitignore`, Docker Compose, k3d config, Go module init, Python requirements, Flutter `pubspec.yaml`, all `.env.example` files |
| 1 | Core Structure | Go config loader, Postgres connection pool, DB models, first migration, minimal Gin server with `/health`, FastAPI app bootstrap with `/health`, Flutter app shell |
| 2 | AI Risk Scoring Engine | Terraform plan parser, LLM client with a risk-analysis system prompt, deterministic threshold-based risk engine, `/scan` endpoint |
| 3 | K8s Drift Controller | Deployment snapshot/diff logic, controller-runtime reconciler, operator entrypoint watching live cluster state for drift |
| 4 | Approval Flow (API + Mobile) | Approval/device-token migrations, FCM push notifier, full REST handler set (scans, decisions, devices, drift events), Flutter models, API service, approval card widget, home screen, and detail screen |
| 5 | Polish & Finalize | Multi-stage Go Dockerfile (server + operator), Python Dockerfile, complete project README |

## Roadmap

- [ ] Next.js dashboard for team-wide visibility into scan and drift history
- [ ] Cost-impact analytics on flagged changes (originally scoped as a separate .NET billing service, folded into a future phase)
- [ ] Support for additional Kubernetes resource kinds beyond Deployments (StatefulSets, ConfigMaps, Services)
- [ ] Slack/Teams notification channel as an alternative to mobile push
- [ ] Authentication and per-user device registration (currently single shared on-call pool)
- [ ] Configurable risk-scoring policy rules layered on top of the LLM's judgment

## Why This Exists

Existing infrastructure governance tools — Spacelift, env0, Firefly — are built dashboard-first, which assumes the person reviewing a change is at their laptop when it matters. In practice, the person who needs to approve a risky Terraform change is often paged while away from a screen. TerraSentry's wedge is closing that specific gap: risky changes reach the on-call engineer wherever they are, and a decision takes two taps instead of a context switch into a web UI.

## Contributing

This is currently a solo portfolio project, but issues and pull requests are welcome if you'd like to extend it — particularly around additional Kubernetes resource support, alternative LLM providers, or the planned Next.js dashboard.

## License

MIT