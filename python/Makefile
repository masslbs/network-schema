# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

OUT=massmarket_hash_event/
PROTO_FILES = transport.proto authentication.proto store_requests.proto store_events.proto error.proto
PROTO_PY_FILES = $(PROTO_FILES:.proto=_pb2.py)
PROTO_PYI_FILES = $(PROTO_FILES:.proto=_pb2.pyi)
PROTO_PY_OUT = $(addprefix ${OUT}, $(PROTO_PY_FILES))
PROTO_PYI_OUT = $(addprefix ${OUT}, $(PROTO_PYI_FILES))

all: ${PROTO_PY_FILES} ${PROTO_PYI_FILES} test

test:
	pytest

build:
	${PYTHON} -m build -n

publish:
	${PYTHON} -m twine upload dist/massmarket_hash_event-${VERSION}.tar.gz

%_pb2.py %_pb2.pyi: ../%.proto
	protoc -I=.. --python_out=${OUT} $<
	protoc -I=.. --pyi_out=${OUT} $<
	sed -i 's#import error_pb2 as error__pb2#from massmarket_hash_event import error_pb2 as error__pb2#'  ${OUT}$*_pb2.py
	sed -i 's#import error_pb2 as _error_pb2#from massmarket_hash_event import error_pb2 as _error_pb2#'  ${OUT}$*_pb2.pyi
