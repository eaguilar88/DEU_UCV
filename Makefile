.PHONY: loadenv

loadenv:
	@set -a && source .env && set +a && env | grep -E '^HTTP_'
