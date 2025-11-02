package services

import (
	"image/color"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

type InfographicService struct {
	colorSchemes []ColorScheme
}

type ColorScheme struct {
	Background color.Color
	Accent     color.Color
	Text       color.Color
}

func NewInfographicService() *InfographicService {
	return &InfographicService{
		colorSchemes: []ColorScheme{
			{color.RGBA{245, 247, 250, 255}, color.RGBA{59, 130, 246, 255}, color.RGBA{17, 24, 39, 255}},
			{color.RGBA{249, 250, 251, 255}, color.RGBA{16, 185, 129, 255}, color.RGBA{17, 24, 39, 255}},
			{color.RGBA{254, 249, 245, 255}, color.RGBA{249, 115, 22, 255}, color.RGBA{17, 24, 39, 255}},
			{color.RGBA{250, 245, 255, 255}, color.RGBA{168, 85, 247, 255}, color.RGBA{17, 24, 39, 255}},
			{color.RGBA{254, 242, 242, 255}, color.RGBA{239, 68, 68, 255}, color.RGBA{17, 24, 39, 255}},
		},
	}
}

func (s *InfographicService) Generate(title, newsletterText, outputPath string) error {
	rand.NewSource(time.Now().UnixNano())
	scheme := s.colorSchemes[rand.Intn(len(s.colorSchemes))]

	const width, height = 1200, 1600
	dc := gg.NewContext(width, height)

	// Background
	dc.SetColor(scheme.Background)
	dc.Clear()

	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return err
	}

	// Top accent bar
	dc.SetColor(scheme.Accent)
	dc.DrawRectangle(0, 0, width, 120)
	dc.Fill()

	// Title
	face := truetype.NewFace(font, &truetype.Options{Size: 52})
	dc.SetFontFace(face)
	dc.SetColor(color.White)
	dc.DrawStringWrapped(title, width/2, 60, 0.5, 0.5, width-100, 1.5, gg.AlignCenter)

	// Decorative line
	dc.SetColor(scheme.Accent)
	dc.SetLineWidth(3)
	dc.DrawLine(100, 180, width-100, 180)
	dc.Stroke()

	// Content
	s.renderContent(dc, font, newsletterText, scheme, width, height)

	// Footer
	s.renderFooter(dc, font, scheme, width, height)

	os.MkdirAll("cache", 0755)

	return dc.SavePNG(outputPath)
}

func (s *InfographicService) renderContent(dc *gg.Context, font *truetype.Font, text string, scheme ColorScheme, width, height float64) {
	face := truetype.NewFace(font, &truetype.Options{Size: 24})
	dc.SetFontFace(face)
	dc.SetColor(scheme.Text)

	paragraphs := strings.Split(text, "\n")
	yPos := 240.0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			yPos += 20
			continue
		}

		if strings.HasPrefix(para, "##") {
			face = truetype.NewFace(font, &truetype.Options{Size: 32})
			dc.SetFontFace(face)
			dc.SetColor(scheme.Accent)
			cleanPara := strings.TrimPrefix(para, "##")
			dc.DrawStringWrapped(strings.TrimSpace(cleanPara), 100, yPos, 0, 0, width-200, 1.5, gg.AlignLeft)
			yPos += 60
			face = truetype.NewFace(font, &truetype.Options{Size: 24})
			dc.SetFontFace(face)
			dc.SetColor(scheme.Text)
		} else {
			dc.DrawStringWrapped(para, 100, yPos, 0, 0, width-200, 1.8, gg.AlignLeft)
			yPos += 100
		}

		if yPos > height-150 {
			break
		}
	}
}

func (s *InfographicService) renderFooter(dc *gg.Context, font *truetype.Font, scheme ColorScheme, width, height float64) {
	dc.SetColor(scheme.Accent)
	dc.DrawRectangle(0, height-80, width, 80)
	dc.Fill()

	face := truetype.NewFace(font, &truetype.Options{Size: 20})
	dc.SetFontFace(face)
	dc.SetColor(color.White)
	dc.DrawString(time.Now().Format("January 2, 2006"), width/2-100, height-40)
}
