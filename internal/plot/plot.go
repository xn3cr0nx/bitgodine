package plot

import (
	"fmt"
	"os"

	"github.com/wcharczuk/go-chart"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// Coordinates wraps x and y coordinates to show on plot
type Coordinates struct {
	X []float64 `json:"x"`
	Y []float64 `json:"y"`
}

// MultipleLineChart saves a multiple linechart based on x and y data passed to the function
func MultipleLineChart(title, xLabel, yLabel string, data map[string]Coordinates) (err error) {
	var lines []chart.Series
	h := 0
	for k, coordinates := range data {
		lines = append(lines, chart.ContinuousSeries{
			Name:            k,
			XValues:         coordinates.X,
			XValueFormatter: chart.FloatValueFormatter,
			YValues:         coordinates.Y,
			YValueFormatter: chart.PercentValueFormatter,
		})
		h++
	}

	graph := chart.Chart{
		Background: chart.Style{
			Padding: chart.Box{
				Top:  50,
				Left: 150,
			},
		},
		XAxis: chart.XAxis{
			Name: xLabel,
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%0.f", vf)
				}
				return ""
			},
			Style: chart.Style{
				TextRotationDegrees: 90,
			},
		},
		YAxis: chart.YAxis{
			Name: yLabel,
		},
		Title:  title,
		Series: lines,
	}

	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph),
	}

	f, err := os.Create("plot.png")
	if err != nil {
		return
	}
	defer f.Close()
	err = graph.Render(chart.PNG, f)
	return
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
