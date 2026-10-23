package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalidURL       = errors.New("url do webhook inválida")
	ErrSSRFBlocked      = errors.New("endereço de destino bloqueado por proteção contra SSRF (IP privado, loopback ou metadata)")
	ErrHTTPSRequired    = errors.New("webhooks em produção exigem HTTPS")
	ErrEndpointInactive = errors.New("endpoint de webhook está inativo")
)

type Endpoint struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	SecretToken string    `json:"secret_token,omitempty"` // Apenas exibido na criação
	Events      []string  `json:"events"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Delivery struct {
	ID           string         `json:"id"`
	OwnerID      string         `json:"owner_id"`
	WebhookID    string         `json:"webhook_id"`
	Event        string         `json:"event"`
	Payload      map[string]any `json:"payload"`
	StatusCode   int            `json:"status_code"`
	Attempts     int            `json:"attempts"`
	ResponseBody string         `json:"response_body"`
	Success      bool           `json:"success"`
	CreatedAt    time.Time      `json:"created_at"`
}

// ComputeHMACSignature calcula a assinatura HMAC-SHA256 para o payload com timestamp para evitar replay attacks
func ComputeHMACSignature(secret string, timestamp int64, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	message := fmt.Sprintf("%d.%s", timestamp, string(payload))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateTargetURL inspeciona o destino e bloqueia acessos a redes internas ou metadata cloud (SSRF)
func ValidateTargetURL(rawURL string, requireHTTPS bool) error {
	if rawURL == "" {
		return ErrInvalidURL
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ErrInvalidURL
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("%w: esquema não suportado", ErrInvalidURL)
	}

	if requireHTTPS && scheme != "https" {
		return ErrHTTPSRequired
	}

	hostname := u.Hostname()
	if hostname == "" {
		return ErrInvalidURL
	}

	// Bloqueia hostname localhost expressamente
	if strings.EqualFold(hostname, "localhost") {
		return ErrSSRFBlocked
	}

	// Bloqueia se resolver para IP privado, loopback ou link-local/cloud metadata
	ips, err := net.LookupIP(hostname)
	if err != nil {
		// Se for IP direto
		ip := net.ParseIP(hostname)
		if ip != nil {
			if isPrivateOrReservedIP(ip) {
				return ErrSSRFBlocked
			}
			return nil
		}
		// Se não conseguir resolver DNS, permite validação em tempo de disparo
		return nil
	}

	for _, ip := range ips {
		if isPrivateOrReservedIP(ip) {
			return ErrSSRFBlocked
		}
	}

	return nil
}

func isPrivateOrReservedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}

	// Range privado 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 {
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		// Cloud metadata AWS/GCP/Azure 169.254.169.254
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	}

	return false
}

// MatchesEvent verifica se o endpoint assina o evento informado (ou wildcard "*")
func (e *Endpoint) MatchesEvent(event string) bool {
	if !e.Active {
		return false
	}
	for _, subscribed := range e.Events {
		if subscribed == "*" || subscribed == event {
			return true
		}
	}
	return false
}
