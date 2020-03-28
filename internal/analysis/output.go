package analysis

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine_server/internal/plot"
)

// renderPercentageTable prints a table with percentages of heuristics success rate based on passed analysis
func renderPercentageTable(data []float64, heuristicsList heuristics.Mask) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Heuristic", "%"})
	// table.SetBorder(false)
	table.SetCaption(true, "Heuristics success rate")
	list := heuristicsList.ToList()
	for h, perc := range data {
		table.Append([]string{list[h].String(), fmt.Sprintf("%4.2f", perc*100)})
	}
	table.Render()
}

// renderTable prints a generic two columns table
func renderTable(data map[string]float64, column, caption string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{column, "%"})
	// table.SetBorder(false)
	table.SetCaption(true, caption)
	for k, perc := range data {
		table.Append([]string{k, fmt.Sprintf("%4.2f", perc*100)})
	}
	table.Render()
}

// PlotHeuristicsTimeline plots timeseries of heuristics percentage effectiveness for each block representing time series
func PlotHeuristicsTimeline(data map[int32][]float64, min int32, heuristicsList heuristics.Mask) (err error) {
	coordinates := make(map[string]plot.Coordinates)
	x := make([]float64, len(data))
	for height := range data {
		x[height-min] = float64(height)
	}
	list := heuristicsList.ToList()
	for h, heuristic := range list {
		y := make([]float64, len(data))
		for height, vulnerabilites := range data {
			y[int(height-min)] = vulnerabilites[h] + float64(h)
		}
		coordinates[heuristic.String()] = plot.Coordinates{X: x, Y: y}
	}

	title := "Heuristics timeline"
	if len(list) == 1 {
		title = list[0].String() + " timeline"
	}
	err = plot.MultipleLineChart(title, "blocks", "heuristics effectiveness", coordinates)

	// _, err = chttp.POST("http://plotter:5000/plot", coordinates, map[string]string{})
	// if err != nil {
	// 	return
	// }

	return
}

func generateOutput(vuln Graph, chart, criteria string, heuristicsList heuristics.Mask, from, to int32) (err error) {
	switch chart {
	case "timeline":
		var data map[int32][]float64
		switch criteria {
		case "offbyone":
			data = vuln.ExtractOffByOneBug(heuristicsList, from, to)
		default:
			data = vuln.ExtractPercentages(heuristicsList, from, to)
		}
		err = PlotHeuristicsTimeline(data, from, heuristicsList)

	case "percentage":
		var data []float64
		switch criteria {
		case "offbyone":
			data = vuln.ExtractGlobalOffByOneBug(heuristicsList, from, to)
		case "securebasis":
			data = vuln.ExtractGlobalSecureBasisPerc(heuristicsList, from, to)
		default:
			data = vuln.ExtractGlobalPercentages(heuristicsList, from, to)
		}
		title := "Heuristics percentages"
		list := heuristicsList.ToList()
		if len(list) == 1 {
			title = list[0].String() + " percentage"
		}
		err = plot.BarChart(title, heuristicsList.ToHeuristicsList(), data)

	case "combination":
		var data map[string]float64
		switch criteria {
		case "majorityvoting":
			data = vuln.ExtractGlobalMajorityVotingPerc(heuristicsList, from, to)
		default:
			data = vuln.ExtractCombinationPercentages(heuristicsList, from, to)
		}
		renderTable(data, "Combination", "Heuristics combination percentages")

	default:
		var data []float64
		switch criteria {
		case "offbyone":
			data = vuln.ExtractGlobalOffByOneBug(heuristicsList, from, to)
		case "securebasis":
			data = vuln.ExtractGlobalSecureBasisPerc(heuristicsList, from, to)
		default:
			data = vuln.ExtractGlobalPercentages(heuristicsList, from, to)
		}

		renderPercentageTable(data, heuristicsList)
	}
	return
}
