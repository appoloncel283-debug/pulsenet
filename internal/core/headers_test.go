package core

import "testing"

func TestAuditHeadersStrong(t *testing.T) {
	result := HTTPResult{
		FinalURL: "https://example.com",
		ResponseHeaders: map[string]string{
			"Strict-Transport-Security":  "max-age=31536000; includeSubDomains",
			"Content-Security-Policy":    "default-src 'self'; frame-ancestors 'none'",
			"X-Content-Type-Options":     "nosniff",
			"Referrer-Policy":            "strict-origin-when-cross-origin",
			"Permissions-Policy":         "camera=()",
			"Cross-Origin-Opener-Policy": "same-origin",
		},
	}
	audit := AuditHeaders(result)
	if audit.Score < 80 {
		t.Fatalf("score = %d", audit.Score)
	}
}

func TestAuditHeadersEmpty(t *testing.T) {
	audit := AuditHeaders(HTTPResult{FinalURL: "https://example.com", ResponseHeaders: map[string]string{}})
	if audit.Score >= 50 {
		t.Fatalf("score = %d", audit.Score)
	}
}
