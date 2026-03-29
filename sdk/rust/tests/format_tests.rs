use anyvali::*;
use serde_json::json;

#[test]
fn email_format_valid() {
    let s = string().format("email");
    assert!(s.parse(&json!("user@example.com")).is_ok());
}

#[test]
fn email_format_rejects_no_at() {
    let s = string().format("email");
    let result = s.safe_parse(&json!("not-an-email"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_string");
    assert_eq!(result.issues[0].expected, "email");
}

#[test]
fn email_format_rejects_no_dot() {
    let s = string().format("email");
    let result = s.safe_parse(&json!("user@localhost"));
    assert!(!result.success);
}

#[test]
fn url_format_accepts_https() {
    let s = string().format("url");
    assert!(s.parse(&json!("https://example.com")).is_ok());
}

#[test]
fn url_format_accepts_http() {
    let s = string().format("url");
    assert!(s.parse(&json!("http://example.com/path?q=1")).is_ok());
}

#[test]
fn url_format_rejects_ftp() {
    let s = string().format("url");
    let result = s.safe_parse(&json!("ftp://files.example.com"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_string");
    assert_eq!(result.issues[0].expected, "url");
}

#[test]
fn uuid_format_valid() {
    let s = string().format("uuid");
    assert!(s
        .parse(&json!("550e8400-e29b-41d4-a716-446655440000"))
        .is_ok());
}

#[test]
fn uuid_format_invalid() {
    let s = string().format("uuid");
    let result = s.safe_parse(&json!("not-a-uuid"));
    assert!(!result.success);
    assert_eq!(result.issues[0].expected, "uuid");
}

#[test]
fn ipv4_format_valid() {
    let s = string().format("ipv4");
    assert!(s.parse(&json!("192.168.1.1")).is_ok());
}

#[test]
fn ipv4_format_rejects_leading_zeros() {
    let s = string().format("ipv4");
    let result = s.safe_parse(&json!("192.168.01.1"));
    assert!(!result.success);
}

#[test]
fn ipv4_format_rejects_out_of_range() {
    let s = string().format("ipv4");
    let result = s.safe_parse(&json!("256.1.1.1"));
    assert!(!result.success);
}

#[test]
fn ipv6_format_valid() {
    let s = string().format("ipv6");
    assert!(s
        .parse(&json!("2001:0db8:85a3:0000:0000:8a2e:0370:7334"))
        .is_ok());
}

#[test]
fn ipv6_format_compressed() {
    let s = string().format("ipv6");
    assert!(s.parse(&json!("::1")).is_ok());
}

#[test]
fn ipv6_format_invalid() {
    let s = string().format("ipv6");
    let result = s.safe_parse(&json!("not:an:ipv6"));
    assert!(!result.success);
}

#[test]
fn date_format_valid() {
    let s = string().format("date");
    assert!(s.parse(&json!("2024-02-29")).is_ok()); // leap year
}

#[test]
fn date_format_rejects_invalid_leap_day() {
    let s = string().format("date");
    let result = s.safe_parse(&json!("2023-02-29")); // not a leap year
    assert!(!result.success);
}

#[test]
fn datetime_format_valid_z() {
    let s = string().format("date-time");
    assert!(s.parse(&json!("2024-01-15T10:30:00Z")).is_ok());
}

#[test]
fn datetime_format_valid_offset() {
    let s = string().format("date-time");
    assert!(s.parse(&json!("2024-01-15T10:30:00+05:30")).is_ok());
}

#[test]
fn datetime_format_rejects_no_timezone() {
    let s = string().format("date-time");
    let result = s.safe_parse(&json!("2024-01-15T10:30:00"));
    assert!(!result.success);
    assert_eq!(result.issues[0].expected, "date-time");
}

#[test]
fn format_combined_with_other_constraints() {
    let s = string().format("email").min_length(5);
    assert!(s.parse(&json!("a@b.c")).is_ok());
    assert!(!s.safe_parse(&json!("a@b")).success);
}

#[test]
fn unknown_format_passes() {
    // Unknown format names should pass by default
    let s = string().format("custom-format");
    assert!(s.parse(&json!("anything")).is_ok());
}
