# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import authentication_pb2 as _authentication_pb2
from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from massmarket_hash_event import error_pb2 as _error_pb2
from massmarket_hash_event import shop_requests_pb2 as _shop_requests_pb2
from massmarket_hash_event import subscription_pb2 as _subscription_pb2
from massmarket_hash_event import transport_pb2 as _transport_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import (
    ClassVar as _ClassVar,
    Mapping as _Mapping,
    Optional as _Optional,
    Union as _Union,
)

DESCRIPTOR: _descriptor.FileDescriptor

class Envelope(_message.Message):
    __slots__ = [
        "request_id",
        "response",
        "event_write_request",
        "subscription_request",
        "subscription_cancel_request",
        "subscription_push_request",
        "sync_status_request",
        "ping_request",
        "get_blob_upload_url_request",
        "auth_request",
        "challenge_solution_request",
    ]

    class GenericResponse(_message.Message):
        __slots__ = ["error", "payload"]
        ERROR_FIELD_NUMBER: _ClassVar[int]
        PAYLOAD_FIELD_NUMBER: _ClassVar[int]
        error: _error_pb2.Error
        payload: bytes
        def __init__(
            self,
            error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ...,
            payload: _Optional[bytes] = ...,
        ) -> None: ...

    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    EVENT_WRITE_REQUEST_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_REQUEST_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_CANCEL_REQUEST_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_PUSH_REQUEST_FIELD_NUMBER: _ClassVar[int]
    SYNC_STATUS_REQUEST_FIELD_NUMBER: _ClassVar[int]
    PING_REQUEST_FIELD_NUMBER: _ClassVar[int]
    GET_BLOB_UPLOAD_URL_REQUEST_FIELD_NUMBER: _ClassVar[int]
    AUTH_REQUEST_FIELD_NUMBER: _ClassVar[int]
    CHALLENGE_SOLUTION_REQUEST_FIELD_NUMBER: _ClassVar[int]
    request_id: _base_types_pb2.RequestId
    response: Envelope.GenericResponse
    event_write_request: _transport_pb2.EventWriteRequest
    subscription_request: _subscription_pb2.SubscriptionRequest
    subscription_cancel_request: _subscription_pb2.SubscriptionCancelRequest
    subscription_push_request: _subscription_pb2.SubscriptionPushRequest
    sync_status_request: _transport_pb2.SyncStatusRequest
    ping_request: _transport_pb2.PingRequest
    get_blob_upload_url_request: _shop_requests_pb2.GetBlobUploadURLRequest
    auth_request: _authentication_pb2.AuthenticateRequest
    challenge_solution_request: _authentication_pb2.ChallengeSolvedRequest
    def __init__(
        self,
        request_id: _Optional[_Union[_base_types_pb2.RequestId, _Mapping]] = ...,
        response: _Optional[_Union[Envelope.GenericResponse, _Mapping]] = ...,
        event_write_request: _Optional[
            _Union[_transport_pb2.EventWriteRequest, _Mapping]
        ] = ...,
        subscription_request: _Optional[
            _Union[_subscription_pb2.SubscriptionRequest, _Mapping]
        ] = ...,
        subscription_cancel_request: _Optional[
            _Union[_subscription_pb2.SubscriptionCancelRequest, _Mapping]
        ] = ...,
        subscription_push_request: _Optional[
            _Union[_subscription_pb2.SubscriptionPushRequest, _Mapping]
        ] = ...,
        sync_status_request: _Optional[
            _Union[_transport_pb2.SyncStatusRequest, _Mapping]
        ] = ...,
        ping_request: _Optional[_Union[_transport_pb2.PingRequest, _Mapping]] = ...,
        get_blob_upload_url_request: _Optional[
            _Union[_shop_requests_pb2.GetBlobUploadURLRequest, _Mapping]
        ] = ...,
        auth_request: _Optional[
            _Union[_authentication_pb2.AuthenticateRequest, _Mapping]
        ] = ...,
        challenge_solution_request: _Optional[
            _Union[_authentication_pb2.ChallengeSolvedRequest, _Mapping]
        ] = ...,
    ) -> None: ...
