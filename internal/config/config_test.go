package config

import "testing"

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "https", value: "https://example.com/send", wantErr: false},
		{name: "localhost", value: "http://localhost:8080/send", wantErr: false},
		{name: "ipv4 loopback", value: "http://127.0.0.1:8080/send", wantErr: false},
		{name: "ipv6 loopback", value: "http://[::1]:8080/send", wantErr: false},
		{name: "public http", value: "http://example.com/send", wantErr: true},
		{name: "missing scheme", value: "example.com/send", wantErr: true},
		{name: "missing host", value: "https:///send", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateWebhookURL(test.value)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateWebhookURL(%q) error = %v, wantErr = %v", test.value, err, test.wantErr)
			}
		})
	}
}
