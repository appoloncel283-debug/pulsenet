package core

import (
	"net/url"
	"strings"
)

func AuditHeaders(result HTTPResult) HeaderAudit {
	checks := make([]HeaderCheck, 0, 8)
	headers := result.ResponseHeaders
	isHTTPS := false
	if parsed, err := url.Parse(result.FinalURL); err == nil {
		isHTTPS = strings.EqualFold(parsed.Scheme, "https")
	}
	add := func(name string, weight int, passed bool, value, passMessage, failMessage string) {
		message := failMessage
		if passed {
			message = passMessage
		}
		checks = append(checks, HeaderCheck{Name: name, Weight: weight, Passed: passed, Value: value, Message: message})
	}

	hsts := headers["Strict-Transport-Security"]
	if isHTTPS {
		add("HSTS", 18, hsts != "", hsts, "HTTPS downgrade protection is enabled.", "Add Strict-Transport-Security on HTTPS responses.")
	}
	csp := headers["Content-Security-Policy"]
	add("Content Security Policy", 20, csp != "", csp, "A Content-Security-Policy is present.", "Add a restrictive Content-Security-Policy.")
	nosniff := headers["X-Content-Type-Options"]
	add("MIME sniffing protection", 12, strings.EqualFold(strings.TrimSpace(nosniff), "nosniff"), nosniff, "MIME sniffing is disabled.", "Set X-Content-Type-Options: nosniff.")
	xfo := headers["X-Frame-Options"]
	frameAncestors := strings.Contains(strings.ToLower(csp), "frame-ancestors")
	add("Clickjacking protection", 14, xfo != "" || frameAncestors, firstNonEmpty(xfo, csp), "Framing restrictions are present.", "Set X-Frame-Options or CSP frame-ancestors.")
	referrer := headers["Referrer-Policy"]
	add("Referrer policy", 10, referrer != "", referrer, "A Referrer-Policy is present.", "Set a privacy-conscious Referrer-Policy.")
	permissions := headers["Permissions-Policy"]
	add("Permissions policy", 8, permissions != "", permissions, "Browser feature permissions are controlled.", "Add a Permissions-Policy header.")
	coop := headers["Cross-Origin-Opener-Policy"]
	add("Cross-origin opener policy", 8, coop != "", coop, "Cross-origin window isolation is configured.", "Consider Cross-Origin-Opener-Policy for sensitive applications.")
	cookies := headers["Set-Cookie"]
	cookiePassed := cookies == "" || (!isHTTPS || strings.Contains(strings.ToLower(cookies), "secure")) && strings.Contains(strings.ToLower(cookies), "httponly")
	add("Cookie flags", 10, cookiePassed, cookies, "Observed cookies use appropriate security flags or no cookies were set.", "Review Set-Cookie flags; use Secure and HttpOnly where appropriate.")

	possible := 0
	earned := 0
	for _, check := range checks {
		possible += check.Weight
		if check.Passed {
			earned += check.Weight
		}
	}
	score := 100
	if possible > 0 {
		score = earned * 100 / possible
	}
	return HeaderAudit{Score: score, Grade: gradeForScore(score), Checks: checks}
}

func gradeForScore(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 65:
		return "C"
	case score >= 50:
		return "D"
	default:
		return "F"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
