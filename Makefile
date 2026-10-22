# sfg — Sunshine Facilities Group website and CMS.
#
# Configuration lives at /etc/sfg/sfg.env, not in the repo. Every target that
# touches the database or the object store sources it, so there is one place
# secrets come from and none of them are here.

BIN  := bin/sfg
ENV  := /etc/sfg/sfg.env
LOAD := set -a; . $(ENV); set +a;
ASSETS := docs/design/handoff/design/assets

.PHONY: help web icons build run migrate-up migrate-down migrate-status seed-assets seed-dev job \
        test test-shuffle e2e vet fmt contrast diagrams typecheck audit check deploy shots seed-sql clean

help: ## show this
	@grep -E '^[a-z-]+:.*?##' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-16s %s\n",$$1,$$2}'

web: ## build the frontend (embedded into the binary, so this runs FIRST)
	cd web && npm ci && npm run build

icons: ## regenerate the slim Phosphor icon module after adding an icon import
	cd web && node scripts/gen-icons.mjs

build: ## build the binary (run `make web` first if the UI changed)
	CGO_ENABLED=0 go build -trimpath -o $(BIN) ./cmd/api

run: build ## run the HTTP service in the foreground
	$(LOAD) $(BIN) serve

migrate-up: build ## apply pending migrations
	$(LOAD) $(BIN) migrate up

migrate-down: build ## roll back the latest migration (refused in production)
	$(LOAD) $(BIN) migrate down

migrate-status: build ## applied and pending migrations
	$(LOAD) $(BIN) migrate status

seed-sql: ## regenerate db/migrations/0002 from scripts/seed/content.py (only before 0002 is applied anywhere that matters)
	python3 -I scripts/seed/gen_seed.py

seed-assets: build ## upload and process the design's images into the object store (idempotent)
	$(LOAD) $(BIN) seed-assets $(ASSETS)

seed-dev: build ## create the development superadmin (refused in production)
	$(LOAD) $(BIN) seed-dev

job: build ## run a job: make job JOB=deliver-mail|certificate-reminders|purge-applications
	$(LOAD) $(BIN) job $(JOB)

vet: ## go vet
	go vet ./...

fmt: ## fail if anything is not gofmt'd
	@out=$$(gofmt -l cmd internal test db web 2>/dev/null); \
	if [ -n "$$out" ]; then echo "unformatted:"; echo "$$out"; exit 1; fi
	@echo "gofmt clean"

test: ## unit + integration + security suites (needs SFG_TEST_DATABASE_URL, a built web/dist)
	$(LOAD) go test ./... -count=1 -p 1

test-shuffle: ## the suite shuffled — order dependence is a defect
	$(LOAD) go test ./... -count=1 -p 1 -shuffle=on

e2e: ## Playwright against the running dev site (http://127.0.0.1:8097)
	cd web && SFG_E2E_PASSWORD=$$(sudo sed -n 's/^SFG_DEV_ADMIN_PASSWORD=//p' $(ENV)) npx playwright test

typecheck: ## TypeScript, no emit
	cd web && npx tsc -b --noEmit

audit: ## runtime npm advisories (build-tool advisories are listed in docs/12-security.md)
	cd web && npm audit --omit=dev --audit-level=low
	@command -v govulncheck >/dev/null && govulncheck ./... || echo "govulncheck not installed (see docs/RUN-WHEN-BACK.md)"

contrast: ## measure every colour pairing against design.md
	python3 -I scripts/contrast.py

diagrams: ## parse every mermaid diagram in docs/
	@node scripts/mermaid-check/check.mjs

check: fmt vet test contrast diagrams typecheck audit ## everything that gates a commit

shots: ## screenshots of every page, site vs approved design, 1440 + 360 (needs xvfb)
	cd web && LIBGL_ALWAYS_SOFTWARE=1 xvfb-run -a -s "-screen 0 1600x1000x24" node scripts/shots.mjs scripts/shots

deploy: web build migrate-up ## build both halves, migrate, restart, check
	sudo systemctl restart sfg
	@sleep 2
	@curl -fsS http://127.0.0.1:8098/readyz && echo
	@systemctl is-active sfg

clean:
	rm -rf bin web/dist/assets web/dist/index.html
