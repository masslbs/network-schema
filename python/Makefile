# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

OUT=massmarket_hash_event/

all: compile test

compile:
	protoc -I=.. --pyi_out=${OUT} --python_out=${OUT} ../*.proto
	${PYTHON} ./tweak_imports.py

test:
	pytest

build:
	${PYTHON} -m build -n

publish:
	${PYTHON} -m twine upload dist/massmarket_hash_event-${VERSION}*
