package core

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"time"
)

func ProbeTLS(parent context.Context, host, port string, timeout time.Duration, insecure bool) TLSResult {
	address := net.JoinHostPort(host, port)
	result := TLSResult{Address: address}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	started := time.Now()

	dialer := &net.Dialer{Timeout: timeout}
	rawConn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		result.Error = err.Error()
		result.DurationMS = time.Since(started).Milliseconds()
		return result
	}
	defer rawConn.Close()

	config := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecure} // #nosec G402 -- controlled by an explicit diagnostic flag.
	conn := tls.Client(rawConn, config)
	if err := conn.HandshakeContext(ctx); err != nil {
		result.Error = err.Error()
		result.DurationMS = time.Since(started).Milliseconds()
		return result
	}
	state := conn.ConnectionState()
	result.DurationMS = time.Since(started).Milliseconds()
	result.Protocol = tlsVersionName(state.Version)
	result.CipherSuite = tls.CipherSuiteName(state.CipherSuite)
	result.ALPN = state.NegotiatedProtocol
	result.Verified = len(state.VerifiedChains) > 0 && !insecure
	result.ChainLength = len(state.PeerCertificates)
	result.OCSPStapled = len(state.OCSPResponse) > 0
	if len(state.PeerCertificates) == 0 {
		result.Error = "server did not provide a certificate"
		return result
	}
	cert := state.PeerCertificates[0]
	fillCertificate(&result, cert)
	if !insecure {
		if err := cert.VerifyHostname(host); err != nil {
			result.Error = err.Error()
		}
	}
	return result
}

func fillCertificate(result *TLSResult, cert *x509.Certificate) {
	result.Subject = cert.Subject.String()
	result.Issuer = cert.Issuer.String()
	result.SerialNumber = cert.SerialNumber.Text(16)
	result.DNSNames = cert.DNSNames
	result.NotBefore = cert.NotBefore
	result.NotAfter = cert.NotAfter
	result.DaysRemaining = int(time.Until(cert.NotAfter).Hours() / 24)
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return "unknown"
	}
}
