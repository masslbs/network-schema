# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

.phony: all lint reuse

all: testVectors.json python-package reuse

python-package:
	make -C python

testVectors.json: python-package generate_testVectors_json.py
	$(PYTHON) ./generate_testVectors_json.py

lint:
	$(PYTHON) ./check.py
	buf format -w
	git ls-files | grep -E '.(py|pyi)$$' | xargs black
	protolint lint *.proto
	reuse lint

LIC := MIT
CPY := "Mass Labs"

reuse:
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights Makefile CHANGELOG.md README.md *.proto flake.nix .gitignore .github/workflows/test.yml *.py python/*.py python/tests/*.py python/massmarket_hash_event/*.pyi python/massmarket_hash_event/*.py python/pyproject.toml python/Makefile go.mod go.sum go/*.go
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights --force-dot-license VERSION *.txt flake.lock
