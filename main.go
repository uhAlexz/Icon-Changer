package main

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"net/http"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
)

// Main function
func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/process", processImageHandler)
	fmt.Println("Server running on :32015")
	r.Run(":32015")
}

// Handler function
func processImageHandler(c *gin.Context) {
	imageURL := c.Query("image")
	hueStr := c.Query("hue")

	if imageURL == "" || hueStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing image or hue"})
		return
	}

	hue, err := strconv.Atoi(hueStr)
	if err != nil || hue < 0 || hue > 360 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hue value (must be 0-360)"})
		return
	}

	img, err := downloadImage(imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download image"})
		return
	}

	modifiedImg := changeHue(img, float64(hue))

	c.Header("Content-Type", "image/png")
	png.Encode(c.Writer, modifiedImg)
}

// Download Image
func downloadImage(imgURL string) (image.Image, error) {
	resp, err := http.Get(imgURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}

// Change Image Hue
func changeHue(img image.Image, hue float64) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	dc := gg.NewContext(w, h)
	dc.DrawImage(img, 0, 0)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			newR, newG, newB := applyHueShift(float64(r>>8), float64(g>>8), float64(b>>8), hue)
			dc.SetRGBA(float64(newR)/255, float64(newG)/255, float64(newB)/255, float64(a)/65535)
			dc.SetPixel(x, y)
		}
	}

	return dc.Image()
}

// Apply Hue Shift
func applyHueShift(r, g, b, hue float64) (float64, float64, float64) {
	u := math.Cos(hue * math.Pi / 180)
	w := math.Sin(hue * math.Pi / 180)

	newR := (.299 + .701*u + .168*w)*r + (.587 - .587*u + .330*w)*g + (.114 - .114*u - .497*w)*b
	newG := (.299 - .299*u - .328*w)*r + (.587 + .413*u + .035*w)*g + (.114 - .114*u + .292*w)*b
	newB := (.299 - .3*u + 1.25*w)*r + (.587 - .588*u - 1.05*w)*g + (.114 + .886*u - .203*w)*b

	return clamp(newR), clamp(newG), clamp(newB)
}

// Clamp values to 0-255
func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}
