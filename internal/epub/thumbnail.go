package epub

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// DefaultMaxCoverWidth represents the maximum width for downscaled thumbnails.
const DefaultMaxCoverWidth = 400

// SaveCoverThumbnail decodes an image from bytes, downscales it to maxWidth if necessary
// (maintaining aspect ratio), and encodes it as a WebP image at destPath.
func SaveCoverThumbnail(coverBytes []byte, destPath string, maxWidth int) error {
	if len(coverBytes) == 0 {
		return fmt.Errorf("empty cover bytes")
	}
	if maxWidth <= 0 {
		maxWidth = DefaultMaxCoverWidth
	}

	srcImg, _, err := image.Decode(bytes.NewReader(coverBytes))
	if err != nil {
		return fmt.Errorf("failed to decode cover image: %w", err)
	}

	bounds := srcImg.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	if origW == 0 || origH == 0 {
		return fmt.Errorf("invalid image dimensions: %dx%d", origW, origH)
	}

	var finalImg image.Image = srcImg

	// Downscale only if original width exceeds maxWidth
	if origW > maxWidth {
		targetW := maxWidth
		targetH := int(float64(origH) * float64(targetW) / float64(origW))
		if targetH < 1 {
			targetH = 1
		}

		dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
		draw.CatmullRom.Scale(dst, dst.Bounds(), srcImg, bounds, draw.Over, nil)
		finalImg = dst
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create thumbnail directory %s: %w", destDir, err)
	}

	// Write atomically using a temporary file
	tmpPath := destPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp thumbnail file: %w", err)
	}

	// Encode to pure-Go WebP
	if err := nativewebp.Encode(tmpFile, finalImg, nil); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to encode webp thumbnail: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp thumbnail file: %w", err)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp thumbnail to %s: %w", destPath, err)
	}

	return nil
}
