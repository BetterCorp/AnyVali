# frozen_string_literal: true

require_relative "test_helper"
require "json"

class SensitiveTest < Minitest::Test
  def test_encrypted_round_trip_and_validation
    schema = AnyVali.object(
      properties: {
        "name" => AnyVali.string,
        "secret" => AnyVali.string.describe("secret", sensitive: true),
        "profile" => AnyVali.object(
          properties: { "token" => AnyVali.string },
          required: ["token"]
        ).describe("profile", sensitive: true),
        "aliases" => AnyVali.array(AnyVali.string.describe("alias", sensitive: true))
      },
      required: %w[name secret profile aliases]
    )
    plain = {
      "name" => "Ada", "secret" => "abc",
      "profile" => { "token" => "xyz" }, "aliases" => %w[one two]
    }

    encrypted = AnyVali.encrypt(schema, plain, ->(_path, value) { "encrypted:#{JSON.generate(value)}" })
    assert AnyVali.safe_parse_encrypted(schema, encrypted).success?
    assert_equal plain, AnyVali.decrypt(
      schema, encrypted, ->(_path, value) { JSON.parse(value.delete_prefix("encrypted:")) }
    )
    assert_raises(AnyVali::ValidationError) do
      AnyVali.encrypt(schema, plain, ->(*) { "broken" })
    end
  end
end
