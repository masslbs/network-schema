# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

OUT=massmarket/

all: compile format test

compile:
	protoc -I=.. --pyi_out=${OUT} --python_out=${OUT} ../*.proto
	${PYTHON} ./tweak_imports.py

format:
	black massmarket
	black tests

test:
	pytest

build:
	${PYTHON} -m build -n

publish:
	${PYTHON} -m twine upload dist/massmarket-${VERSION}*
