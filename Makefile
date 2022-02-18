.PHONY: migrate, migrate-down

guard-%: ## Checks that env var is set else exits with non 0 mainly used in CI;
	@if [ -z '${${*}}' ]; then echo 'Environment variable $* not set' && exit 1; fi

migrate: guard-DATA_SOURCE_NAME
	@migrate -source file://migrations -database $(DATA_SOURCE_NAME) up

migrate-down: guard-DATA_SOURCE_NAME
	@migrate -source file://migrations -database $(DATA_SOURCE_NAME) down