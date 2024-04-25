from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import any_pb2 as _any_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Person(_message.Message):
    __slots__ = ["name", "id", "age"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    AGE_FIELD_NUMBER: _ClassVar[int]
    name: str
    id: int
    age: int
    def __init__(self, name: _Optional[str] = ..., id: _Optional[int] = ..., age: _Optional[int] = ...) -> None: ...

class Robot(_message.Message):
    __slots__ = ["name", "id", "features"]
    class Feature(_message.Message):
        __slots__ = ["name", "description"]
        NAME_FIELD_NUMBER: _ClassVar[int]
        DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
        name: str
        description: str
        def __init__(self, name: _Optional[str] = ..., description: _Optional[str] = ...) -> None: ...
    NAME_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    FEATURES_FIELD_NUMBER: _ClassVar[int]
    name: str
    id: int
    features: _containers.RepeatedCompositeFieldContainer[Robot.Feature]
    def __init__(self, name: _Optional[str] = ..., id: _Optional[int] = ..., features: _Optional[_Iterable[_Union[Robot.Feature, _Mapping]]] = ...) -> None: ...

class Task(_message.Message):
    __slots__ = ["title", "description", "due_date", "done_by"]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    DUE_DATE_FIELD_NUMBER: _ClassVar[int]
    DONE_BY_FIELD_NUMBER: _ClassVar[int]
    title: str
    description: str
    due_date: _timestamp_pb2.Timestamp
    done_by: _any_pb2.Any
    def __init__(self, title: _Optional[str] = ..., description: _Optional[str] = ..., due_date: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., done_by: _Optional[_Union[_any_pb2.Any, _Mapping]] = ...) -> None: ...
