package chart

import (
	"bufio"
	"bytes"
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
	"code.google.com/p/plotinum/vg/vgsvg"
	"encoding/base64"
	"fmt"
	"github.com/datastream/holtwinters"
	"image/color"
	"math"
)

type CustomChart struct {
	Plot       *plot.Plot
	LineDone   chan string
	LineColors map[string]color.Color
}

func New(title, xtext, ytext string) *CustomChart {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = title
	p.X.Label.Text = xtext
	p.X.Tick.Marker = DefaultTicks
	p.Y.Label.Text = ytext
	p.Add(plotter.NewGrid())
	p.Legend.Top = true
	p.Legend.Left = true
	c := make(chan string)
	return &CustomChart{
		Plot:     p,
		LineDone: c,
		LineColors: map[string]color.Color{
			"green":  color.RGBA{G: 255, A: 255},
			"blue":   color.RGBA{B: 255, A: 255},
			"red":    color.RGBA{R: 255, A: 255},
			"gray":   color.RGBA{R: 180, G: 180, B: 180, A: 255},
			"yellow": color.RGBA{R: 255, G: 220, A: 255},
		},
	}
}

func (cc *CustomChart) createLine(pval []float64, color string) *plotter.Line {
	lineData := xysPoints(pval)
	l, err := plotter.NewLine(lineData)
	if err != nil {
		panic(err)
	}
	l.LineStyle.Width = vg.Points(1)
	l.LineStyle.Color = cc.LineColors[color]

	return l
}

func (cc *CustomChart) addLine(legend string, l *plotter.Line) {
	cc.Plot.Add(l)
	cc.Plot.Legend.Add(legend, l)
	cc.LineDone <- "Done"
}

func (cc *CustomChart) CreatePreviousDayLine(pval []float64, color string) {
	go func() {
		l := cc.createLine(pval, color)
		c := cc.LineColors[color]
		l.ShadeColor = &c
		cc.addLine("prev", l)
	}()
}

func (cc *CustomChart) CreateCurrentDayLine(pval []float64, color string) {
	go func() {
		l := cc.createLine(pval, color)
		cc.addLine("curent", l)
	}()
}

func (cc *CustomChart) CreatePredictLine(pval []float64, color string) {
	go func() {
		prediction, _ := holtwinters.Forecast(pval, 0.1, 0.0035, 0.1, 4, 4)
		l := cc.createLine(prediction, color)
		cc.addLine("predict", l)
	}()
}

func (cc *CustomChart) GetRawDataImg(w, h float64) (imgData string) {
	c := vgsvg.New(vg.Points(w), vg.Points(h))
	// Draw to the Canvas.
	da := plot.MakeDrawArea(c)
	cc.Plot.Draw(da)

	// Write the Canvas to a io.Writer.
	var b bytes.Buffer
	buf := bufio.NewWriter(&b)
	defer buf.Flush()
	if _, err := c.WriteTo(buf); err != nil {
		panic(err)
	}
	imgData = base64.StdEncoding.EncodeToString(b.Bytes())
	return
}

func xysPoints(p []float64) plotter.XYs {
	pts := make(plotter.XYs, len(p))
	for i := range pts {
		pts[i].X = float64(i * 15)
		pts[i].Y = p[i]
	}
	return pts
}

func DefaultTicks(min, max float64) (ticks []plot.Tick) {
	const SuggestedTicks = 3
	if max < min {
		panic("illegal range")
	}
	tens := math.Pow10(int(math.Floor(math.Log10(max - min))))
	n := (max - min) / tens
	for n < SuggestedTicks {
		tens /= 10
		n = (max - min) / tens
	}

	majorMult := int(n / SuggestedTicks)
	switch majorMult {
	case 7:
		majorMult = 6
	case 9:
		majorMult = 8
	}
	majorDelta := float64(majorMult) * tens
	val := math.Floor(min/majorDelta) * majorDelta
	for val <= max {
		if val >= min && val <= max {
			ticks = append(ticks, plot.Tick{Value: val, Label: fmt.Sprintf("%g", float32(val))})
		}
		if math.Nextafter(val, val+majorDelta) == val {
			break
		}
		val += majorDelta
	}

	minorDelta := majorDelta / 2
	switch majorMult {
	case 3, 6:
		minorDelta = majorDelta / 3
	case 5:
		minorDelta = majorDelta / 5
	}

	val = math.Floor(min/minorDelta) * minorDelta
	for val <= max {
		found := false
		for _, t := range ticks {
			if t.Value == val {
				found = true
			}
		}
		if val >= min && val <= max && !found {
			ticks = append(ticks, plot.Tick{Value: val})
		}
		if math.Nextafter(val, val+minorDelta) == val {
			break
		}
		val += minorDelta
	}
	return
}
