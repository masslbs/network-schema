# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import pytest

from massmarket_hash_event.cbor.ethereum_address import EthereumAddress
from massmarket_hash_event.cbor.chain_address import ChainAddress
from massmarket_hash_event.cbor.uint256 import Uint256

def test_cbor_ethereum_address():
    # Test creation from hex string
    addr = EthereumAddress("0x0102030405060708090a0b0c0d0e0f1011121314")
    assert str(addr) == "0x0102030405060708090a0b0c0d0e0f1011121314"

    # Test creation from bytes
    addr2 = EthereumAddress(bytes.fromhex("0102030405060708090a0b0c0d0e0f1011121314"))
    assert str(addr2) == "0x0102030405060708090a0b0c0d0e0f1011121314"
    assert addr == addr2

    # Test invalid length
    with pytest.raises(ValueError):
        EthereumAddress("0x0102")  # Too short
    
    # Test invalid format
    with pytest.raises(ValueError):
        EthereumAddress("0xXYZ")  # Invalid hex

def test_cbor_chain_address():
    # Test creation with string address
    chain_addr = ChainAddress(
        chain_id=1337,
        address="0x0102030405060708090a0b0c0d0e0f1011121314"
    )
    assert chain_addr.chain_id == 1337
    assert str(chain_addr.address) == "0x0102030405060708090a0b0c0d0e0f1011121314"

    # Test invalid chain_id
    with pytest.raises(ValueError):
        ChainAddress(chain_id=0, address="0x0102030405060708090a0b0c0d0e0f1011121314")

    # Test CBOR dict conversion
    cbor_dict = chain_addr.to_cbor_dict()
    assert cbor_dict["ChainID"] == 1337
    assert isinstance(cbor_dict["Address"], bytes)

    # Test round-trip through CBOR dict
    chain_addr2 = ChainAddress.from_cbor_dict(cbor_dict)
    assert chain_addr == chain_addr2

def test_cbor_uint256():
    # Test creation from int
    num = Uint256(12345)
    assert int(num) == 12345

    # Test creation from hex string
    num2 = Uint256("0xff")
    assert int(num2) == 255

    # Test max value
    max_value = (1 << 256) - 1
    num3 = Uint256(max_value)
    assert int(num3) == max_value

    # Test value too large
    with pytest.raises(ValueError):
        Uint256(1 << 256)

    # Test negative value
    with pytest.raises(ValueError):
        Uint256(-1)

    # Test bytes conversion
    num4 = Uint256(0x1234)
    bytes_val = num4.to_bytes()
    assert len(bytes_val) == 32  # Should be 32 bytes
    num5 = Uint256.from_bytes(bytes_val)
    assert num4 == num5 
    