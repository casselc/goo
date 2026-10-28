.PHONY: up down reset smoke test test-reference image
up:
	docker compose up --build -d
down:
	docker compose --profile hyperdx down
reset:
	docker compose --profile hyperdx down -v
smoke:
	python3 scripts/smoke.py
test:
	go test ./...
	npm run test:visualizations

test-reference:
	python3 -m unittest discover -s tests -v

# Install the pinned tool as documented; this target only builds into the local
# Docker daemon and never publishes to a registry.
KO ?= ko
IMAGE_PLATFORM ?= linux/amd64
IMAGE_REFS ?= /tmp/goo-ko-image.refs
image:
	@case "$$($(KO) version)" in 0.19.1|v0.19.1) ;; *) printf '%s\n' 'Install ko v0.19.1; see docs/CONTAINER_IMAGES.md.' >&2; exit 1 ;; esac
	KO_DOCKER_REPO=ko.local $(KO) build --platform=$(IMAGE_PLATFORM) --tags=dev --tag-only --image-refs=$(IMAGE_REFS) ./cmd/goo
