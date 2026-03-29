package anyvali

import "testing"

func TestValidateFormatEmail(t *testing.T) {
	valid := []string{
		"test@example.com",
		"user.name+tag@domain.co.uk",
		"a@b.c",
		"user@localhost",
	}
	invalid := []string{
		"",
		"not-an-email",
		"@missing.user",
		"missing@",
		"user@@domain.com",
		"spaces in@domain.com",
	}
	for _, v := range valid {
		if !validateFormat(v, "email") {
			t.Errorf("expected %q to be valid email", v)
		}
	}
	for _, v := range invalid {
		if validateFormat(v, "email") {
			t.Errorf("expected %q to be invalid email", v)
		}
	}
}

func TestValidateFormatURL(t *testing.T) {
	valid := []string{
		"http://example.com",
		"https://example.com",
		"https://example.com/path",
		"http://localhost",
		"https://example.com/path?q=1#hash",
		"http://a",
	}
	invalid := []string{
		"",
		"example.com",
		"ftp://example.com",
		"http://",
		"https://",
		"not-a-url",
	}
	for _, v := range valid {
		if !validateFormat(v, "url") {
			t.Errorf("expected %q to be valid URL", v)
		}
	}
	for _, v := range invalid {
		if validateFormat(v, "url") {
			t.Errorf("expected %q to be invalid URL", v)
		}
	}
}

func TestValidateFormatUUID(t *testing.T) {
	valid := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"00000000-0000-0000-0000-000000000000",
		"FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF",
	}
	invalid := []string{
		"",
		"not-a-uuid",
		"550e8400-e29b-41d4-a716",
		"550e8400e29b41d4a716446655440000",
		"550e8400-e29b-41d4-a716-44665544000g",
	}
	for _, v := range valid {
		if !validateFormat(v, "uuid") {
			t.Errorf("expected %q to be valid UUID", v)
		}
	}
	for _, v := range invalid {
		if validateFormat(v, "uuid") {
			t.Errorf("expected %q to be invalid UUID", v)
		}
	}
}

func TestValidateFormatIPv4(t *testing.T) {
	valid := []string{
		"192.168.1.1",
		"0.0.0.0",
		"255.255.255.255",
		"127.0.0.1",
	}
	invalid := []string{
		"",
		"256.1.1.1",
		"1.2.3",
		"1.2.3.4.5",
		"::1",
		"not-an-ip",
	}
	for _, v := range valid {
		if !validateFormat(v, "ipv4") {
			t.Errorf("expected %q to be valid IPv4", v)
		}
	}
	for _, v := range invalid {
		if validateFormat(v, "ipv4") {
			t.Errorf("expected %q to be invalid IPv4", v)
		}
	}
}

func TestValidateFormatIPv6(t *testing.T) {
	valid := []string{
		"::1",
		"fe80::1",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		"::",
	}
	invalid := []string{
		"",
		"192.168.1.1",
		"not-an-ipv6",
		"gggg::1",
	}
	for _, v := range valid {
		if !validateFormat(v, "ipv6") {
			t.Errorf("expected %q to be valid IPv6", v)
		}
	}
	for _, v := range invalid {
		if validateFormat(v, "ipv6") {
			t.Errorf("expected %q to be invalid IPv6", v)
		}
	}
}

func TestValidateFormatDate(t *testing.T) {
	valid := []string{
		"2024-01-15",
		"2024-12-31",
		"2000-01-01",
	}
	invalid := []string{
		"",
		"2024-13-01",
		"2024-00-01",
		"2024-01-32",
		"2024-1-1",
		"not-a-date",
		"01-15-2024",
	}
	for _, v := range valid {
		if !validateFormat(v, "date") {
			t.Errorf("expected %q to be valid date", v)
		}
	}
	for _, v := range invalid {
		if validateFormat(v, "date") {
			t.Errorf("expected %q to be invalid date", v)
		}
	}
}

func TestValidateFormatDateTime(t *testing.T) {
	valid := []string{
		"2024-01-15T10:30:00Z",
		"2024-01-15T10:30:00+05:30",
		"2024-01-15T10:30:00-08:00",
		"2024-01-15T10:30:00.123Z",
		"2024-01-15T23:59:59Z",
	}
	invalid := []string{
		"",
		"2024-01-15",
		"2024-01-15T10:30:00",
		"not-a-datetime",
		"2024-01-15 10:30:00Z",
		"2024-13-15T10:30:00Z",
	}
	for _, v := range valid {
		if !validateFormat(v, "date-time") {
			t.Errorf("expected %q to be valid date-time", v)
		}
	}
	for _, v := range invalid {
		if validateFormat(v, "date-time") {
			t.Errorf("expected %q to be invalid date-time", v)
		}
	}
}

func TestValidateFormatUnknown(t *testing.T) {
	if validateFormat("anything", "unknown-format") {
		t.Fatal("expected unknown format to return false")
	}
	if validateFormat("anything", "") {
		t.Fatal("expected empty format to return false")
	}
}

func TestValidateURLEdgeCases(t *testing.T) {
	// URL with path, query, hash
	if !validateURL("https://example.com/path?q=1#fragment") {
		t.Fatal("expected valid URL with path/query/hash")
	}
	// HTTP scheme
	if !validateURL("http://a") {
		t.Fatal("expected valid minimal http URL")
	}
	// No scheme
	if validateURL("example.com") {
		t.Fatal("expected invalid URL without scheme")
	}
}
