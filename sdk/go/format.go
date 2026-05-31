package anyvali

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	dateRegex     = regexp.MustCompile(`^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])$`)
	dateTimeRegex = regexp.MustCompile(`^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$`)
)

// validateFormat checks if a string value matches the specified format.
func validateFormat(value string, format string) bool {
	switch format {
	case "email":
		return emailRegex.MatchString(value)
	case "url":
		return validateURL(value)
	case "uuid":
		return uuidRegex.MatchString(value)
	case "ipv4":
		return validateIPv4(value)
	case "ipv6":
		return validateIPv6(value)
	case "date":
		return dateRegex.MatchString(value) && isValidCalendarDate(value)
	case "date-time":
		return dateTimeRegex.MatchString(value) && isValidCalendarDate(value[:10])
	default:
		return false
	}
}

func validateURL(value string) bool {
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		return false
	}
	// Must have something after the scheme
	var rest string
	if strings.HasPrefix(value, "https://") {
		rest = value[8:]
	} else {
		rest = value[7:]
	}
	if len(rest) == 0 {
		return false
	}
	// Must have a host part (at least one dot or localhost)
	hostEnd := strings.IndexAny(rest, "/?#")
	host := rest
	if hostEnd >= 0 {
		host = rest[:hostEnd]
	}
	if len(host) == 0 {
		return false
	}
	return true
}

func isValidCalendarDate(dateStr string) bool {
	// dateStr must be YYYY-MM-DD
	if len(dateStr) != 10 {
		return false
	}
	year, err1 := strconv.Atoi(dateStr[0:4])
	month, err2 := strconv.Atoi(dateStr[5:7])
	day, err3 := strconv.Atoi(dateStr[8:10])
	if err1 != nil || err2 != nil || err3 != nil {
		return false
	}
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return t.Year() == year && int(t.Month()) == month && t.Day() == day
}

func validateIPv4(value string) bool {
	ip := net.ParseIP(value)
	if ip == nil {
		return false
	}
	return ip.To4() != nil && strings.Count(value, ".") == 3 && !strings.Contains(value, ":")
}

func validateIPv6(value string) bool {
	ip := net.ParseIP(value)
	if ip == nil {
		return false
	}
	return strings.Contains(value, ":")
}
