import json

import pytest

import anyvali as v


def test_sensitive_data_round_trip_and_validation():
    schema = v.object_({
        "name": v.string(),
        "secret": v.string().min_length(3).describe("secret", sensitive=True),
        "profile": v.object_({"token": v.string()}).describe("profile", sensitive=True),
        "aliases": v.array(v.string().describe("alias", sensitive=True)),
        "note": v.nullable(v.string().describe("note", sensitive=True)),
    })
    plain = {
        "name": "Ada",
        "secret": "abc",
        "profile": {"token": "xyz"},
        "aliases": ["one", "two"],
        "note": None,
    }

    seen = []

    def enc(path, value):
        seen.append(path)
        return "encrypted:" + json.dumps(value, separators=(",", ":"))

    encrypted = v.encrypt(schema, plain, enc)
    assert encrypted == {
        "name": "Ada",
        "secret": 'encrypted:"abc"',
        "profile": 'encrypted:{"token":"xyz"}',
        "aliases": ['encrypted:"one"', 'encrypted:"two"'],
        "note": None,
    }
    assert seen == [["secret"], ["profile"], ["aliases", 0], ["aliases", 1]]
    assert v.safe_parse_encrypted(schema, encrypted).success
    assert v.decrypt(schema, encrypted, lambda _path, value: json.loads(value[10:])) == plain

    assert not v.safe_parse_encrypted(schema, {**plain, "secret": "abc"}).success
    with pytest.raises(v.ValidationError):
        v.encrypt(schema, plain, lambda _path, _value: "broken")
