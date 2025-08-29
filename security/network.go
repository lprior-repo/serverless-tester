package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// NetworkSecurityConfig holds configuration for network security validation
type NetworkSecurityConfig struct {
	RequireHTTPS       bool
	MinTLSVersion      uint16
	RequiredCiphers    []uint16
	CertificateValidation bool
	AllowedHosts       []string
	BlockedHosts       []string
	MaxRedirects       int
	Timeout           time.Duration
	ProxyURL          string
}

// TLSValidationResult represents the result of TLS validation
type TLSValidationResult struct {
	Endpoint          string
	IsSecure          bool
	TLSVersion        uint16
	CipherSuite       uint16
	Certificate       *x509.Certificate
	CertificateChain  []*x509.Certificate
	Issues            []string
	ValidatedAt       time.Time
}

// NetworkSecurityValidator validates network security configurations
type NetworkSecurityValidator struct {
	config NetworkSecurityConfig
	client *http.Client
}

// NewNetworkSecurityValidator creates a new network security validator
func NewNetworkSecurityValidator(config NetworkSecurityConfig) *NetworkSecurityValidator {
	// Set default values
	if config.MinTLSVersion == 0 {
		config.MinTLSVersion = tls.VersionTLS12 // Minimum TLS 1.2
	}
	
	if config.MaxRedirects == 0 {
		config.MaxRedirects = 5
	}
	
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Configure HTTP client with security settings
	tlsConfig := &tls.Config{
		MinVersion:         config.MinTLSVersion,
		InsecureSkipVerify: !config.CertificateValidation,
	}
	
	if len(config.RequiredCiphers) > 0 {
		tlsConfig.CipherSuites = config.RequiredCiphers
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialTimeout:     config.Timeout,
		TLSHandshakeTimeout: config.Timeout,
	}

	// Configure proxy if specified
	if config.ProxyURL != "" {
		if proxyURL, err := url.Parse(config.ProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
			}
			return nil
		},
	}

	return &NetworkSecurityValidator{
		config: config,
		client: client,
	}
}

// ValidateTLSEndpoint validates TLS configuration of an endpoint
func (nsv *NetworkSecurityValidator) ValidateTLSEndpoint(endpoint string) (*TLSValidationResult, error) {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Check if HTTPS is required
	if nsv.config.RequireHTTPS && parsedURL.Scheme != "https" {
		return &TLSValidationResult{
			Endpoint:    endpoint,
			IsSecure:    false,
			Issues:      []string{"HTTPS is required but endpoint uses HTTP"},
			ValidatedAt: time.Now(),
		}, nil
	}

	// Skip TLS validation for non-HTTPS endpoints
	if parsedURL.Scheme != "https" {
		return &TLSValidationResult{
			Endpoint:    endpoint,
			IsSecure:    false,
			Issues:      []string{"Non-HTTPS endpoint"},
			ValidatedAt: time.Now(),
		}, nil
	}

	// Check if host is allowed
	if !nsv.isHostAllowed(parsedURL.Host) {
		return &TLSValidationResult{
			Endpoint:    endpoint,
			IsSecure:    false,
			Issues:      []string{fmt.Sprintf("Host %s is not in allowed list", parsedURL.Host)},
			ValidatedAt: time.Now(),
		}, nil
	}

	// Check if host is blocked
	if nsv.isHostBlocked(parsedURL.Host) {
		return &TLSValidationResult{
			Endpoint:    endpoint,
			IsSecure:    false,
			Issues:      []string{fmt.Sprintf("Host %s is in blocked list", parsedURL.Host)},
			ValidatedAt: time.Now(),
		}, nil
	}

	// Perform TLS handshake
	host := parsedURL.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}

	conn, err := tls.Dial("tcp", host, &tls.Config{
		MinVersion:         nsv.config.MinTLSVersion,
		InsecureSkipVerify: !nsv.config.CertificateValidation,
		ServerName:         parsedURL.Hostname(),
	})
	if err != nil {
		return &TLSValidationResult{
			Endpoint:    endpoint,
			IsSecure:    false,
			Issues:      []string{fmt.Sprintf("TLS handshake failed: %v", err)},
			ValidatedAt: time.Now(),
		}, nil
	}
	defer conn.Close()

	state := conn.ConnectionState()
	result := &TLSValidationResult{
		Endpoint:         endpoint,
		IsSecure:         true,
		TLSVersion:       state.Version,
		CipherSuite:      state.CipherSuite,
		ValidatedAt:      time.Now(),
	}

	// Collect certificate information
	if len(state.PeerCertificates) > 0 {
		result.Certificate = state.PeerCertificates[0]
		result.CertificateChain = state.PeerCertificates
	}

	// Validate TLS version
	if state.Version < nsv.config.MinTLSVersion {
		result.Issues = append(result.Issues, 
			fmt.Sprintf("TLS version %s is below minimum required %s", 
				tlsVersionString(state.Version), 
				tlsVersionString(nsv.config.MinTLSVersion)))
	}

	// Validate cipher suite
	if len(nsv.config.RequiredCiphers) > 0 {
		cipherAllowed := false
		for _, allowedCipher := range nsv.config.RequiredCiphers {
			if state.CipherSuite == allowedCipher {
				cipherAllowed = true
				break
			}
		}
		if !cipherAllowed {
			result.Issues = append(result.Issues, 
				fmt.Sprintf("Cipher suite %s is not in allowed list", 
					tls.CipherSuiteName(state.CipherSuite)))
		}
	}

	// Validate certificate if certificate validation is enabled
	if nsv.config.CertificateValidation && result.Certificate != nil {
		issues := nsv.validateCertificate(result.Certificate, parsedURL.Hostname())
		result.Issues = append(result.Issues, issues...)
	}

	// Determine if the endpoint is secure based on issues
	result.IsSecure = len(result.Issues) == 0

	return result, nil
}

// ValidateHTTPSRedirect checks if HTTP endpoints redirect to HTTPS
func (nsv *NetworkSecurityValidator) ValidateHTTPSRedirect(httpEndpoint string) error {
	parsedURL, err := url.Parse(httpEndpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	if parsedURL.Scheme != "http" {
		return fmt.Errorf("endpoint must use HTTP scheme for redirect validation")
	}

	// Create a client that doesn't follow redirects automatically
	client := &http.Client{
		Transport: nsv.client.Transport,
		Timeout:   nsv.config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	resp, err := client.Get(httpEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to HTTP endpoint: %w", err)
	}
	defer resp.Body.Close()

	// Check for redirect status codes
	if resp.StatusCode != http.StatusMovedPermanently && 
	   resp.StatusCode != http.StatusFound && 
	   resp.StatusCode != http.StatusSeeOther && 
	   resp.StatusCode != http.StatusTemporaryRedirect && 
	   resp.StatusCode != http.StatusPermanentRedirect {
		return fmt.Errorf("HTTP endpoint does not redirect to HTTPS (status: %d)", resp.StatusCode)
	}

	// Check if redirect location is HTTPS
	location := resp.Header.Get("Location")
	if location == "" {
		return fmt.Errorf("redirect response missing Location header")
	}

	redirectURL, err := url.Parse(location)
	if err != nil {
		return fmt.Errorf("invalid redirect URL: %w", err)
	}

	if redirectURL.Scheme != "https" {
		return fmt.Errorf("HTTP endpoint redirects to non-HTTPS URL: %s", location)
	}

	return nil
}

// ValidateSecurityHeaders checks for security-related HTTP headers
func (nsv *NetworkSecurityValidator) ValidateSecurityHeaders(endpoint string) (map[string]string, []string, error) {
	resp, err := nsv.client.Get(endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch endpoint: %w", err)
	}
	defer resp.Body.Close()

	securityHeaders := map[string]string{
		"Strict-Transport-Security": resp.Header.Get("Strict-Transport-Security"),
		"Content-Security-Policy":   resp.Header.Get("Content-Security-Policy"),
		"X-Frame-Options":          resp.Header.Get("X-Frame-Options"),
		"X-Content-Type-Options":   resp.Header.Get("X-Content-Type-Options"),
		"Referrer-Policy":          resp.Header.Get("Referrer-Policy"),
		"X-XSS-Protection":         resp.Header.Get("X-XSS-Protection"),
	}

	var issues []string

	// Check for missing critical security headers
	if securityHeaders["Strict-Transport-Security"] == "" {
		issues = append(issues, "Missing Strict-Transport-Security header")
	}

	if securityHeaders["Content-Security-Policy"] == "" {
		issues = append(issues, "Missing Content-Security-Policy header")
	}

	if securityHeaders["X-Frame-Options"] == "" {
		issues = append(issues, "Missing X-Frame-Options header")
	}

	if securityHeaders["X-Content-Type-Options"] == "" {
		issues = append(issues, "Missing X-Content-Type-Options header")
	}

	return securityHeaders, issues, nil
}

// isHostAllowed checks if a host is in the allowed list
func (nsv *NetworkSecurityValidator) isHostAllowed(host string) bool {
	if len(nsv.config.AllowedHosts) == 0 {
		return true // No restrictions if no allowed hosts specified
	}

	hostname := strings.Split(host, ":")[0] // Remove port if present

	for _, allowedHost := range nsv.config.AllowedHosts {
		if matched, _ := regexp.MatchString(allowedHost, hostname); matched {
			return true
		}
	}

	return false
}

// isHostBlocked checks if a host is in the blocked list
func (nsv *NetworkSecurityValidator) isHostBlocked(host string) bool {
	hostname := strings.Split(host, ":")[0] // Remove port if present

	for _, blockedHost := range nsv.config.BlockedHosts {
		if matched, _ := regexp.MatchString(blockedHost, hostname); matched {
			return true
		}
	}

	return false
}

// validateCertificate validates SSL/TLS certificate
func (nsv *NetworkSecurityValidator) validateCertificate(cert *x509.Certificate, hostname string) []string {
	var issues []string

	// Check certificate expiration
	now := time.Now()
	if now.After(cert.NotAfter) {
		issues = append(issues, "Certificate has expired")
	} else if now.Add(30*24*time.Hour).After(cert.NotAfter) {
		issues = append(issues, "Certificate expires within 30 days")
	}

	// Check if certificate is valid for the hostname
	if err := cert.VerifyHostname(hostname); err != nil {
		issues = append(issues, fmt.Sprintf("Certificate hostname verification failed: %v", err))
	}

	// Check certificate validity period
	if cert.NotBefore.After(now) {
		issues = append(issues, "Certificate is not yet valid")
	}

	// Check key usage
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		issues = append(issues, "Certificate missing digital signature key usage")
	}

	// Check for weak signature algorithm
	weakAlgorithms := []x509.SignatureAlgorithm{
		x509.MD2WithRSA, x509.MD5WithRSA, x509.SHA1WithRSA,
	}
	for _, weakAlg := range weakAlgorithms {
		if cert.SignatureAlgorithm == weakAlg {
			issues = append(issues, fmt.Sprintf("Certificate uses weak signature algorithm: %s", cert.SignatureAlgorithm))
			break
		}
	}

	return issues
}

// tlsVersionString converts TLS version number to string
func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (%d)", version)
	}
}

// ValidateIPWhitelist validates if an IP address is in the allowed list
func ValidateIPWhitelist(ipAddress string, allowedIPs []string) bool {
	if len(allowedIPs) == 0 {
		return true // No restrictions if no allowed IPs specified
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false // Invalid IP address
	}

	for _, allowedIP := range allowedIPs {
		// Check for CIDR notation
		if strings.Contains(allowedIP, "/") {
			_, network, err := net.ParseCIDR(allowedIP)
			if err == nil && network.Contains(ip) {
				return true
			}
		} else {
			// Exact IP match
			if allowedIP == ipAddress {
				return true
			}
		}
	}

	return false
}

// ValidatePortRange checks if a port is within allowed ranges
func ValidatePortRange(port int, allowedRanges []string) bool {
	if len(allowedRanges) == 0 {
		return true // No restrictions if no ranges specified
	}

	for _, rangeStr := range allowedRanges {
		if strings.Contains(rangeStr, "-") {
			// Port range (e.g., "8000-9000")
			parts := strings.Split(rangeStr, "-")
			if len(parts) == 2 {
				var start, end int
				if fmt.Sscanf(parts[0], "%d", &start) == 1 &&
				   fmt.Sscanf(parts[1], "%d", &end) == 1 {
					if port >= start && port <= end {
						return true
					}
				}
			}
		} else {
			// Single port
			var allowedPort int
			if fmt.Sscanf(rangeStr, "%d", &allowedPort) == 1 && port == allowedPort {
				return true
			}
		}
	}

	return false
}

// CreateSecureHTTPClient creates an HTTP client with security best practices
func CreateSecureHTTPClient(config NetworkSecurityConfig) *http.Client {
	tlsConfig := &tls.Config{
		MinVersion:               config.MinTLSVersion,
		PreferServerCipherSuites: true,
		InsecureSkipVerify:      !config.CertificateValidation,
	}

	if len(config.RequiredCiphers) > 0 {
		tlsConfig.CipherSuites = config.RequiredCiphers
	}

	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: config.Timeout,
		DialTimeout:         config.Timeout,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
			}
			return nil
		},
	}
}

// GetRecommendedCipherSuites returns recommended cipher suites for security
func GetRecommendedCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}
}