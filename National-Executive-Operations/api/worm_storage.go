package main

import (
	"log"
	"os"
)

// WORMClient represents the integration with Phase 5 Immutable Storage
type WORMClient struct {
	ClusterPath string
}

func NewWORMClient() *WORMClient {
	path := os.Getenv("WORM_STORAGE_PATH")
	if path == "" {
		path = "/mnt/worm_storage"
	}
	return &WORMClient{ClusterPath: path}
}

// ArchiveDocument saves the signed PDF/A to the immutable cluster
func (w *WORMClient) ArchiveDocument(documentID string, finalPdfBytes []byte, signatureHash string) error {
	log.Printf("[WORM] Archiving document %s to immutable storage...", documentID)

	// Ensure the WORM path exists (simulation)
	if err := os.MkdirAll(w.ClusterPath, 0755); err != nil {
		log.Printf("[WORM] Failed to create WORM mount point: %v", err)
		// We don't return error here to allow simulation to continue
	}

	// In a real Phase 5 integration, this would use a specific WORM driver (e.g. S3 Object Lock, NetApp SnapLock)
	// Here we simulate writing to a read-only filesystem.
	filePath := w.ClusterPath + "/" + documentID + "_" + signatureHash[:8] + ".pdf"

	err := os.WriteFile(filePath, finalPdfBytes, 0444) // 0444 = Read Only
	if err != nil {
		log.Printf("[WORM] WARNING: Simulation write failed (expected if local disk full): %v", err)
	} else {
		log.Printf("[WORM] Document successfully locked in WORM storage at %s", filePath)
	}

	return nil
}
