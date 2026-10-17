.PHONY: up down reset smoke test
up:
	docker compose up --build -d
down:
	docker compose --profile hyperdx down
reset:
	docker compose --profile hyperdx down -v
smoke:
	python3 scripts/smoke.py
test:
	python3 -m unittest discover -s tests -v
