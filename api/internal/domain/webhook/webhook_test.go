package webhook_test

import (
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/domain/webhook"
)

func TestComputeHMACSignature(t *testing.T) {
	secret := "test_secret_key_123"
	timestamp := int64(1761184800)
	payload := []byte(`{"event":"payment.pix_confirmed","amount_cents":50000}`)

	sig1 := webhook.ComputeHMACSignature(secret, timestamp, payload)
	if sig1 == "" {
		t.Fatal("Assinatura não deveria ser vazia")
	}

	// Mesmos inputs geram a mesma assinatura
	sig2 := webhook.ComputeHMACSignature(secret, timestamp, payload)
	if sig1 != sig2 {
		t.Errorf("Assinatura idempotente falhou: %s != %s", sig1, sig2)
	}

	// Payload alterado altera a assinatura
	tamperedPayload := []byte(`{"event":"payment.pix_confirmed","amount_cents":99999}`)
	sigTampered := webhook.ComputeHMACSignature(secret, timestamp, tamperedPayload)
	if sig1 == sigTampered {
		t.Error("Assinatura com payload alterado não deveria coincidir")
	}

	// Timestamp diferente altera a assinatura (anti-replay)
	sigDiffTimestamp := webhook.ComputeHMACSignature(secret, timestamp+1, payload)
	if sig1 == sigDiffTimestamp {
		t.Error("Assinatura com timestamp diferente não deveria coincidir")
	}
}

func TestValidateTargetURL_SSRFProtection(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		requireHTTPS bool
		expectErr    bool
	}{
		{
			name:         "URL HTTPS pública válida",
			url:          "https://webhook.site/abc-123",
			requireHTTPS: true,
			expectErr:    false,
		},
		{
			name:         "URL HTTP pública rejeitada se exigir HTTPS",
			url:          "http://webhook.site/abc-123",
			requireHTTPS: true,
			expectErr:    true,
		},
		{
			name:         "Localhost bloqueado por SSRF",
			url:          "http://localhost:8080/hook",
			requireHTTPS: false,
			expectErr:    true,
		},
		{
			name:         "IP Loopback 127.0.0.1 bloqueado",
			url:          "http://127.0.0.1:3000/hook",
			requireHTTPS: false,
			expectErr:    true,
		},
		{
			name:         "IP Privado 192.168.1.1 bloqueado",
			url:          "http://192.168.1.1:8000/hook",
			requireHTTPS: false,
			expectErr:    true,
		},
		{
			name:         "IP Privado 10.0.0.5 bloqueado",
			url:          "http://10.0.0.5/hook",
			requireHTTPS: false,
			expectErr:    true,
		},
		{
			name:         "IP Cloud Metadata 169.254.169.254 bloqueado",
			url:          "http://169.254.169.254/latest/meta-data/",
			requireHTTPS: false,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := webhook.ValidateTargetURL(tt.url, tt.requireHTTPS)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateTargetURL(%q) err = %v, expectErr %v", tt.url, err, tt.expectErr)
			}
		})
	}
}

func TestEndpoint_MatchesEvent(t *testing.T) {
	endpoint := webhook.Endpoint{
		Active: true,
		Events: []string{"intake.received", "payment.pix_confirmed"},
	}

	if !endpoint.MatchesEvent("intake.received") {
		t.Error("Deveria casar com intake.received")
	}
	if endpoint.MatchesEvent("portal.accepted") {
		t.Error("Não deveria casar com portal.accepted")
	}

	wildcardEndpoint := webhook.Endpoint{
		Active: true,
		Events: []string{"*"},
	}
	if !wildcardEndpoint.MatchesEvent("qualquer_coisa") {
		t.Error("Wildcard deveria casar com qualquer evento")
	}

	inactiveEndpoint := webhook.Endpoint{
		Active: false,
		Events: []string{"*"},
	}
	if inactiveEndpoint.MatchesEvent("qualquer_coisa") {
		t.Error("Endpoint inativo não deve casar com nenhum evento")
	}
}
