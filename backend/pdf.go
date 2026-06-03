package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
)

func GenerateInvitationPDF(inv *Invitation, w io.Writer) error {
	// Parse layout defaults
	width := inv.Layout.WidthMm
	if width <= 0 {
		width = 210 // A4 default width
	}
	height := inv.Layout.HeightMm
	if height <= 0 {
		height = 297 // A4 default height
	}

	marginLeft := inv.Layout.MarginLeft
	if marginLeft <= 0 {
		marginLeft = 12 // default 12mm for better spacing
	}
	marginRight := inv.Layout.MarginRight
	if marginRight <= 0 {
		marginRight = 12
	}
	marginTop := inv.Layout.MarginTop
	if marginTop <= 0 {
		marginTop = 12
	}
	marginBottom := inv.Layout.MarginBottom
	if marginBottom <= 0 {
		marginBottom = 12
	}

	fontScale := inv.Layout.FontSizeScale
	if fontScale <= 0 {
		fontScale = 1.0
	}

	// Initialize Custom PDF Size
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: width, Ht: height},
	})
	
	pdf.SetMargins(marginLeft, marginTop, marginRight)
	pdf.SetAutoPageBreak(true, marginBottom)
	pdf.AddPage()

	// Unicode translator for UTF-8 support (Spanish accents like á, é, í, ó, ú, ñ)
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Colors based on theme
	var primaryColorRGB = [3]int{40, 40, 40} // Default dark charcoal for clean official look
	switch inv.Theme {
	case "coffee":
		primaryColorRGB = [3]int{111, 78, 55} // Sepia Brown
	case "blue":
		primaryColorRGB = [3]int{37, 99, 235} // Corporate Blue
	case "minimal":
		primaryColorRGB = [3]int{75, 85, 99} // Slate Gray
	}

	// Helper to scale fonts
	getFontSize := func(base float64) float64 {
		return base * fontScale
	}

	// Page boundary dimensions
	printableWidth := width - marginLeft - marginRight
	pageCenterX := width / 2.0

	// 1. Draw Outer Decorative Frame
	pdf.SetLineWidth(0.2) // Thinner lines look more modern and formal
	pdf.SetHeaderFunc(nil) // disable default headers
	pdf.SetDrawColor(primaryColorRGB[0], primaryColorRGB[1], primaryColorRGB[2])
	pdf.Rect(marginLeft, marginTop, printableWidth, height-marginTop-marginBottom, "D")

	// 2. Draw Document Header Grid (Spacious: headerHeight is 48.0mm)
	headerHeight := 48.0
	pdf.Line(marginLeft, marginTop+headerHeight, marginLeft+printableWidth, marginTop+headerHeight) // horizontal separator
	pdf.Line(pageCenterX, marginTop, pageCenterX, marginTop+headerHeight)                           // vertical division

	// Middle Letter Box "R" or "OC"
	letterBoxW := 18.0
	letterBoxH := 18.0
	letterBoxX := pageCenterX - (letterBoxW / 2)
	letterBoxY := marginTop

	// Draw the box for the document type letter
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(letterBoxX, letterBoxY, letterBoxW, letterBoxH, "FD")

	// Letter content
	pdf.SetTextColor(primaryColorRGB[0], primaryColorRGB[1], primaryColorRGB[2])
	pdf.SetFont("Helvetica", "B", getFontSize(22)) // Clean typography
	letter := "R"
	subLetter := "REMITO"
	if inv.Type == "orden" {
		letter = "OC"
		subLetter = "ORDEN COMPRA"
	}
	
	pdf.Text(letterBoxX + (letterBoxW-pdf.GetStringWidth(letter))/2, letterBoxY + 11.5, letter)

	// Draw subLetter inside the letter box
	pdf.SetFont("Helvetica", "B", getFontSize(4))
	pdf.Text(letterBoxX + (letterBoxW-pdf.GetStringWidth(subLetter))/2, letterBoxY + 15.5, subLetter)

	// Document non-valid notice under the letter box
	pdf.SetFont("Helvetica", "B", getFontSize(5.5))
	pdf.SetTextColor(80, 80, 80)
	noticeText := "DOCUMENTO NO VÁLIDO COMO FACTURA"
	pdf.Text(pageCenterX - pdf.GetStringWidth(noticeText)/2, letterBoxY + letterBoxH + 4, tr(noticeText))

	// Sub-type text (e.g. Original / Triplicado)
	pdf.SetFont("Helvetica", "", getFontSize(7))
	statusNotice := "ORIGINAL"
	if inv.Status == "accepted" {
		statusNotice = "DUPLICADO (RECEPCIÓN FIRMADA)"
	}
	pdf.Text(pageCenterX - pdf.GetStringWidth(statusNotice)/2, letterBoxY + letterBoxH + 8, tr(statusNotice))

	// 3. Left Header Block: Emisor Info & Logo
	// Logo & Company name placement
	textShiftX := 5.0
	if inv.Emisor.Logo != "" {
		tmpPath, err := writeBase64Image(inv.Emisor.Logo)
		if err == nil {
			defer os.Remove(tmpPath)
			// Place logo at top left of emisor header
			pdf.ImageOptions(tmpPath, marginLeft + 5, marginTop + 4, 13, 13, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
			textShiftX = 20.0
		}
	}

	pdf.SetTextColor(primaryColorRGB[0], primaryColorRGB[1], primaryColorRGB[2])
	pdf.SetFont("Helvetica", "B", getFontSize(12.5))
	emisorName := inv.Emisor.Company
	if emisorName == "" {
		emisorName = inv.Emisor.Name
	}
	if len(emisorName) > 22 {
		emisorName = emisorName[:22]
	}
	pdf.Text(marginLeft + textShiftX, marginTop + 8, tr(emisorName))

	pdf.SetFont("Helvetica", "", getFontSize(8))
	pdf.SetTextColor(60, 60, 60)
	pdf.Text(marginLeft + 5, marginTop + 17, tr(fmt.Sprintf("De: %s", inv.Emisor.Name)))
	pdf.Text(marginLeft + 5, marginTop + 22, tr(fmt.Sprintf("Dirección: %s", inv.Emisor.Address)))
	pdf.Text(marginLeft + 5, marginTop + 27, tr(fmt.Sprintf("Email: %s", inv.Emisor.Email)))
	
	cuitEmisor := inv.Emisor.Cuit
	if cuitEmisor == "" {
		cuitEmisor = "30-71452968-3"
	}
	pdf.Text(marginLeft + 5, marginTop + 32, tr(fmt.Sprintf("CUIT: %s", cuitEmisor)))
	pdf.Text(marginLeft + 5, marginTop + 37, tr("I.V.A.: Responsable Inscripto"))

	// 4. Right Header Block: Document metadata
	pdf.SetTextColor(primaryColorRGB[0], primaryColorRGB[1], primaryColorRGB[2])
	pdf.SetFont("Helvetica", "B", getFontSize(11.5))
	docTitle := "REMITO DE VENTA"
	if inv.Type == "orden" {
		docTitle = "ORDEN DE COMPRA"
	}
	pdf.Text(pageCenterX + 8, marginTop + 8, tr(docTitle))

	// Document Number details
	docNum := inv.DocNumber
	if docNum == "" {
		docNum = "0001-00000042"
	}
	pdf.SetFont("Helvetica", "B", getFontSize(10.5))
	pdf.Text(pageCenterX + 8, marginTop + 14, tr(fmt.Sprintf("N° %s", docNum)))

	// Date details
	pdf.SetFont("Helvetica", "", getFontSize(8))
	pdf.SetTextColor(60, 60, 60)
	pdf.Text(pageCenterX + 8, marginTop + 20, tr(fmt.Sprintf("Fecha Emisión: %s", inv.CreatedAt.Format("02/01/2006"))))
	pdf.Text(pageCenterX + 8, marginTop + 25, tr(fmt.Sprintf("Fecha Evento: %s", inv.EventDate.Format("02/01/2006 15:04"))))
	pdf.Text(pageCenterX + 8, marginTop + 30, tr(fmt.Sprintf("CUIT Emisor: %s", cuitEmisor)))
	pdf.Text(pageCenterX + 8, marginTop + 35, tr("Ing. Brutos: 30-71452968-3 (Exento)"))

	// 5. Receptor (Client) Block (Spacious)
	receptorY := marginTop + headerHeight + 5.0
	receptorH := 30.0
	pdf.SetLineWidth(0.2)
	pdf.Rect(marginLeft + 3, receptorY, printableWidth - 6, receptorH, "D")

	pdf.SetFont("Helvetica", "B", getFontSize(8))
	pdf.SetTextColor(primaryColorRGB[0], primaryColorRGB[1], primaryColorRGB[2])
	pdf.Text(marginLeft + 6, receptorY + 5, tr("RECEPTOR / DESTINATARIO:"))

	pdf.SetFont("Helvetica", "", getFontSize(8))
	pdf.SetTextColor(50, 50, 50)
	
	receptorCompany := inv.Receptor.Company
	if receptorCompany == "" {
		receptorCompany = "(Particular / Personal)"
	}
	pdf.Text(marginLeft + 6, receptorY + 11, tr(fmt.Sprintf("Señor/es: %s (%s)", inv.Receptor.Name, receptorCompany)))
	pdf.Text(marginLeft + 6, receptorY + 16, tr(fmt.Sprintf("Domicilio Comercial: %s", inv.Receptor.Address)))
	pdf.Text(marginLeft + 6, receptorY + 21, tr(fmt.Sprintf("Email: %s", inv.Receptor.Email)))
	
	cuitReceptor := inv.Receptor.Cuit
	if cuitReceptor == "" {
		cuitReceptor = "Consumidor Final"
	}
	pdf.Text(marginLeft + 6, receptorY + 26, tr(fmt.Sprintf("CUIT: %s", cuitReceptor)))
	pdf.Text(pageCenterX + 10, receptorY + 26, tr("Condición IVA: Consumidor Final"))

	// 5.5 Commercial Conditions Block (Very Realistic - Official Spacing)
	condY := receptorY + receptorH + 4.0
	condH := 8.0
	pdf.Rect(marginLeft + 3, condY, printableWidth - 6, condH, "D")
	pdf.SetFont("Helvetica", "", getFontSize(7.5))
	pdf.Text(marginLeft + 6, condY + 5.0, tr("Condición de Venta: [ ] Contado    [ ] Cta. Cte.    [X] Invitación Especial"))
	pdf.Text(pageCenterX + 10, condY + 5.0, tr("Transporte / Logística: [X] Sin Cargo    [ ] Retira Destinatario"))

	// 6. Watermark stamp (Diagonal back)
	pdf.SetFont("Helvetica", "B", getFontSize(32))
	pdf.SetTextColor(240, 240, 240)
	
	pdf.SetAlpha(0.12, "Normal")
	var stampText string
	if inv.Status == "accepted" {
		pdf.SetTextColor(40, 167, 69) // green stamp
		stampText = "RECIBIDO - ASISTE"
	} else if inv.Status == "declined" {
		pdf.SetTextColor(220, 53, 69) // red stamp
		stampText = "RECHAZADO - NO ASISTE"
	} else {
		pdf.SetTextColor(108, 117, 125) // grey stamp
		stampText = "PENDIENTE DE FIRMA"
	}
	
	pdf.Text(pageCenterX - (pdf.GetStringWidth(tr(stampText)) / 2), height * 0.6, tr(stampText))
	pdf.SetAlpha(1.0, "Normal") // reset alpha

	// 7. Items Table (Breathing room added)
	tableY := condY + condH + 5.0
	tableH := height - tableY - marginBottom - 45.0 // dynamic table height
	
	// Draw table container
	pdf.SetTextColor(primaryColorRGB[0], primaryColorRGB[1], primaryColorRGB[2])
	pdf.Rect(marginLeft + 3, tableY, printableWidth - 6, tableH, "D")

	// Table column widths based on layout percentages
	qPct := inv.Layout.ColQtyWidth
	dPct := inv.Layout.ColDescWidth
	pPct := inv.Layout.ColPriceWidth
	tPct := inv.Layout.ColTotalWidth
	
	if math.Abs((qPct + dPct + pPct + tPct) - 100) > 1.0 {
		qPct = 15.0
		dPct = 50.0
		pPct = 18.0
		tPct = 17.0
	}

	tableW := printableWidth - 6
	wQty := tableW * qPct / 100
	wDesc := tableW * dPct / 100
	wPrice := tableW * pPct / 100
	wTotal := tableW * tPct / 100

	// Draw table headers (Spacious cell heights)
	pdf.SetLineWidth(0.2)
	pdf.SetFillColor(246, 248, 250)
	
	pdf.SetXY(marginLeft + 3, tableY)
	pdf.SetFont("Helvetica", "B", getFontSize(8))
	pdf.SetTextColor(primaryColorRGB[0], primaryColorRGB[1], primaryColorRGB[2])
	
	pdf.CellFormat(wQty, 7.5, tr("Cant."), "1", 0, "C", true, 0, "")
	pdf.CellFormat(wDesc, 7.5, tr("Detalle / Descripción del Concepto"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(wPrice, 7.5, tr("P. Unitario"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(wTotal, 7.5, tr("Importe"), "1", 1, "R", true, 0, "") // newline after header

	// Draw vertical column grid lines inside table
	colX1 := marginLeft + 3 + wQty
	colX2 := colX1 + wDesc
	colX3 := colX2 + wPrice
	pdf.Line(colX1, tableY, colX1, tableY + tableH)
	pdf.Line(colX2, tableY, colX2, tableY + tableH)
	pdf.Line(colX3, tableY, colX3, tableY + tableH)

	// Draw items (Row heights increased to 7.0mm for padding/breathing room)
	pdf.SetFont("Helvetica", "", getFontSize(8))
	pdf.SetTextColor(60, 60, 60)
	
	currY := tableY + 7.5
	totalAmount := 0.0

	for i, item := range inv.Items {
		if currY + 7.0 > tableY + tableH - 7.5 {
			// Prevent page overflow inside the same table
			pdf.SetXY(marginLeft + 3, tableY + tableH - 7.5)
			pdf.SetFont("Helvetica", "I", getFontSize(7.5))
			pdf.CellFormat(tableW, 7.5, tr("--- Continúa en hoja siguiente ---"), "T", 1, "C", false, 0, "")
			break
		}

		pdf.SetXY(marginLeft + 3, currY)
		
		// Quantity format
		qtyStr := fmt.Sprintf("%.1f", item.Qty)
		if item.Qty == float64(int(item.Qty)) {
			qtyStr = fmt.Sprintf("%d", int(item.Qty))
		}
		
		// Prices
		priceStr := "$0.00"
		if item.UnitPrice > 0 {
			priceStr = fmt.Sprintf("$%.2f", item.UnitPrice)
		} else {
			priceStr = "GRATIS"
		}
		
		totalItem := item.Qty * item.UnitPrice
		totalStr := "$0.00"
		if totalItem > 0 {
			totalStr = fmt.Sprintf("$%.2f", totalItem)
		} else {
			totalStr = "INVALUABLE"
		}
		
		totalAmount += totalItem

		// Description wrap
		desc := item.Description
		if len(desc) > 55 {
			desc = desc[:52] + "..."
		}

		// Light gray row backgrounds for alternating items
		rowFill := false
		if i%2 == 1 {
			rowFill = true
			pdf.SetFillColor(250, 252, 255)
		}

		pdf.CellFormat(wQty, 7.0, tr(qtyStr), "", 0, "C", rowFill, 0, "")
		pdf.CellFormat(wDesc, 7.0, "  " + tr(desc), "", 0, "L", rowFill, 0, "")
		pdf.CellFormat(wPrice, 7.0, tr(priceStr) + "  ", "", 0, "R", rowFill, 0, "")
		pdf.CellFormat(wTotal, 7.0, tr(totalStr) + "  ", "", 1, "R", rowFill, 0, "")
		
		currY += 7.0
	}

	// 8. Draw Total Row at the bottom of the table
	pdf.SetXY(marginLeft + 3, tableY + tableH - 7.5)
	pdf.SetFont("Helvetica", "B", getFontSize(8))
	pdf.SetTextColor(primaryColorRGB[0], primaryColorRGB[1], primaryColorRGB[2])
	pdf.SetFillColor(240, 242, 245)
	
	totalText := "TOTAL DE LA INVITACIÓN"
	priceTotalText := "SONRISAS / $0.00"
	if totalAmount > 0 {
		priceTotalText = fmt.Sprintf("$%.2f", totalAmount)
	}

	pdf.CellFormat(wQty + wDesc, 7.5, "  " + tr(totalText), "T", 0, "L", true, 0, "")
	pdf.CellFormat(wPrice, 7.5, tr("Importe Total: "), "T", 0, "R", true, 0, "")
	pdf.CellFormat(wTotal, 7.5, tr(priceTotalText) + "  ", "T", 1, "R", true, 0, "")

	// 9. Footer: Comments & Signatures
	footerY := tableY + tableH + 4.0
	footerH := height - footerY - marginBottom - 3.5

	// Draw comments box
	commentsBoxW := tableW * 0.55
	commentsBoxH := footerH - 4.0
	pdf.SetLineWidth(0.2)
	pdf.Rect(marginLeft + 3, footerY, commentsBoxW, commentsBoxH, "D")
	
	pdf.SetFont("Helvetica", "B", getFontSize(7.5))
	pdf.Text(marginLeft + 5, footerY + 4, tr("NOTAS / CONDICIONES GENERALES:"))

	pdf.SetFont("Helvetica", "", getFontSize(7.5))
	pdf.SetTextColor(80, 80, 80)
	
	// Multi-line comments writing
	comments := inv.Comments
	if comments == "" {
		comments = "Al firmar este remito, el receptor se compromete formalmente a asistir al evento detallado en la descripción y a traer excelente actitud. El emisor no se responsabiliza por sobredosis de medialunas."
	}
	
	// Wrap text in comments box
	commentLines := pdf.SplitText(comments, commentsBoxW - 4.0)
	lineY := footerY + 8.5
	for idx, line := range commentLines {
		if idx >= 3 { // max 3 lines to make room for CAI/Transport
			break
		}
		pdf.Text(marginLeft + 5, lineY, tr(line))
		lineY += 3.5
	}

	// Draw a dotted separator line at the bottom of the comments box
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.1)
	pdf.Line(marginLeft + 5.0, footerY + commentsBoxH - 7.5, marginLeft + commentsBoxW - 2.0, footerY + commentsBoxH - 7.5)
	
	// Reset colors and write transport/CAI text
	pdf.SetFont("Helvetica", "", getFontSize(5.5))
	pdf.SetTextColor(100, 100, 100)
	pdf.Text(marginLeft + 5.0, footerY + commentsBoxH - 4.5, tr("Flete: POTENCIAPP Logística S.A.  Patente: AAA-999-PP"))
	caiText := "CAI N°: 42026159846271  Vto: 31/12/2026"
	pdf.Text(marginLeft + commentsBoxW - pdf.GetStringWidth(caiText) - 2.0, footerY + commentsBoxH - 4.5, tr(caiText))

	// Signatures / Receipt stamp block (Clean and spacious)
	sigBoxW := tableW - commentsBoxW - 2.0
	sigBoxX := marginLeft + 3 + commentsBoxW + 2.0
	sigBoxH := footerH - 4.0
	pdf.SetDrawColor(primaryColorRGB[0], primaryColorRGB[1], primaryColorRGB[2])
	pdf.SetLineWidth(0.2)
	pdf.Rect(sigBoxX, footerY, sigBoxW, sigBoxH, "D")

	pdf.SetFont("Helvetica", "B", getFontSize(7))
	pdf.Text(sigBoxX + 2, footerY + 4, tr("CONFORME RECEPTOR:"))

	// Signature display
	if inv.Status == "accepted" {
		if inv.Signature != "" {
			tmpPath, err := writeBase64Image(inv.Signature)
			if err == nil {
				defer os.Remove(tmpPath) // delete after rendering
				imgW := sigBoxW - 6.0
				imgH := sigBoxH - 12.0
				imgX := sigBoxX + 3.0
				imgY := footerY + 5.0
				pdf.ImageOptions(tmpPath, imgX, imgY, imgW, imgH, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
			} else {
				pdf.SetFont("Courier", "BI", getFontSize(10))
				pdf.SetTextColor(0, 100, 0)
				pdf.Text(sigBoxX + 5, footerY + (sigBoxH/2), tr("/ FIRMADO CONFORME /"))
			}
		} else {
			pdf.SetFont("Courier", "BI", getFontSize(9))
			pdf.SetTextColor(0, 100, 0)
			pdf.Text(sigBoxX + 5, footerY + (sigBoxH/2), tr("ACEPTADO - SIN FIRMA DIBUJADA"))
		}
		
		// Signee details
		pdf.SetFont("Helvetica", "", getFontSize(6.5))
		pdf.SetTextColor(60, 60, 60)
		pdf.Text(sigBoxX + 3, footerY + sigBoxH - 5, tr(fmt.Sprintf("Aclaración: %s", inv.Receptor.Name)))
		pdf.Text(sigBoxX + 3, footerY + sigBoxH - 2, tr(fmt.Sprintf("Recibido el: %s", time.Now().Format("02/01/2006 15:04"))))
	} else if inv.Status == "declined" {
		pdf.SetFont("Helvetica", "B", getFontSize(8))
		pdf.SetTextColor(220, 53, 69)
		pdf.Text(sigBoxX + 3, footerY + (sigBoxH / 2) - 2, tr("X RECHAZADO / NO ASISTE"))
		
		reason := inv.RejectionReason
		if reason == "" {
			reason = "No especificado"
		}
		if len(reason) > 28 {
			reason = reason[:25] + "..."
		}
		pdf.SetFont("Helvetica", "I", getFontSize(6.5))
		pdf.SetTextColor(80, 80, 80)
		pdf.Text(sigBoxX + 3, footerY + (sigBoxH / 2) + 2, tr(fmt.Sprintf("Motivo: %s", reason)))
	} else {
		// Draw dotted lines for manual signature
		pdf.SetLineWidth(0.1)
		pdf.SetDrawColor(180, 180, 180)
		pdf.Line(sigBoxX + 4, footerY + sigBoxH - 8, sigBoxX + sigBoxW - 4, footerY + sigBoxH - 8)
		
		pdf.SetFont("Helvetica", "", getFontSize(6.5))
		pdf.SetTextColor(120, 120, 120)
		pdf.Text(sigBoxX + (sigBoxW-pdf.GetStringWidth(tr("Firma y Aclaración")))/2, footerY + sigBoxH - 4, tr("Firma y Aclaración"))
	}

	// 10. Draw Mock Barcode at the very bottom
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.2)
	
	barcodeY := height - marginBottom - 3.5
	barcodeX := marginLeft + 5.0
	barcodeW := 35.0
	barcodeH := 2.5
	
	// Create barcode lines
	for bx := 0.0; bx < barcodeW; bx += 0.8 {
		if int(bx*10)%3 == 0 {
			pdf.Line(barcodeX + bx, barcodeY, barcodeX + bx, barcodeY + barcodeH)
		}
		if int(bx*10)%7 == 0 {
			pdf.Line(barcodeX + bx + 0.2, barcodeY, barcodeX + bx + 0.2, barcodeY + barcodeH)
		}
	}
	
	// Barcode label
	pdf.SetFont("Helvetica", "", getFontSize(5))
	pdf.SetTextColor(100, 100, 100)
	pdf.Text(barcodeX + barcodeW + 2, barcodeY + 2, tr(fmt.Sprintf("C.O.T. N° 42026%s%s", inv.CreatedAt.Format("0102"), inv.DocNumber)))

	// Renders buffer to writer
	return pdf.Output(w)
}

// writeBase64Image Decodes the base64 png string and writes it to a temporary file, returning its path
func writeBase64Image(b64Str string) (string, error) {
	// Strip headers like "data:image/png;base64,"
	if idx := strings.Index(b64Str, ","); idx != -1 {
		b64Str = b64Str[idx+1:]
	}

	dec, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return "", err
	}

	// Validate the bytes are a decodable PNG before handing them to the PDF
	// engine. An invalid image would otherwise set an internal error on the
	// PDF object and make Output() write a 0-byte file. Returning an error here
	// makes the caller fall back to the textual "/ FIRMADO CONFORME /" stamp.
	if _, err := png.Decode(bytes.NewReader(dec)); err != nil {
		return "", err
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "sig-*.png")
	if err != nil {
		return "", err
	}
	
	defer tmpFile.Close()

	if _, err := tmpFile.Write(dec); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
