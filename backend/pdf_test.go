package main

import (
	"bytes"
	"testing"
	"time"
)

func TestGenerateInvitationPDF(t *testing.T) {
	mockInv := &Invitation{
		ID:        "test-uuid-1234",
		Type:      "remito",
		DocNumber: "0001-00000042",
		Emisor: CompanyInfo{
			Name:    "Test Sender",
			Company: "Test Emisor S.A.",
			Cuit:    "30-11111111-3",
			Address: "123 Sender St",
			Email:   "sender@test.com",
		},
		Receptor: CompanyInfo{
			Name:    "Test Recipient",
			Company: "Test Receptor S.A.",
			Cuit:    "30-22222222-4",
			Address: "456 Recipient St",
			Email:   "recipient@test.com",
		},
		EventDate: time.Now().Add(24 * time.Hour),
		Items: []Item{
			{Qty: 1, Description: "Café Caliente", UnitPrice: 0, Total: 0},
			{Qty: 2, Description: "Facturas Dulce de Leche", UnitPrice: 10.5, Total: 21.0},
		},
		Comments: "Test comments",
		Theme:    "coffee",
		Status:   "accepted",
		Signature: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==", // 1x1 white pixel PNG
		Layout: LayoutConfig{
			PageSize:      "A4",
			WidthMm:       210,
			HeightMm:      297,
			MarginTop:     10,
			MarginBottom:  10,
			MarginLeft:    10,
			MarginRight:   10,
			ColQtyWidth:   15,
			ColDescWidth:  50,
			ColPriceWidth: 18,
			ColTotalWidth: 17,
			FontSizeScale: 1.0,
		},
		CreatedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := GenerateInvitationPDF(mockInv, &buf)
	if err != nil {
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Generated PDF is empty (0 bytes written)")
	}

	t.Logf("PDF generated successfully! Size: %d bytes", buf.Len())
}
