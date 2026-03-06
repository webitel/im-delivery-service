export SUPPRESS_NO_CONFIG_WARNING=y
export NODE_ENV=development

.PHONY: gen-api gen-ts gen-docs

gen-api: gen-ts gen-docs

gen-ts:
	@echo "---- Generating TypeScript ----"
	./scripts/gen/gen-ts.sh

gen-docs:
	@echo "---- Generating Docs ----"
	./scripts/gen/gen-docs.sh