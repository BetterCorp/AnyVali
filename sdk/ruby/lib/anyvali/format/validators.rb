# frozen_string_literal: true

require "uri"
require "ipaddr"
require "date"

module AnyVali
  module Format
    module Validators
      EMAIL_REGEX = /\A[^@\s]+@[^@\s]+\.[^@\s]+\z/
      UUID_REGEX = /\A[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\z/i
      # Strict IPv4: no leading zeros, octets 0-255
      IPV4_REGEX = /\A(?:(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)\.){3}(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)\z/
      DATETIME_REGEX = /\A\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})\z/
      DATE_REGEX = /\A\d{4}-\d{2}-\d{2}\z/

      module_function

      def validate(value, format, path, issues)
        valid = case format
                when "email" then valid_email?(value)
                when "url" then valid_url?(value)
                when "uuid" then valid_uuid?(value)
                when "ipv4" then valid_ipv4?(value)
                when "ipv6" then valid_ipv6?(value)
                when "date" then valid_date?(value)
                when "date-time" then valid_datetime?(value)
                else false
                end

        unless valid
          issues << ValidationIssue.new(
            code: IssueCodes::INVALID_STRING,
            path: path,
            expected: format,
            received: value
          )
        end
      end

      def valid_email?(value)
        EMAIL_REGEX.match?(value)
      end

      def valid_url?(value)
        uri = URI.parse(value)
        %w[http https].include?(uri.scheme) && !uri.host.nil?
      rescue URI::InvalidURIError
        false
      end

      def valid_uuid?(value)
        UUID_REGEX.match?(value)
      end

      def valid_ipv4?(value)
        # Must match strict pattern (no leading zeros)
        return false unless IPV4_REGEX.match?(value)
        # Check for leading zeros explicitly
        parts = value.split(".")
        parts.length == 4 && parts.all? { |p| p == "0" || !p.start_with?("0") }
      end

      def valid_ipv6?(value)
        addr = IPAddr.new(value)
        addr.ipv6?
      rescue IPAddr::InvalidAddressError, IPAddr::AddressFamilyError
        false
      rescue ArgumentError
        false
      end

      def valid_date?(value)
        return false unless DATE_REGEX.match?(value)
        Date.parse(value)
        true
      rescue Date::Error, ArgumentError
        false
      end

      def valid_datetime?(value)
        return false unless DATETIME_REGEX.match?(value)
        # Validate the date part
        date_part = value[0..9]
        valid_date?(date_part)
      end
    end
  end
end
