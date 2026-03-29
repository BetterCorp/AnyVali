use regex::Regex;

/// Validate a string value against a named format.
pub fn validate_format(format: &str, value: &str) -> bool {
    match format {
        "email" => validate_email(value),
        "url" => validate_url(value),
        "uuid" => validate_uuid(value),
        "ipv4" => validate_ipv4(value),
        "ipv6" => validate_ipv6(value),
        "date" => validate_date(value),
        "date-time" => validate_datetime(value),
        _ => true, // Unknown formats pass by default
    }
}

fn validate_email(value: &str) -> bool {
    // Basic email validation: must have @, then a dot after @
    let re = Regex::new(r"^[^\s@]+@[^\s@]+\.[^\s@]+$").unwrap();
    re.is_match(value)
}

fn validate_url(value: &str) -> bool {
    // Must start with http:// or https://
    let re = Regex::new(r"^https?://[^\s]+$").unwrap();
    re.is_match(value)
}

fn validate_uuid(value: &str) -> bool {
    let re = Regex::new(
        r"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$",
    )
    .unwrap();
    re.is_match(value)
}

fn validate_ipv4(value: &str) -> bool {
    let parts: Vec<&str> = value.split('.').collect();
    if parts.len() != 4 {
        return false;
    }
    for part in parts {
        // No leading zeros (except "0" itself)
        if part.len() > 1 && part.starts_with('0') {
            return false;
        }
        match part.parse::<u16>() {
            Ok(n) if n <= 255 => {}
            _ => return false,
        }
    }
    true
}

fn validate_ipv6(value: &str) -> bool {
    // Handle :: expansion
    let parts: Vec<&str> = value.split("::").collect();

    match parts.len() {
        1 => {
            // No ::, must be exactly 8 groups
            let groups: Vec<&str> = value.split(':').collect();
            if groups.len() != 8 {
                return false;
            }
            groups.iter().all(|g| is_valid_ipv6_group(g))
        }
        2 => {
            // Has one ::
            let left: Vec<&str> = if parts[0].is_empty() {
                vec![]
            } else {
                parts[0].split(':').collect()
            };
            let right: Vec<&str> = if parts[1].is_empty() {
                vec![]
            } else {
                parts[1].split(':').collect()
            };
            let total = left.len() + right.len();
            if total > 7 {
                return false;
            }
            left.iter().all(|g| is_valid_ipv6_group(g))
                && right.iter().all(|g| is_valid_ipv6_group(g))
        }
        _ => false, // Multiple :: is invalid
    }
}

fn is_valid_ipv6_group(group: &str) -> bool {
    if group.is_empty() || group.len() > 4 {
        return false;
    }
    group.chars().all(|c| c.is_ascii_hexdigit())
}

fn validate_date(value: &str) -> bool {
    // YYYY-MM-DD
    let re = Regex::new(r"^(\d{4})-(\d{2})-(\d{2})$").unwrap();
    if let Some(caps) = re.captures(value) {
        let year: u32 = caps[1].parse().unwrap();
        let month: u32 = caps[2].parse().unwrap();
        let day: u32 = caps[3].parse().unwrap();
        is_valid_date(year, month, day)
    } else {
        false
    }
}

fn validate_datetime(value: &str) -> bool {
    // YYYY-MM-DDTHH:MM:SS with timezone (Z or +/-HH:MM)
    let re = Regex::new(
        r"^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(\.\d+)?(Z|[+-]\d{2}:\d{2})$",
    )
    .unwrap();
    if let Some(caps) = re.captures(value) {
        let year: u32 = caps[1].parse().unwrap();
        let month: u32 = caps[2].parse().unwrap();
        let day: u32 = caps[3].parse().unwrap();
        let hour: u32 = caps[4].parse().unwrap();
        let minute: u32 = caps[5].parse().unwrap();
        let second: u32 = caps[6].parse().unwrap();
        is_valid_date(year, month, day) && hour < 24 && minute < 60 && second < 60
    } else {
        false
    }
}

fn is_valid_date(year: u32, month: u32, day: u32) -> bool {
    if month < 1 || month > 12 || day < 1 {
        return false;
    }
    let days_in_month = match month {
        1 | 3 | 5 | 7 | 8 | 10 | 12 => 31,
        4 | 6 | 9 | 11 => 30,
        2 => {
            if is_leap_year(year) {
                29
            } else {
                28
            }
        }
        _ => return false,
    };
    day <= days_in_month
}

fn is_leap_year(year: u32) -> bool {
    (year % 4 == 0 && year % 100 != 0) || (year % 400 == 0)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_email_valid() {
        assert!(validate_email("user@example.com"));
        assert!(validate_email("a.b@c.d.e"));
    }

    #[test]
    fn test_email_invalid() {
        assert!(!validate_email("not-an-email"));
        assert!(!validate_email("user@localhost"));
        assert!(!validate_email("@example.com"));
    }

    #[test]
    fn test_url_valid() {
        assert!(validate_url("https://example.com"));
        assert!(validate_url("http://example.com/path?q=1"));
    }

    #[test]
    fn test_url_invalid() {
        assert!(!validate_url("ftp://files.example.com"));
        assert!(!validate_url("not a url"));
    }

    #[test]
    fn test_uuid_valid() {
        assert!(validate_uuid("550e8400-e29b-41d4-a716-446655440000"));
    }

    #[test]
    fn test_uuid_invalid() {
        assert!(!validate_uuid("not-a-uuid"));
        assert!(!validate_uuid("550e8400-e29b-41d4-a716-44665544000")); // too short
    }

    #[test]
    fn test_ipv4_valid() {
        assert!(validate_ipv4("192.168.1.1"));
        assert!(validate_ipv4("0.0.0.0"));
        assert!(validate_ipv4("255.255.255.255"));
    }

    #[test]
    fn test_ipv4_invalid() {
        assert!(!validate_ipv4("192.168.01.1")); // leading zero
        assert!(!validate_ipv4("256.1.1.1")); // out of range
        assert!(!validate_ipv4("1.2.3")); // too few parts
    }

    #[test]
    fn test_ipv6_valid() {
        assert!(validate_ipv6("2001:0db8:85a3:0000:0000:8a2e:0370:7334"));
        assert!(validate_ipv6("::1"));
        assert!(validate_ipv6("fe80::1"));
    }

    #[test]
    fn test_ipv6_invalid() {
        assert!(!validate_ipv6("not:an:ipv6"));
        assert!(!validate_ipv6(":::1")); // triple colon
    }

    #[test]
    fn test_date_valid() {
        assert!(validate_date("2024-02-29")); // leap year
        assert!(validate_date("2024-01-15"));
    }

    #[test]
    fn test_date_invalid() {
        assert!(!validate_date("2023-02-29")); // not a leap year
        assert!(!validate_date("2024-13-01")); // month 13
        assert!(!validate_date("not-a-date"));
    }

    #[test]
    fn test_datetime_valid() {
        assert!(validate_datetime("2024-01-15T10:30:00Z"));
        assert!(validate_datetime("2024-01-15T10:30:00+05:30"));
    }

    #[test]
    fn test_datetime_invalid() {
        assert!(!validate_datetime("2024-01-15T10:30:00")); // no timezone
        assert!(!validate_datetime("not-a-datetime"));
    }
}
