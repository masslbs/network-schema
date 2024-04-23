# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

.phony: all lint reuse

all: testVectors.json typedData.json

transport_pb2.py: transport.proto
	protoc --python_out=. transport.proto
	protoc --pyi_out=. transport.proto

store_events_pb2.py: store_events.proto
	protoc --python_out=. store_events.proto
	protoc --pyi_out=. store_events.proto

typedData.json: store_events_pb2.py generate_typedData_json.py
	$(PYTHON) ./generate_typedData_json.py

testVectors.json: store_events_pb2.py generate_testVectors_json.py typedData.json
	PYTHONPATH=$$PYTHONPATH:$(PWD)/python $(PYTHON) ./generate_testVectors_json.py

lint:
	protolint lint *.proto
	$(PYTHON) ./check.py
	reuse lint

LIC := MIT
CPY := "Mass Labs"

reuse:
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights Makefile CHANGELOG.md README.md schema.proto flake.nix .gitignore .github/workflows/test.yml *.py *.pyi python/tests/*.py python/massmarket_hash_event/*.py python/pyproject.toml
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights --force-dot-license VERSION *.txt flake.lock *.json

