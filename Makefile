# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

.phony: all lint reuse

all: testVectors.json typedData.json

typedData.json: schema_pb2.py generate_typedData_json.py
	$(PYTHON) ./generate_typedData_json.py

testVectors.json: schema_pb2.py generate_testVectors_json.py
	PYTHONPATH=$$PYTHONPATH:$(PWD)/python $(PYTHON) ./generate_testVectors_json.py

schema_pb2.py: schema.proto
	protoc --python_out=. schema.proto
	protoc --pyi_out=. schema.proto

lint:
	protolint lint schema.proto
	$(PYTHON) ./check.py
	reuse lint

LIC := MIT
CPY := "Mass Labs"

reuse:
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights Makefile CHANGELOG.md README.md schema.proto flake.nix .gitignore .github/workflows/test.yml *.py *.pyi python/tests/*.py python/massmarket_hash_event/*.py python/pyproject.toml
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights --force-dot-license VERSION *.txt flake.lock *.json

