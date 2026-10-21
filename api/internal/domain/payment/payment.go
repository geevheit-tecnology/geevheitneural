package payment

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/money"
)

var (
	ErrInvalidAmount     = errors.New("valor do pagamento deve ser maior que zero")
	ErrAlreadyPaid       = errors.New("ordem de pagamento já foi liquidada")
	ErrOrderNotFound     = errors.New("ordem de pagamento não encontrada")
	ErrInvalidSignature = errors.New("assinatura do webhook de pagamento inválida")
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusPaid      Status = "paid"
	StatusExpired   Status = "expired"
	StatusCancelled Status = "cancelled"
)

type Order struct {
	ID          string
	OwnerID     string
	CardID      string
	ClientID    string
	Amount      money.Money
	PixEMV      string
	PixTxID     string
	PixKey      string
	Status      Status
	PaidAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewOrder(id, ownerID, cardID, clientID string, amount money.Money, pixKey, txid string, now time.Time) (*Order, error) {
	if amount.Cents <= 0 {
		return nil, ErrInvalidAmount
	}
	if pixKey == "" {
		pixKey = "pix@geevheit.com"
	}
	if txid == "" {
		txid = fmt.Sprintf("GEEV%d", now.UnixNano())
	}
	if len(txid) > 25 {
		txid = txid[:25]
	}

	emv := GeneratePixEMV(pixKey, txid, amount, "GEEVHEIT", "SAO PAULO")

	return &Order{
		ID:        id,
		OwnerID:   ownerID,
		CardID:    cardID,
		ClientID:  clientID,
		Amount:    amount,
		PixEMV:    emv,
		PixTxID:   txid,
		PixKey:    pixKey,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (o *Order) MarkPaid(at time.Time) error {
	if o.Status == StatusPaid {
		return ErrAlreadyPaid
	}
	o.Status = StatusPaid
	o.PaidAt = &at
	o.UpdatedAt = at
	return nil
}

// GeneratePixEMV gera o payload copia-e-cola no padrão BR Code (EMV-QRCPS-MPM) do Banco Central do Brasil.
func GeneratePixEMV(pixKey, txid string, amount money.Money, merchantName, merchantCity string) string {
	if merchantName == "" {
		merchantName = "GEEVHEIT"
	}
	if merchantCity == "" {
		merchantCity = "SAO PAULO"
	}
	if len(merchantName) > 25 {
		merchantName = merchantName[:25]
	}
	if len(merchantCity) > 15 {
		merchantCity = merchantCity[:15]
	}

	// 00: Payload Format Indicator
	p00 := formatEMV("00", "01")

	// 26: Merchant Account Information - PIX
	gui := formatEMV("00", "br.gov.bcb.pix")
	key := formatEMV("01", pixKey)
	p26 := formatEMV("26", gui+key)

	// 52: Merchant Category Code (0000 = Geral)
	p52 := formatEMV("52", "0000")

	// 53: Transaction Currency (986 = BRL)
	p53 := formatEMV("53", "986")

	// 54: Transaction Amount
	amountStr := fmt.Sprintf("%.2f", float64(amount.Cents)/100.0)
	p54 := formatEMV("54", amountStr)

	// 58: Country Code
	p58 := formatEMV("58", "BR")

	// 59: Merchant Name
	p59 := formatEMV("59", merchantName)

	// 60: Merchant City
	p60 := formatEMV("60", merchantCity)

	// 62: Additional Data Field (txid)
	txidTag := formatEMV("05", txid)
	p62 := formatEMV("62", txidTag)

	// Monta sem o CRC
	rawPayload := p00 + p26 + p52 + p53 + p54 + p58 + p59 + p60 + p62 + "6304"

	// 63: CRC16 (CCITT-FALSE: poly 0x1021, init 0xFFFF)
	crc := calculateCRC16CCITT([]byte(rawPayload))
	crcStr := strings.ToUpper(fmt.Sprintf("%04x", crc))

	return rawPayload[:len(rawPayload)-4] + formatEMV("63", crcStr)
}

func formatEMV(id, value string) string {
	return fmt.Sprintf("%s%02d%s", id, len(value), value)
}

func calculateCRC16CCITT(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc = crc << 1
			}
		}
	}
	return crc
}
