package chart

import (
	// "ccscharts/wsconn"
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
	"github.com/datastream/holtwinters"
	"image/color"
)

type CustomChart struct {
	Plot     *plot.Plot
	LineDone chan string
}

func New() *CustomChart {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Calls data"
	p.X.Label.Text = "Minutes"
	p.Y.Label.Text = "Calls"
	p.Add(plotter.NewGrid())
	p.Legend.Top = true
	p.Legend.Left = true
	c := make(chan string)
	return &CustomChart{
		Plot:     p,
		LineDone: c,
	}
}

func (cc *CustomChart) CreateCurrentLine(pval []float64) {
	go func() {
		lineData := xysPoints(pval)
		l, err := plotter.NewLine(lineData)
		if err != nil {
			panic(err)
		}
		gray := color.Color(color.RGBA{R: 180, G: 180, B: 180, A: 1})
		l.ShadeColor = &gray
		l.LineStyle.Width = vg.Points(1)
		l.LineStyle.Color = color.RGBA{R: 180, G: 180, B: 180, A: 255}
		cc.Plot.Add(l)
		cc.Plot.Legend.Add("line", l)
		cc.LineDone <- "Done"
	}()
}

func (cc *CustomChart) CreatePredictLine(pval []float64) {
	go func() {
		prediction, _ := holtwinters.Forecast(pval, 0.1, 0.0035, 0.1, 4, 4)
		predictionData := xysPoints(prediction)
		prl, err := plotter.NewLine(predictionData)
		if err != nil {
			panic(err)
		}
		prl.LineStyle.Width = vg.Points(1)
		prl.LineStyle.Color = color.RGBA{G: 255, A: 255}
		cc.Plot.Add(prl)
		cc.Plot.Legend.Add("predict", prl)
		cc.LineDone <- "Done"
	}()
}

// randomPoints returns some random x, y points.
func xysPoints(p []float64) plotter.XYs {
	pts := make(plotter.XYs, len(p))
	for i := range pts {
		pts[i].X = float64(i * 15)
		pts[i].Y = p[i]
	}
	return pts
}
