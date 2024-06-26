# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

.phony: all lint reuse

all: testVectors.json python-package reuse

python-package:
	make -C python

testVectors.json: generate_testVectors_json.py
	PYTHONPATH=$$PYTHONPATH:$(PWD)/python $(PYTHON) ./generate_testVectors_json.py

lint:
	protolint lint *.proto
	$(PYTHON) ./check.py
	reuse lint

LIC := MIT
CPY := "Mass Labs"

reuse:
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights Makefile CHANGELOG.md README.md *.proto flake.nix .gitignore .github/workflows/test.yml *.py python/tests/*.py python/massmarket_hash_event/*.pyi python/massmarket_hash_event/*.py python/pyproject.toml python/Makefile
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights --force-dot-license VERSION *.txt flake.lock *.json
