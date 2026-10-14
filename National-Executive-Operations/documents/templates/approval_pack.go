package templates

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

// DocumentData represents the variables passed to the templates
type DocumentData struct {
	Title         string
	Department    string
	TrackingID    string
	Signatures    []SignatureData
	GeneratedDate string
}

type SignatureData struct {
	Role      string
	SignerID  string
	Timestamp string
	Hash      string
}

// GenerateOfficialCoverPage generates the HTML template for the cover page
func GenerateOfficialCoverPage(data DocumentData) (string, error) {
	tmpl := `
	<html>
	<head><style>body{font-family:serif; text-align:center; padding: 50px;}</style></head>
	<body>
		<div style="border: 4px solid double #000; padding: 40px;">
			<h1>RÉPUBLIQUE D'HAÏTI</h1>
			<hr style="width: 50%; margin-bottom: 30px;">
			<h2>{{.Department}}</h2>
			<br><br>
			<h3>{{.Title}}</h3>
			<br><br>
			<p><strong>Tracking ID :</strong> {{.TrackingID}}</p>
			<p><strong>Date de génération :</strong> {{.GeneratedDate}}</p>
		</div>
	</body>
	</html>
	`
	return render(tmpl, data)
}

// GenerateVisaCircuitTemplate generates the HTML for the validation tracking sheet
func GenerateVisaCircuitTemplate(data DocumentData) (string, error) {
	tmpl := `
	<html>
	<head><style>table {width: 100%; border-collapse: collapse;} th, td {border: 1px solid black; padding: 10px; text-align: left;}</style></head>
	<body>
		<h2>Feuille de Suivi des Visas (Visa Circuit)</h2>
		<p>Document: {{.Title}} ({{.TrackingID}})</p>
		<table>
			<tr>
				<th>Rôle / Ministère</th>
				<th>Signataire</th>
				<th>Horodatage</th>
				<th>Signature Cryptographique (SHA-256 / QES)</th>
			</tr>
			{{range .Signatures}}
			<tr>
				<td>{{.Role}}</td>
				<td>{{.SignerID}}</td>
				<td>{{.Timestamp}}</td>
				<td style="font-family: monospace; font-size: 0.8em;">{{.Hash}}</td>
			</tr>
			{{end}}
		</table>
	</body>
	</html>
	`
	return render(tmpl, data)
}

func render(tmplStr string, data DocumentData) (string, error) {
	t, err := template.New("doc").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// PackDocuments generate a full HTML pack ready for PDF conversion
func PackDocuments(title, department, id string) string {
	data := DocumentData{
		Title:         title,
		Department:    department,
		TrackingID:    id,
		GeneratedDate: time.Now().Format(time.RFC3339),
		Signatures: []SignatureData{
			{Role: "Ministre de la Santé", SignerID: "MIN-001", Timestamp: "EN ATTENTE", Hash: "-"},
		},
	}

	cover, _ := GenerateOfficialCoverPage(data)
	visa, _ := GenerateVisaCircuitTemplate(data)

	// En production, un moteur comme wkhtmltopdf ou puppeteer convertit ces chaînes en PDF/A.
	return fmt.Sprintf("<!-- COVER PAGE -->\n%s\n<!-- VISA CIRCUIT -->\n%s", cover, visa)
}
