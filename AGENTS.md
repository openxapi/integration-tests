# Repository Guidelines

## Project Structure & Module Organization
- Tests live under `src/{exchange}/{language}/{protocol}/{module}/`.
- Keep changes isolated to a single module; do not cross‑edit other modules.
- Each module includes `README.md`, `env.example`, and often `API_COVERAGE.md`.
- Example: Go Spot WS tests at `src/binance/go/ws/spot/`; REST Spot at `src/binance/go/rest/spot/`; Python Spot WS at `src/binance/python/ws/spot/`.

## Build, Test, and Development Commands
- Go (run inside a module with its own `go.mod`):
  - `cd src/binance/go/ws/spot && go test -v ./...`
  - Single test: `go test -v -run TestFullIntegrationSuite ./...`
  - Maintenance: `go clean -testcache` (from anywhere), `go mod download` (inside module).
- Python (module venv):
  - `cd src/binance/python/ws/spot && python -m venv .venv && source .venv/bin/activate`
  - `pip install -r requirements.txt -r requirements-test.txt`
  - `pytest -q` or focused: `pytest -k test_streams`.
- Environment: copy and edit `env.example` to `env.local`, then `source env.local` before running tests.

## Coding Style & Naming Conventions
- Go: follow standard formatting. Run `gofmt -s -w .` (or `go fmt ./...`). Test files: `*_test.go`. Package‑local naming; use clear, descriptive test names.
- Python: PEP 8, 4‑space indentation. Tests named `test_*.py`; functions `test_*`. Prefer asyncio‑friendly patterns where applicable.
- Docs: keep module `README.md` and `API_COVERAGE.md` up to date when adding tests.

## Testing Guidelines
- Use provided fixtures/config in each module (`conftest.py` for Python, helpers in Go).
- Default to testnet when available; some modules (e.g., Options, Portfolio Margin) may require mainnet.
- Respect rate limits; avoid tight loops. Prefer existing helpers that throttle requests.
- Add focused, independent tests; avoid cross‑module dependencies.

## Commit & Pull Request Guidelines
- Commits: short, imperative subject; include module scope prefix when helpful (e.g., `binance/go/ws/spot: update streams tests`).
- PRs: include motivation, affected module path, setup steps (env vars), and sample command to reproduce (`go test -v ./...` or `pytest -q`). Link issues and attach logs/screenshots when relevant.

## Security & Configuration Tips
- Never commit secrets. Keep `.env`, `env.local`, keys, and tokens out of VCS.
- Document required vars in the module `env.example`; prefer testnet keys for CI/local runs.
