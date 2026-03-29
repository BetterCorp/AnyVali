#pragma once

#include <string>
#include <regex>
#include <cstdlib>
#include <sstream>
#include <vector>

namespace anyvali {
namespace format {

inline bool validate_email(const std::string& s) {
    // Basic email: something@something.something
    static const std::regex email_re(R"([^@\s]+@[^@\s]+\.[^@\s]+)");
    return std::regex_match(s, email_re);
}

inline bool validate_url(const std::string& s) {
    // Only http and https
    static const std::regex url_re(R"(https?://[^\s]+)");
    return std::regex_match(s, url_re);
}

inline bool validate_uuid(const std::string& s) {
    static const std::regex uuid_re(
        R"([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})");
    return std::regex_match(s, uuid_re);
}

inline bool validate_ipv4(const std::string& s) {
    // Split by '.'
    std::vector<std::string> parts;
    std::istringstream iss(s);
    std::string part;
    while (std::getline(iss, part, '.')) {
        parts.push_back(part);
    }
    if (parts.size() != 4) return false;
    for (const auto& p : parts) {
        if (p.empty() || p.size() > 3) return false;
        // Reject leading zeros (except for "0" itself)
        if (p.size() > 1 && p[0] == '0') return false;
        for (char c : p) {
            if (c < '0' || c > '9') return false;
        }
        int val = std::stoi(p);
        if (val < 0 || val > 255) return false;
    }
    return true;
}

inline bool validate_ipv6(const std::string& s) {
    // Allow compressed forms with ::
    // Simple validation: must have 2-8 groups of hex separated by ':'
    // with at most one '::'
    if (s.empty()) return false;

    // Count double colons
    size_t dc = 0;
    size_t pos = 0;
    while ((pos = s.find("::", pos)) != std::string::npos) {
        dc++;
        pos += 2;
    }
    if (dc > 1) return false;

    // Split by ':'
    std::vector<std::string> parts;
    std::istringstream iss(s);
    std::string part;
    // Handle leading/trailing ::
    std::string work = s;
    // Replace :: with a placeholder
    if (dc == 1) {
        size_t dpos = work.find("::");
        std::string left = work.substr(0, dpos);
        std::string right = work.substr(dpos + 2);

        std::vector<std::string> left_parts, right_parts;
        if (!left.empty()) {
            std::istringstream ls(left);
            while (std::getline(ls, part, ':')) left_parts.push_back(part);
        }
        if (!right.empty()) {
            std::istringstream rs(right);
            while (std::getline(rs, part, ':')) right_parts.push_back(part);
        }

        int total = static_cast<int>(left_parts.size() + right_parts.size());
        if (total > 7) return false;
        int fill = 8 - total;

        for (const auto& p : left_parts) parts.push_back(p);
        for (int i = 0; i < fill; i++) parts.push_back("0");
        for (const auto& p : right_parts) parts.push_back(p);
    } else {
        std::istringstream ss(work);
        while (std::getline(ss, part, ':')) parts.push_back(part);
    }

    if (parts.size() != 8) return false;

    static const std::regex hex_re(R"([0-9a-fA-F]{1,4})");
    for (const auto& p : parts) {
        if (p == "0") continue;
        if (!std::regex_match(p, hex_re)) return false;
    }
    return true;
}

inline bool is_leap_year(int year) {
    return (year % 4 == 0 && year % 100 != 0) || (year % 400 == 0);
}

inline int days_in_month(int year, int month) {
    static const int days[] = {0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31};
    if (month == 2 && is_leap_year(year)) return 29;
    return days[month];
}

inline bool validate_date(const std::string& s) {
    // YYYY-MM-DD
    static const std::regex date_re(R"((\d{4})-(\d{2})-(\d{2}))");
    std::smatch m;
    if (!std::regex_match(s, m, date_re)) return false;
    int year = std::stoi(m[1]);
    int month = std::stoi(m[2]);
    int day = std::stoi(m[3]);
    if (month < 1 || month > 12) return false;
    if (day < 1 || day > days_in_month(year, month)) return false;
    return true;
}

inline bool validate_datetime(const std::string& s) {
    // ISO 8601: YYYY-MM-DDTHH:MM:SS with timezone (Z or +HH:MM or -HH:MM)
    static const std::regex dt_re(
        R"((\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.(\d+))?(Z|[+-]\d{2}:\d{2}))");
    std::smatch m;
    if (!std::regex_match(s, m, dt_re)) return false;
    int year = std::stoi(m[1]);
    int month = std::stoi(m[2]);
    int day = std::stoi(m[3]);
    int hour = std::stoi(m[4]);
    int minute = std::stoi(m[5]);
    int second = std::stoi(m[6]);
    if (month < 1 || month > 12) return false;
    if (day < 1 || day > days_in_month(year, month)) return false;
    if (hour > 23 || minute > 59 || second > 59) return false;
    return true;
}

inline bool validate_format(const std::string& format, const std::string& value) {
    if (format == "email") return validate_email(value);
    if (format == "url") return validate_url(value);
    if (format == "uuid") return validate_uuid(value);
    if (format == "ipv4") return validate_ipv4(value);
    if (format == "ipv6") return validate_ipv6(value);
    if (format == "date") return validate_date(value);
    if (format == "date-time") return validate_datetime(value);
    return false;
}

} // namespace format
} // namespace anyvali
