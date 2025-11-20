package telemetry

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Token sanitization
		{
			name:     "API token removal - sk prefix",
			input:    "authentication failed with token sk_abc123def456",
			expected: "authentication failed with token [TOKEN]",
		},
		{
			name:     "GitHub token removal",
			input:    "git push failed: ghp_abc123def456xyz",
			expected: "git push failed: [TOKEN]",
		},
		{
			name:     "GitLab token removal",
			input:    "failed with glpat-abc123def456",
			expected: "failed with [TOKEN]",
		},
		{
			name:     "NPM token removal",
			input:    "npm auth failed npm_abc123def456",
			expected: "npm auth failed [TOKEN]",
		},

		// AWS key sanitization
		{
			name:     "AWS key removal",
			input:    "AWS auth failed: AKIAIOSFODNN7EXAMPLE",
			expected: "AWS auth failed: [AWS_KEY]",
		},

		// Email sanitization
		{
			name:     "Email removal",
			input:    "sent notification to user@example.com",
			expected: "sent notification to [EMAIL]",
		},
		{
			name:     "Email with subdomain",
			input:    "error from admin@mail.company.co.uk",
			expected: "error from [EMAIL]",
		},

		// URL with credentials
		{
			name:     "URL with credentials",
			input:    "failed to connect to https://admin:password123@example.com/api",
			expected: "failed to connect to [URL]",
		},
		{
			name:     "HTTP URL with credentials",
			input:    "http://user:secret@192.168.1.1:8080",
			expected: "[URL]",
		},

		// Database connection strings
		{
			name:     "PostgreSQL connection string",
			input:    "connection error: postgresql://user:pass@localhost:5432/db",
			expected: "connection error: postgresql://[CONNECTION]",
		},
		{
			name:     "MongoDB connection string",
			input:    "failed to connect: mongodb://admin:secret@cluster0.mongodb.net/db",
			expected: "failed to connect: mongodb://[CONNECTION]",
		},
		{
			name:     "MongoDB SRV connection string",
			input:    "error with mongodb+srv://user:pass@cluster.mongodb.net/db",
			expected: "error with mongodb+srv://[CONNECTION]",
		},
		{
			name:     "MySQL connection string",
			input:    "mysql://root:password@localhost:3306/database",
			expected: "mysql://[CONNECTION]",
		},
		{
			name:     "Redis connection string",
			input:    "redis://user:password@localhost:6379",
			expected: "redis://[CONNECTION]",
		},

		// IP address sanitization
		{
			name:     "IPv4 address removal",
			input:    "connection refused to 192.168.1.100",
			expected: "connection refused to [IP]",
		},
		{
			name:     "Multiple IPv4 addresses",
			input:    "failed to connect from 10.0.0.1 to 172.16.0.254",
			expected: "failed to connect from [IP] to [IP]",
		},
		{
			name:     "IPv6 address removal - full",
			input:    "connection to 2001:0db8:85a3:0000:0000:8a2e:0370:7334 failed",
			expected: "connection to [IP] failed",
		},
		{
			name:     "IPv6 address removal - compressed",
			input:    "error connecting to fe80::1",
			expected: "error connecting to [IP]",
		},

		// File path sanitization
		{
			name:     "Unix home directory",
			input:    "failed to read /home/john/.config/secret.yaml",
			expected: "failed to read [HOME]/.config/secret.yaml",
		},
		{
			name:     "macOS home directory",
			input:    "cannot access /Users/alice/Documents/private.txt",
			expected: "cannot access [HOME]/Documents/private.txt",
		},
		{
			name:     "Windows home directory",
			input:    "failed to open C:\\Users\\Bob\\Desktop\\secret.txt",
			expected: "failed to open [HOME]\\Desktop\\secret.txt",
		},
		{
			name:     "Temp directory",
			input:    "error writing /tmp/sensitive-data-12345.txt",
			expected: "error writing [TEMP]",
		},
		{
			name:     "Var temp directory",
			input:    "failed to create /var/tmp/upload-file.dat",
			expected: "failed to create [TEMP]",
		},
		{
			name:     "Home relative path - single segment",
			input:    "cannot read ~/config.yaml",
			expected: "cannot read [PATH]",
		},
		{
			name:     "Home relative path - multiple segments",
			input:    "error accessing ~/.config/app/settings.json",
			expected: "error accessing [PATH]",
		},
		{
			name:     "Absolute path with multiple segments",
			input:    "failed to load /usr/local/bin/myapp",
			expected: "failed to load [PATH]",
		},
		{
			name:     "Windows absolute path",
			input:    "error reading C:\\Program Files\\MyApp\\config.ini",
			expected: "error reading [PATH]",
		},

		// API paths should NOT be sanitized
		{
			name:     "API endpoint should not be sanitized",
			input:    "failed to call /v3/api/endpoint",
			expected: "failed to call /v3/api/endpoint",
		},
		{
			name:     "API path with query should not be sanitized",
			input:    "GET /api/v1/users?id=123 returned 404",
			expected: "GET /api/v1/users?id=123 returned 404",
		},

		// Password/secret sanitization
		{
			name:     "Password in error message",
			input:    "authentication failed: password=secret123 invalid",
			expected: "authentication failed: password=[REDACTED] invalid",
		},
		{
			name:     "Password with colon separator",
			input:    "error: password: mysecretpassword invalid",
			expected: "error: password=[REDACTED] invalid",
		},
		{
			name:     "API key in error",
			input:    "invalid api_key: 1234567890abcdef",
			expected: "invalid api_key=[REDACTED]",
		},
		{
			name:     "API key with dash",
			input:    "failed: api-key=xyz123abc456",
			expected: "failed: api-key=[REDACTED]",
		},
		{
			name:     "Secret value",
			input:    "config error: secret=topsecretvalue123",
			expected: "config error: secret=[REDACTED]",
		},
		{
			name:     "Password with quotes",
			input:    `auth failed: password="my secret" invalid`,
			expected: "auth failed: password=[REDACTED] invalid",
		},

		// Multiple sensitive items
		{
			name:     "Multiple different sensitive items",
			input:    "auth failed for user@example.com with token sk_123456 at 192.168.1.1",
			expected: "auth failed for [EMAIL] with token [TOKEN] at [IP]",
		},
		{
			name:     "Multiple paths",
			input:    "copying /home/user/file.txt to /tmp/backup.txt failed",
			expected: "copying [HOME]/file.txt to [TEMP] failed",
		},

		// Edge cases
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "No sensitive data",
			input:    "operation completed successfully",
			expected: "operation completed successfully",
		},
		{
			name:     "Path at start of message",
			input:    "/usr/bin/app failed to start",
			expected: "[PATH] failed to start",
		},
		{
			name:     "Path with spaces around it",
			input:    "error with /var/log/app/error.log and something else",
			expected: "error with [PATH] and something else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeErrorMessage() failed\ninput:    %q\nexpected: %q\ngot:      %q", tt.input, tt.expected, result)
			}
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		// APIError type
		{
			name: "APIError type",
			err: platformclient.APIError{
				Method:     "POST",
				Endpoint:   "/api/v1/test",
				StatusCode: 500,
				Message:    "Internal server error",
			},
			expected: "APIError",
		},

		// Network errors
		{
			name:     "Network error",
			err:      errors.New("network connection failed"),
			expected: "NetworkError",
		},
		{
			name:     "Connection error",
			err:      errors.New("connection refused"),
			expected: "NetworkError",
		},
		{
			name:     "Dial TCP error",
			err:      errors.New("dial tcp: connection timeout"),
			expected: "NetworkError",
		},
		{
			name:     "No such host",
			err:      errors.New("no such host api.example.com"),
			expected: "NetworkError",
		},
		{
			name:     "DNS error",
			err:      errors.New("dns lookup failed"),
			expected: "NetworkError",
		},

		// Timeout errors
		{
			name:     "Timeout error",
			err:      errors.New("operation timeout"),
			expected: "TimeoutError",
		},
		{
			name:     "Timed out",
			err:      errors.New("request timed out after 30s"),
			expected: "TimeoutError",
		},
		{
			name:     "Context deadline exceeded",
			err:      errors.New("context deadline exceeded"),
			expected: "TimeoutError",
		},
		{
			name:     "Context canceled",
			err:      errors.New("context canceled"),
			expected: "TimeoutError",
		},

		// TLS/Certificate errors
		{
			name:     "Certificate error",
			err:      errors.New("certificate has expired"),
			expected: "TLSError",
		},
		{
			name:     "TLS error",
			err:      errors.New("tls handshake failed"),
			expected: "TLSError",
		},
		{
			name:     "SSL error",
			err:      errors.New("ssl verification failed"),
			expected: "TLSError",
		},
		{
			name:     "x509 error",
			err:      errors.New("x509: certificate signed by unknown authority"),
			expected: "TLSError",
		},

		// Not found errors
		{
			name:     "Not found error",
			err:      errors.New("resource not found"),
			expected: "NotFoundError",
		},
		{
			name:     "Does not exist",
			err:      errors.New("file does not exist"),
			expected: "NotFoundError",
		},
		{
			name:     "No such file",
			err:      errors.New("no such file or directory"),
			expected: "NotFoundError",
		},

		// Permission errors
		{
			name:     "Permission denied",
			err:      errors.New("permission denied"),
			expected: "PermissionError",
		},
		{
			name:     "Forbidden",
			err:      errors.New("403 forbidden"),
			expected: "PermissionError",
		},
		{
			name:     "Unauthorized",
			err:      errors.New("401 unauthorized"),
			expected: "PermissionError",
		},
		{
			name:     "Access denied",
			err:      errors.New("access denied to resource"),
			expected: "PermissionError",
		},

		// Validation/Parse errors
		{
			name:     "Parse error",
			err:      errors.New("failed to parse JSON"),
			expected: "ValidationError",
		},
		{
			name:     "Unmarshal error",
			err:      errors.New("json unmarshal failed"),
			expected: "ValidationError",
		},
		{
			name:     "Decode error",
			err:      errors.New("failed to decode response"),
			expected: "ValidationError",
		},
		{
			name:     "Invalid syntax",
			err:      errors.New("invalid syntax in configuration"),
			expected: "ValidationError",
		},
		{
			name:     "Invalid argument",
			err:      errors.New("invalid argument provided"),
			expected: "ValidationError",
		},

		// Generic errors
		{
			name:     "Generic error",
			err:      errors.New("something went wrong"),
			expected: "Error",
		},
		{
			name:     "Unknown error type",
			err:      errors.New("unexpected situation occurred"),
			expected: "Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorType(tt.err)
			if result != tt.expected {
				t.Errorf("getErrorType() = %q, want %q for error: %q", result, tt.expected, tt.err.Error())
			}
		})
	}
}

func TestCheckConfigFileExists(t *testing.T) {
	// This test just verifies the function doesn't panic
	// In a real environment, you'd mock the filesystem or use a temp directory
	result := checkConfigFileExists()
	// Result should be boolean, no error expected
	_ = result
	// If we got here without panicking, the test passes
}

