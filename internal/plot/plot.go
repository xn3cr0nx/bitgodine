package plot

import (
	"math/rand"
	"os"

	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgsvg"
)

func HeuristicsTimeline() (err error) {
	rand.Seed(int64(0))

	p, err := plot.New()
	if err != nil {
		return
	}

	p.Title.Text = "Heuristics timeline"
	p.X.Label.Text = "blocks"
	p.Y.Label.Text = "heuristics effectiveness"

	err = plotutil.AddLinePoints(p,
		"First", randomPoints(15),
		"Second", randomPoints(15),
		"Third", randomPoints(15))
	if err != nil {
		return
	}

	// Save the plot to a PNG file.
	if err = p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		return
	}

	c := vgsvg.New(3*vg.Inch, 3*vg.Inch)
	// Draw to the Canvas.
	p.Draw(draw.New(c))
	// Write the Canvas to a io.Writer (in this case, os.Stdout).
	if _, err := c.WriteTo(os.Stdout); err != nil {
		panic(err)
	}

	return
}

// randomPoints returns some random x, y points.
func randomPoints(n int) plotter.XYs {
	pts := make(plotter.XYs, n)
	for i := range pts {
		if i == 0 {
			pts[i].X = rand.Float64()
		} else {
			pts[i].X = pts[i-1].X + rand.Float64()
		}
		pts[i].Y = pts[i].X + 10*rand.Float64()
	}
	return pts
}

// HeuristicsPercentages plots the percentage effectiveness for each heuristic in the passed time range
func HeuristicsPercentages(percentages []float64) (err error) {
	p, err := plot.New()
	if err != nil {
		return
	}
	w := vg.Points(20)
	p.Title.Text = "Heuristics effectiveness"
	p.Y.Label.Text = "Percentage"
	p.Y.Min = 0
	p.Y.Max = 100
	// p.X.Label.Text = "Heuristics"

	for h := 0; h < len(percentages); h++ {
		bar, e := plotter.NewBarChart(plotter.Values{percentages[h] * 100}, w)
		if e != nil {
			err = e
			return
		}
		bar.LineStyle.Width = vg.Length(0)
		bar.Color = plotutil.Color(h)
		// bar.Offset = -w
		bar.Offset = -w * vg.Length(h)
		p.Add(bar)
		p.Legend.Add(heuristics.Heuristic(h).String(), bar)
	}

	p.Legend.Top = true
	p.HideX()
	// p.NominalX(heuristics.List()...)
	err = p.Save(5*vg.Inch, 3*vg.Inch, "barchart.png")
	return
}
