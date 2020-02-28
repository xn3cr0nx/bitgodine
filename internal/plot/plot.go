package plot

import (
	"errors"
	"fmt"
	"os"

	"github.com/wcharczuk/go-chart"
)

// Coordinates wraps x and y coordinates to show on plot
type Coordinates struct {
	X []float64 `json:"x"`
	Y []float64 `json:"y"`
}

// MultipleLineChart saves a multiple linechart based on x and y data passed to the function
func MultipleLineChart(title, xLabel, yLabel string, data map[string]Coordinates) (err error) {
	var lines []chart.Series
	for k, coordinates := range data {
		line := chart.ContinuousSeries{
			Name:            k,
			XValues:         coordinates.X,
			XValueFormatter: chart.FloatValueFormatter,
			YValues:         coordinates.Y,
			YValueFormatter: chart.PercentValueFormatter,
		}
		if len(data) == 1 {
			// for k := range data {
			line.Style.FillColor = chart.GetDefaultColor(0)
			line.Style.StrokeColor = chart.GetDefaultColor(0)
			// }
		}
		lines = append(lines, line)
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
				FontSize:            15,
			},
		},
		YAxis: chart.YAxis{
			Name: yLabel,
			Style: chart.Style{
				FontSize: 15,
			},
		},
		Title:  title,
		Series: lines,
		Width:  1920,
		Height: 1080,
	}

	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph),
	}

	f, err := os.Create("timeline.png")
	if err != nil {
		return
	}
	defer f.Close()
	err = graph.Render(chart.PNG, f)
	return
}

// BarChart plots a multibar chart
func BarChart(title string, xLabel []string, percentages []float64) (err error) {
	if len(xLabel) != len(percentages) {
		err = errors.New("Wrong arguments length")
		return
	}

	var bars []chart.Value
	for i, v := range percentages {
		bars = append(bars, chart.Value{
			Value: v,
			Label: xLabel[i],
			// Style: chart.Style{
			// 	// StrokeWidth: .01,
			// 	FillColor:   heuristics.Color(xLabel[i]),
			// 	StrokeColor: heuristics.Color(xLabel[i]),
			// },
		})
	}

	stackedBarChart := chart.BarChart{
		Title:      title,
		TitleStyle: chart.Shown(),
		Background: chart.Style{
			Padding: chart.Box{
				Top: 100,
			},
		},
		Width:  1920,
		Height: 1080,
		XAxis: chart.Style{
			FontSize: 15,
		},
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: 1,
			},
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%0.f", vf*float64(100))
				}
				return ""
			},
			Style: chart.Style{
				FontSize: 15,
			},
		},
		BarSpacing: 100,
		BarWidth:   150,
		Bars:       bars,
	}

	f, err := os.Create("percentage.png")
	if err != nil {
		return
	}
	defer f.Close()
	err = stackedBarChart.Render(chart.PNG, f)
	return
}

// // MultipleBarChart saves a multiple barchart based on x and y data passed to the function
// func MultipleBarChart(title, xLabel, yLabel string, data map[string]Coordinates) (err error) {
// 	// bars := make(map[string][]chart.Value, len(data))
// 	var bars []chart.Value
// 	for _, coordinates := range data {
// 		for _, x := range coordinates.X {
// 			// bars[k] = append(bars[k], chart.Value{
// 			// 	Value: x,
// 			// 	Label: coordinates.Y[i],
// 			// })
// 			bars = append(bars, chart.Value{
// 				Value: x,
// 				Label: coordinates.Y[i],
// 			})
// 		}
// 	}

// 	stackedBarChart := chart.BarChart{
// 		Title:      title,
// 		TitleStyle: chart.Shown(),
// 		Background: chart.Style{
// 			Padding: chart.Box{
// 				Top: 100,
// 			},
// 		},
// 		Width:  1920,
// 		Height: 1080,
// 		XAxis: chart.Style{
// 			FontSize: 15,
// 		},
// 		YAxis: chart.YAxis{
// 			Range: &chart.ContinuousRange{
// 				Min: 0,
// 				Max: 1,
// 			},
// 			ValueFormatter: func(v interface{}) string {
// 				if vf, isFloat := v.(float64); isFloat {
// 					return fmt.Sprintf("%0.f", vf*float64(100))
// 				}
// 				return ""
// 			},
// 			Style: chart.Style{
// 				FontSize: 15,
// 			},
// 		},
// 		BarSpacing: 100,
// 		BarWidth:   150,
// 		Bars:       bars,
// 	}

// 	f, err := os.Create("timebar.png")
// 	if err != nil {
// 		return
// 	}
// 	defer f.Close()
// 	err = stackedBarChart.Render(chart.PNG, f)
// 	return
// }
