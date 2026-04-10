.PHONY: run dev

run:
	go run .

dev:
	air -c .air.toml
