package chart

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
	"github.com/datastream/holtwinters"
	"image/color"
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
	p.Y.Label.Text = ytext
	p.Add(plotter.NewGrid())
	p.Legend.Top = true
	p.Legend.Left = true
	c := make(chan string)
	return &CustomChart{
		Plot:     p,
		LineDone: c,
		LineColors: map[string]color.Color{
			"green": color.RGBA{G: 255, A: 255},
			"blue":  color.RGBA{B: 255, A: 255},
			"red":   color.RGBA{R: 255, A: 255},
			"gray":  color.RGBA{R: 180, G: 180, B: 180, A: 255},
		},
	}
}

func (cc *CustomChart) CreateLine(pval []float64, color string) *plotter.Line {
	lineData := xysPoints(pval)
	l, err := plotter.NewLine(lineData)
	if err != nil {
		panic(err)
	}
	l.LineStyle.Width = vg.Points(1)
	l.LineStyle.Color = cc.LineColors[color]

	return l
}

func (cc *CustomChart) AddLine(legend string, l *plotter.Line) {
	cc.Plot.Add(l)
	cc.Plot.Legend.Add(legend, l)
	cc.LineDone <- "Done"
}

func (cc *CustomChart) CreatePreviousDayLine(pval []float64, color string) {
	go func() {
		l := cc.CreateLine(pval, color)
		c := cc.LineColors[color]
		l.ShadeColor = &c
		cc.AddLine("line", l)
	}()
}

func (cc *CustomChart) CreateCurrentDayLine(pval []float64, color string) {
	go func() {
		l := cc.CreateLine(pval, color)
		cc.AddLine("line", l)
	}()
}

func (cc *CustomChart) CreatePredictLine(pval []float64, color string) {
	go func() {
		prediction, _ := holtwinters.Forecast(pval, 0.1, 0.0035, 0.1, 4, 4)
		l := cc.CreateLine(prediction, color)
		cc.AddLine("predict", l)
	}()
}

func xysPoints(p []float64) plotter.XYs {
	pts := make(plotter.XYs, len(p))
	for i := range pts {
		pts[i].X = float64(i * 15)
		pts[i].Y = p[i]
	}
	return pts
}
