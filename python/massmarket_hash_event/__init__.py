# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

__all__ = ['hash_event', 'Hasher',
           'transport_pb2', 'authentication_pb2', 'shop_requests_pb2', 'error_pb2',
           'shop_events_pb2']

import json
import binascii
from pprint import pprint
from importlib.resources import files

from eth_account.messages import encode_structured_data

from massmarket_hash_event import shop_events_pb2

event_typedData_spec = json.loads(files("massmarket_hash_event").joinpath("typedData.json").read_text())

class Hasher:
    def __init__(self, chain_id: int, shopRegAddress: str):
        self.chain_id = chain_id
        if not shopRegAddress.startswith("0x"):
            raise Exception(f"Invalid contract address: {shopRegAddress}")
        data = binascii.unhexlify(shopRegAddress[2:])
        if len(data) != 20:
            raise Exception(f"Invalid contract address: {shopRegAddress}")
        self.shopRegAddress = shopRegAddress
    def hash_event(self, evt: shop_events_pb2.ShopEvent):
        return hash_event(evt, self.chain_id, self.shopRegAddress)

def hash_event(evt: shop_events_pb2.ShopEvent, chain_id, shopRegAddress):
    event_type = getattr(evt, evt.WhichOneof('union'))
    event_name = event_type.DESCRIPTOR.name
    event_td_spec = event_typedData_spec[event_name].copy()
    #print(f"hash_event({event_name})")

    types = {
        "EIP712Domain": [
            {"name": "name", "type": "string"},
            {"name": "version", "type": "string"},
            {"name": "chainId", "type": "uint256"},
            {"name": "verifyingContract", "type": "address"},
        ],
    }

    # TODO: this should be generalized / annotate which fields are optional
    # {name, type} information should be taken from typedData.json


    # step 1: remove fields from event_td_spec that are not set in event_td_data
    if event_name == "UpdateShopManifest":
        event_td_spec = event_td_spec[:1] # only event_id

        if event_type.HasField("domain"):
            event_td_spec.append({"name": "domain", "type": "string"})

        if event_type.HasField("published_tag_id"):
            event_td_spec.append({"name": "published_tag_id", "type": "bytes32"})

        if event_type.HasField("add_erc20_addr"):
            event_td_spec.append({"name": "add_erc20_addr", "type": "address"})

        if event_type.HasField("remove_erc20_addr"):
            event_td_spec.append({"name": "remove_erc20_addr", "type": "address"})

        if event_type.HasField("name"):
            event_td_spec.append({"name": "name", "type": "string"})

        if event_type.HasField("description"):
            event_td_spec.append({"name": "description", "type": "string"})

        if event_type.HasField("profile_picture_url"):
            event_td_spec.append({"name": "profile_picture_url", "type": "string"})

    elif event_name == "UpdateItem":
        event_td_spec = event_td_spec[:2] # event_id and item_id

        if event_type.HasField("price"):
            event_td_spec.append({"name": "price", "type": "string"})

        if event_type.HasField("metadata"):
            event_td_spec.append({"name": "metadata", "type": "bytes"})

    elif event_name == "UpdateTag":
        event_td_spec = event_td_spec[:2] # event_id and tag_id

        if event_type.HasField("rename"):
            event_td_spec.append({"name": "rename", "type": "string"})
        if event_type.HasField("delete"):
            event_td_spec.append({"name": "delete", "type": "bool"})
        if event_type.HasField("add_item_id"):
            event_td_spec.append({"name": "add_item_id", "type": "bytes32"})
        if event_type.HasField("remove_item_id"):
            event_td_spec.append({"name": "remove_item_id", "type": "bytes32"})


    elif event_name == "UpdateOrder":
        # UpdateOrder follows a nested message pattern and thus needs slightly differnt handling
        # we are not just pruning the unused values but add the used type to typed_data
        action = event_type.WhichOneof("action")

        # build map of action > message
        name2msgtd = {}
        msgs = [td for td in event_td_spec if "message" in td]
        for td in msgs:
            name2msgtd[td["name"]] = td["message"]

        # other fields are used as is
        event_td_spec = [td for td in event_td_spec if "message" not in td]

        if action not in name2msgtd:
         raise Exception(f"Unhandled action on UpdateOrder: {action}")

        event_td_spec.append({"name": action, "type": action})
        types[action] = name2msgtd[action].copy()


    # step 2: convert event to typedData
    event_td_data = event_to_typedData_dict(evt, event_td_spec)
    if event_name == "ChangeStock":
        # remove order_id if empty
        if len(event_td_data["order_id"]) == 0:
            event_td_spec.remove({"name": "order_id", "type": "bytes32"})
            event_td_spec.remove({"name": "tx_hash", "type": "bytes32"})
            del event_td_data["order_id"]
            del event_td_data["tx_hash"]
    elif event_name == "UpdateOrder":
        # we need to change the nested pb message into a subscriptable thing
        # for encode_structured_data to be happy
        action = event_type.WhichOneof("action")
        action_msg = getattr(event_type, action)
        action_obj = {}

        action_fields = action_msg.ListFields()
        for field in action_fields:
            name = field[0].name
            action_obj[name] = getattr(action_msg, name)

        # Bools that are false are not set on a pb message
        if "is_payment_endpoint" not in action_obj:
            action_obj["is_payment_endpoint"] = False

        # remove currency_addr if empty
        if action == "items_finalized" and "currency_addr" not in action_obj:
            types[action].remove({"name": "currency_addr", "type": "address"})

        # replace nested protobuf thing
        event_td_data[action] = action_obj

    types[event_name] = event_td_spec
    typed_data = {
        "types": types,
        "primaryType": event_name,
        "domain": {
            "name": "MassMarket",
            "version": "1",
            "chainId": chain_id,
            "verifyingContract": shopRegAddress,
        },
        "message": event_td_data
    }
    #pprint(typed_data)
    return encode_structured_data(typed_data)

def event_to_typedData_dict(evt: shop_events_pb2.ShopEvent, spec_for_event):
    which = evt.WhichOneof("union")
    if which is None:
        raise Exception("No event type set")
    unwrapped = getattr(evt, which)
    field_names = [field["name"] for field in spec_for_event]
    event_td_data = {}
    for field in field_names:
        event_td_data[field] = getattr(unwrapped, field)
    return event_td_data
