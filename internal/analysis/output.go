package analysis

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine_server/internal/plot"
)

// GlobalPercentages prints a table with percentages of heuristics success rate based on passed analysis
func GlobalPercentages(data []float64, heuristicsList []string) (err error) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Heuristic", "%"})
	// table.SetBorder(false)
	table.SetCaption(true, "Heuristics success rate")
	for h, perc := range data {
		table.Append([]string{heuristicsList[h], fmt.Sprintf("%4.2f", perc*100)})
	}
	table.Render()
	return
}

// PlotHeuristicsTimeline plots timeseries of heuristics percentage effectiveness for each block representing time series
func PlotHeuristicsTimeline(data map[int32][]float64, min int32, heuristicsList []string) (err error) {
	coordinates := make(map[string]plot.Coordinates)
	x := make([]float64, len(data))
	for height := range data {
		x[height-min] = float64(height)
	}

	for h, heuristic := range heuristicsList {
		y := make([]float64, len(data))
		for height, vulnerabilites := range data {
			y[int(height-min)] = vulnerabilites[h] + float64(h)
		}
		coordinates[heuristic] = plot.Coordinates{X: x, Y: y}
	}

	title := "Heuristics timeline"
	if len(heuristicsList) == 1 {
		title = heuristicsList[0] + " timeline"
	}
	err = plot.MultipleLineChart(title, "blocks", "heuristics effectiveness", coordinates)

	return
}

func generateOutput(vuln Graph, chart string, heuristicsList []string, from, to int32) (err error) {
	switch chart {
	case "timeline":
		data := vuln.ExtractPercentages(heuristicsList, from, to)
		err = PlotHeuristicsTimeline(data, from, heuristicsList)
	case "percentage":
		data := vuln.ExtractGlobalPercentages(heuristicsList, from, to)
		title := "Heuristics percentages"
		if len(heuristicsList) == 1 {
			title = heuristicsList[0] + " percentage"
		}
		err = plot.BarChart(title, heuristicsList, data)
	default:
		data := vuln.ExtractGlobalPercentages(heuristicsList, from, to)
		err = GlobalPercentages(data, heuristicsList)
	}
	return
}
