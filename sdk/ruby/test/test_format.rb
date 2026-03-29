# frozen_string_literal: true

require_relative "test_helper"

class TestFormat < Minitest::Test
  def test_email_valid
    s = AnyVali.string.format("email")
    assert_equal "user@example.com", s.parse("user@example.com")
  end

  def test_email_invalid_no_at
    s = AnyVali.string.format("email")
    result = s.safe_parse("not-an-email")
    assert result.failure?
    assert_equal "invalid_string", result.issues.first.code
    assert_equal "email", result.issues.first.expected
  end

  def test_email_invalid_no_dot
    s = AnyVali.string.format("email")
    result = s.safe_parse("user@localhost")
    assert result.failure?
  end

  def test_url_valid_https
    s = AnyVali.string.format("url")
    assert_equal "https://example.com", s.parse("https://example.com")
  end

  def test_url_valid_http
    s = AnyVali.string.format("url")
    assert_equal "http://example.com/path?q=1", s.parse("http://example.com/path?q=1")
  end

  def test_url_rejects_ftp
    s = AnyVali.string.format("url")
    result = s.safe_parse("ftp://files.example.com")
    assert result.failure?
    assert_equal "invalid_string", result.issues.first.code
  end

  def test_uuid_valid
    s = AnyVali.string.format("uuid")
    assert_equal "550e8400-e29b-41d4-a716-446655440000", s.parse("550e8400-e29b-41d4-a716-446655440000")
  end

  def test_uuid_invalid
    s = AnyVali.string.format("uuid")
    result = s.safe_parse("not-a-uuid")
    assert result.failure?
  end

  def test_ipv4_valid
    s = AnyVali.string.format("ipv4")
    assert_equal "192.168.1.1", s.parse("192.168.1.1")
  end

  def test_ipv4_rejects_leading_zeros
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("192.168.01.1")
    assert result.failure?
  end

  def test_ipv4_rejects_out_of_range
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("256.1.1.1")
    assert result.failure?
  end

  def test_ipv6_valid_full
    s = AnyVali.string.format("ipv6")
    assert_equal "2001:0db8:85a3:0000:0000:8a2e:0370:7334", s.parse("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
  end

  def test_ipv6_valid_compressed
    s = AnyVali.string.format("ipv6")
    assert_equal "::1", s.parse("::1")
  end

  def test_ipv6_invalid
    s = AnyVali.string.format("ipv6")
    result = s.safe_parse("not:an:ipv6")
    assert result.failure?
  end

  def test_date_valid
    s = AnyVali.string.format("date")
    assert_equal "2024-02-29", s.parse("2024-02-29")
  end

  def test_date_invalid_leap
    s = AnyVali.string.format("date")
    result = s.safe_parse("2023-02-29")
    assert result.failure?
  end

  def test_datetime_valid_z
    s = AnyVali.string.format("date-time")
    assert_equal "2024-01-15T10:30:00Z", s.parse("2024-01-15T10:30:00Z")
  end

  def test_datetime_valid_offset
    s = AnyVali.string.format("date-time")
    assert_equal "2024-01-15T10:30:00+05:30", s.parse("2024-01-15T10:30:00+05:30")
  end

  def test_datetime_rejects_no_timezone
    s = AnyVali.string.format("date-time")
    result = s.safe_parse("2024-01-15T10:30:00")
    assert result.failure?
  end
end
