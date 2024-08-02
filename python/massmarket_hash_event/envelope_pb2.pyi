# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import authentication_pb2 as _authentication_pb2
from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from massmarket_hash_event import shop_requests_pb2 as _shop_requests_pb2
from massmarket_hash_event import subscription_pb2 as _subscription_pb2
from massmarket_hash_event import transport_pb2 as _transport_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Envelope(_message.Message):
    __slots__ = ["request_id", "event_write_request", "event_write_response", "subscription_request", "subscription_response", "subscription_cancel", "subscription_push", "subscription_push_response", "sync_status_request", "sync_status_response", "ping_request", "ping_response", "get_blob_upload_url_request", "get_blob_upload_url_response", "auth_request", "auth_response", "challenge_solution_request", "challenge_solution_response"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_WRITE_REQUEST_FIELD_NUMBER: _ClassVar[int]
    EVENT_WRITE_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_REQUEST_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_CANCEL_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_PUSH_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_PUSH_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    SYNC_STATUS_REQUEST_FIELD_NUMBER: _ClassVar[int]
    SYNC_STATUS_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    PING_REQUEST_FIELD_NUMBER: _ClassVar[int]
    PING_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    GET_BLOB_UPLOAD_URL_REQUEST_FIELD_NUMBER: _ClassVar[int]
    GET_BLOB_UPLOAD_URL_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    AUTH_REQUEST_FIELD_NUMBER: _ClassVar[int]
    AUTH_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    CHALLENGE_SOLUTION_REQUEST_FIELD_NUMBER: _ClassVar[int]
    CHALLENGE_SOLUTION_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    request_id: _base_types_pb2.RequestId
    event_write_request: _transport_pb2.EventWriteRequest
    event_write_response: _transport_pb2.EventWriteResponse
    subscription_request: _subscription_pb2.SubscriptionRequest
    subscription_response: _subscription_pb2.SubscriptionResponse
    subscription_cancel: _subscription_pb2.SubscriptionCancel
    subscription_push: _subscription_pb2.SubscriptionPush
    subscription_push_response: _subscription_pb2.SubscriptionPushResponse
    sync_status_request: _transport_pb2.SyncStatusRequest
    sync_status_response: _transport_pb2.SyncStatusResponse
    ping_request: _transport_pb2.PingRequest
    ping_response: _transport_pb2.PingResponse
    get_blob_upload_url_request: _shop_requests_pb2.GetBlobUploadURLRequest
    get_blob_upload_url_response: _shop_requests_pb2.GetBlobUploadURLResponse
    auth_request: _authentication_pb2.AuthenticateRequest
    auth_response: _authentication_pb2.AuthenticateResponse
    challenge_solution_request: _authentication_pb2.ChallengeSolvedRequest
    challenge_solution_response: _authentication_pb2.ChallengeSolvedResponse
    def __init__(self, request_id: _Optional[_Union[_base_types_pb2.RequestId, _Mapping]] = ..., event_write_request: _Optional[_Union[_transport_pb2.EventWriteRequest, _Mapping]] = ..., event_write_response: _Optional[_Union[_transport_pb2.EventWriteResponse, _Mapping]] = ..., subscription_request: _Optional[_Union[_subscription_pb2.SubscriptionRequest, _Mapping]] = ..., subscription_response: _Optional[_Union[_subscription_pb2.SubscriptionResponse, _Mapping]] = ..., subscription_cancel: _Optional[_Union[_subscription_pb2.SubscriptionCancel, _Mapping]] = ..., subscription_push: _Optional[_Union[_subscription_pb2.SubscriptionPush, _Mapping]] = ..., subscription_push_response: _Optional[_Union[_subscription_pb2.SubscriptionPushResponse, _Mapping]] = ..., sync_status_request: _Optional[_Union[_transport_pb2.SyncStatusRequest, _Mapping]] = ..., sync_status_response: _Optional[_Union[_transport_pb2.SyncStatusResponse, _Mapping]] = ..., ping_request: _Optional[_Union[_transport_pb2.PingRequest, _Mapping]] = ..., ping_response: _Optional[_Union[_transport_pb2.PingResponse, _Mapping]] = ..., get_blob_upload_url_request: _Optional[_Union[_shop_requests_pb2.GetBlobUploadURLRequest, _Mapping]] = ..., get_blob_upload_url_response: _Optional[_Union[_shop_requests_pb2.GetBlobUploadURLResponse, _Mapping]] = ..., auth_request: _Optional[_Union[_authentication_pb2.AuthenticateRequest, _Mapping]] = ..., auth_response: _Optional[_Union[_authentication_pb2.AuthenticateResponse, _Mapping]] = ..., challenge_solution_request: _Optional[_Union[_authentication_pb2.ChallengeSolvedRequest, _Mapping]] = ..., challenge_solution_response: _Optional[_Union[_authentication_pb2.ChallengeSolvedResponse, _Mapping]] = ...) -> None: ...
