package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	outputDir := "../documents/generated_decrees"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	decrees := []struct {
		Title      string
		ID         string
		Department string
		Content    string
	}{
		{
			Title:      "Décret d'Interopérabilité Nationale",
			ID:         "DEC-PHASE4-001",
			Department: "Primature",
			Content:    "Article 1: Il est formellement interdit à toute entité étatique de créer une base de données d'identité isolée.\nArticle 2: L'usage de la National API Gateway est rendu obligatoire.",
		},
		{
			Title:      "Décret Zero Trust & Cyber-Défense",
			ID:         "DEC-PHASE6-001",
			Department: "Présidence de la République",
			Content:    "Article 1: Le CISO National est doté du pouvoir absolu d'isoler réseau (Drop Traffic) tout ministère présentant une faille critique non patchée.",
		},
		{
			Title:      "Décret d'Ouverture Bancaire (Monétisation)",
			ID:         "DEC-PHASE10-001",
			Department: "Ministère de l'Économie et des Finances",
			Content:    "Article 1: Le service national d'identité (KYC) est ouvert aux banques privées sous contrat financier.\nArticle 2: Les quotas d'API doivent être strictement respectés.",
		},
	}

	for _, d := range decrees {
		// Simulation de l'utilisation de "l'Approval Pack" de la Phase 13
		htmlContent := fmt.Sprintf(`
		<html>
		<head><style>body{font-family:serif; text-align:center; padding: 50px;}</style></head>
		<body>
			<div style="border: 4px solid double #000; padding: 40px;">
				<h1>RÉPUBLIQUE D'HAÏTI</h1>
				<h2>%s</h2>
				<h3>%s</h3>
				<hr>
				<p style="text-align:left; white-space: pre-wrap;">%s</p>
				<br><br>
				<p><strong>Tracking ID :</strong> %s</p>
				<p><strong>Date :</strong> %s</p>
				<br><hr><br>
				<p><em>Prêt pour signature dans le Parapheur Électronique (Phase 13)</em></p>
			</div>
		</body>
		</html>
		`, d.Department, d.Title, d.Content, d.ID, time.Now().Format("2006-01-02"))

		filename := filepath.Join(outputDir, d.ID+".html")
		if err := os.WriteFile(filename, []byte(htmlContent), 0644); err != nil {
			log.Printf("Erreur lors de la génération de %s: %v", d.ID, err)
		} else {
			log.Printf("Généré avec succès : %s", filename)
		}
	}

	log.Println("Tous les décrets de la Phase 14 ont été générés et sont prêts à être intégrés au Parapheur.")
}
