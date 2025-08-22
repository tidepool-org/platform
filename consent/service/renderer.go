package service

import (
	"fmt"
	"io"
	"strings"

	mdtopdf "github.com/solworktech/md2pdf/v2"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/errors"
)

const (
	dateFormat = "January 2, 2006"
)

type ConsentRenderer struct {
	consent consent.Consent
	record  consent.Record
}

func NewMarkdownConsentRenderer(cons consent.Consent, record consent.Record) (*ConsentRenderer, error) {
	if cons.ContentType != consent.ContentTypeMarkdown {
		return nil, errors.Newf("cannot render %s consent content", cons.ContentType)
	}

	return &ConsentRenderer{
		consent: cons,
		record:  record,
	}, nil
}

func (m *ConsentRenderer) RenderPDF(writer io.Writer) error {
	styler := mdtopdf.Styler{
		Font: "helvetica", Style: "", Size: 12, Spacing: 2,
		TextColor: mdtopdf.Colorlookup("black"),
		FillColor: mdtopdf.Colorlookup("white"),
	}
	renderer := mdtopdf.NewPdfRendererWithDefaultStyler("", "", "", "", styler, nil, mdtopdf.LIGHT)
	if err := renderer.Pdf.Error(); err != nil {
		return errors.Wrap(err, "unable to instantiate markdown to pdf renderer")
	}

	renderer.Pdf.SetFooterFunc(m.footerRenderFunc(renderer))
	if err := renderer.Run([]byte(m.consent.Content)); err != nil {
		return err
	}

	return renderer.Pdf.Output(writer)
}

func (m *ConsentRenderer) footerRenderFunc(renderer *mdtopdf.PdfRenderer) func() {
	return func() {
		var signedBy []string
		if m.record.ParentGuardianName != nil {
			signedBy = []string{*m.record.ParentGuardianName}
		}
		signedBy = append(signedBy, m.record.OwnerName)

		// Capture current font settings
		ptSize, _ := renderer.Pdf.GetFontSize()
		fontFamily := renderer.Pdf.GetFontFamily()
		fontStyle := renderer.Pdf.GetFontStyle()

		// Reset the font family and size
		footerFontSize := float64(12)
		renderer.Pdf.SetFont("helvetica", "", footerFontSize)

		pageWidth, pageHeight := renderer.Pdf.GetPageSize()
		leftMargin, _, rightMargin, bottomMargin := renderer.Pdf.GetMargins()

		// Render horizontal line where the footer starts
		renderer.Pdf.SetDrawColor(0, 0, 0)
		renderer.Pdf.SetLineWidth(1)
		renderer.Pdf.Line(leftMargin, pageHeight-bottomMargin, pageWidth-rightMargin, pageHeight-bottomMargin)

		// Leave some space between the footer and the horizontal line
		footerHeight := bottomMargin - footerFontSize/2

		electronicSignatureWidth := pageWidth/2 - leftMargin
		electronicSignatureText := fmt.Sprintf("Electronic Signature: %s", strings.Join(signedBy, ", "))
		renderer.Pdf.SetXY(leftMargin, -footerHeight)
		renderer.Pdf.CellFormat(electronicSignatureWidth, footerHeight, electronicSignatureText, "", 0, "LT", false, 0, "")

		dateWidth := pageWidth/2 - rightMargin
		dateText := fmt.Sprintf("Date: %s", m.record.GrantTime.Format(dateFormat))
		renderer.Pdf.SetXY(leftMargin+electronicSignatureWidth, -footerHeight)
		renderer.Pdf.CellFormat(dateWidth, footerHeight, dateText, "", 0, "RT", false, 0, "")

		// Restore font settings
		renderer.Pdf.SetFont(fontFamily, fontStyle, ptSize)
	}
}
