package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"

	"resumetake/models"
)

func ValidatePDFHeader(data []byte) bool {
	if len(data) < 5 {
		return false
	}
	return string(data[:5]) == "%PDF-"
}

func ExtractPDFText(pdfData []byte) (text string, pageCount int, err error) {
	// R51-B3: recover from panics in the third-party pdf library
	// (github.com/ledongthuc/pdf). Malicious PDFs can trigger nil pointer
	// dereferences or slice out-of-bounds. Named return values let us
	// set err in the deferred recover instead of propagating to Fiber.
	defer func() {
		if r := recover(); r != nil {
			text = ""
			err = fmt.Errorf("PDF parsing panic (likely malformed PDF): %v", r)
		}
	}()

	if !ValidatePDFHeader(pdfData) {
		return "", 0, fmt.Errorf("invalid PDF file header")
	}

	reader := bytes.NewReader(pdfData)
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfData)))
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse PDF: %w", err)
	}

	pageCount = pdfReader.NumPage()
	if pageCount > models.MaxPDFPages {
		return "", pageCount, fmt.Errorf("PDF has %d pages, maximum allowed is %d", pageCount, models.MaxPDFPages)
	}

	var texts []string
	fonts := make(map[string]*pdf.Font)

	for i := 1; i <= pageCount; i++ {
		page := pdfReader.Page(i)
		text, err := page.GetPlainText(fonts)
		if err != nil {
			continue
		}
		texts = append(texts, text)
	}

	result := strings.Join(texts, "\n")
	result = strings.ReplaceAll(result, "\r\n", "\n")
	result = strings.TrimSpace(result)

	return result, pageCount, nil
}

func ExtractPDFTextWithOCR(pdfData []byte) (string, int, error) {
	text, pageCount, err := ExtractPDFText(pdfData)
	if err != nil {
		return "", pageCount, err
	}

	if len(strings.TrimSpace(text)) >= 50 {
		return text, pageCount, nil
	}

	log.Printf("[PDF] Text extraction yielded %d chars, attempting OCR...", len(text))

	ocrText, ocrErr := performOCR(pdfData)
	if ocrErr != nil {
		log.Printf("[PDF] OCR failed: %v, falling back to extracted text", ocrErr)
		return text, pageCount, nil
	}

	if len(strings.TrimSpace(ocrText)) > len(strings.TrimSpace(text)) {
		return ocrText, pageCount, nil
	}

	return text, pageCount, nil
}

func performOCR(pdfData []byte) (string, error) {
	tmpDir, err := os.MkdirTemp("", "ocr-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pdfPath := filepath.Join(tmpDir, "input.pdf")
	if err := os.WriteFile(pdfPath, pdfData, 0644); err != nil {
		return "", fmt.Errorf("failed to write temp PDF: %w", err)
	}

	// Use CommandContext so a hung pdftoppm/tesseract subprocess is killed
	// instead of blocking the handler goroutine forever. A maliciously
	// crafted PDF can cause the renderer to loop or stall; without a
	// timeout the goroutine and subprocess leak until the process OOMs.
	// Each subprocess gets its own timeout — the prior code shared a single
	// 30s budget across pdftoppm + ALL tesseract calls, so a multi-page PDF
	// would silently run out of time mid-OCR (tesseract errors were logged
	// but not surfaced to the user).

	prefix := filepath.Join(tmpDir, "page")
	renderCtx, renderCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer renderCancel()
	cmd := exec.CommandContext(renderCtx, "pdftoppm", "-png", "-r", "300", pdfPath, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pdftoppm failed: %s: %w", string(out), err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", fmt.Errorf("failed to read temp dir: %w", err)
	}

	var results []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".png") {
			imgPath := filepath.Join(tmpDir, entry.Name())
			// R58-B-M1: use an immediately-invoked closure so tessCancel is
			// deferred per-iteration. Previously tessCancel() was called
			// explicitly after CombinedOutput — if CombinedOutput panicked,
			// the context leaked until its 15s timer fired. A plain defer
			// in the for loop would accumulate until function return, so
			// the closure ensures cleanup after each page.
			out, tessErr := func() ([]byte, error) {
				tessCtx, tessCancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer tessCancel()
				tessCmd := exec.CommandContext(tessCtx, "tesseract", imgPath, "stdout", "-l", "eng+chi_sim+jpn+kor+ara+hin")
				return tessCmd.CombinedOutput()
			}()
			if tessErr == nil {
				trimmed := strings.TrimSpace(string(out))
				if len(trimmed) > 0 {
					results = append(results, trimmed)
				}
			} else {
				log.Printf("[PDF] Tesseract failed for %s: %v", entry.Name(), tessErr)
			}
		}
	}

	return strings.Join(results, "\n\n"), nil
}
