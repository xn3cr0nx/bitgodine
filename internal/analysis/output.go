package analysis

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine/internal/plot"
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

// renderFullPercentageTable prints a table with percentages of heuristics success rate based on passed analysis
func renderFullPercentageTable(data AnalysisSet, caption string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Combination", "secure perc (%)", "# secure set", "total perc (%)", "# total set"})
	// table.SetBorder(false)
	table.SetCaption(true, caption)
	for h, n := range data.Counters {
		table.Append([]string{fmt.Sprintf("%b", h), fmt.Sprintf("%4.2f", data.LocalPercentages[h]*100), fmt.Sprintf("%4.0f", data.LocalCounters[h]), fmt.Sprintf("%4.2f", data.Percentages[h]*100), fmt.Sprintf("%4.0f", n)})
	}
	table.Render()
}

// renderComparingPercentageTable prints a table with percentages of heuristics success rate based on passed analysis
func renderComparingPercentageTable(base, data AnalysisSet, caption string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Combination", "secure perc (%)", "# secure set", "total perc (%)", "# total set"})
	// table.SetBorder(false)
	table.SetCaption(true, caption)
	localVars := make([]float64, 0)
	for h, n := range base.Counters {
		localSetVar := data.LocalCounters[h] - base.LocalCounters[h]
		localSet := fmt.Sprintf("%.0f (%.0f)", data.LocalCounters[h], localSetVar)

		localPercVar := data.LocalPercentages[h] - base.LocalPercentages[h]
		localPerc := fmt.Sprintf("%.2f (%.2f)", data.LocalPercentages[h]*100, localPercVar*100)
		localVars = append(localVars, localPercVar)

		setVar := data.Counters[h] - n
		set := fmt.Sprintf("%.0f (%.0f)", data.Counters[h], setVar)

		percVar := data.Percentages[h] - base.Percentages[h]
		perc := fmt.Sprintf("%.2f (%.2f)", data.Percentages[h]*100, percVar*100)

		table.Append([]string{fmt.Sprintf("%b", h), localPerc, localSet, perc, set})
	}
	table.Render()

	tot := float64(0)
	for _, e := range localVars {
		tot += e
	}
	fmt.Println("Secure Perc offset mean", tot/float64(len(localVars))*100)
}

// renderOverlapTable prints a table with percentages of heuristics success rate based on passed analysis
func renderOverlapTable(data AnalysisSet, heuristicsList heuristics.Mask, wide string) {
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"Heuristic"}
	header = append(header, heuristicsList.ToHeuristicsList()...)
	table.SetHeader(header)
	// table.SetBorder(false)
	table.SetCaption(true, "Heuristics overlap")
	list := heuristicsList.ToList()
	set := data.LocalCounters
	if wide == "majority" {
		set = data.Counters
	}
	if wide == "full" {
		set = data.Combinations
	}
	for _, h := range list {
		var overlap [8]float64
		counter := 0
		for k := range set {
			mask := heuristics.Mask([3]byte{k, byte(0), byte(0)})
			if mask.VulnerableMask(h) {
				counter++
				for i := heuristics.Heuristic(0); i < 8; i++ {
					if mask.VulnerableMask(i) {
						overlap[int(i)] = overlap[int(i)] + 1
					}
				}
			}
		}

		if counter == 0 {
			table.Append([]string{h.String(), "0", "0", "0", "0", "0", "0", "0", "0"})
			continue
		}
		row := []string{h.String()}
		for i := 0; i < 8; i++ {
			row = append(row, fmt.Sprintf("%.2f", (overlap[i]/float64(counter))*100))
		}
		table.Append(row)
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
		case "fullmajorityanalysis":
			set := vuln.MajorityFullAnalysis(heuristicsList, from, to)
			renderFullPercentageTable(set, "Majority voting combination full analysis")
			return
		case "reducingmajorityanalysis":
			full := vuln.MajorityFullAnalysis(heuristicsList, from, to)
			reducing := []heuristics.Heuristic{0, 1, 2, 3, 4, 7}
			for _, r := range reducing {
				set := vuln.MajorityFullAnalysis(heuristicsList, from, to, r)
				renderComparingPercentageTable(full, set, fmt.Sprintf("Majority voting combination full analysis, reduced by heuristic %s", r.String()))
			}
			return
		case "fullmajorityvoting":
			data = vuln.ExtractGlobalFullMajorityVotingPerc(heuristicsList, from, to)
		case "majorityvoting":
			data = vuln.ExtractGlobalMajorityVotingPerc(heuristicsList, from, to)
		case "strictmajorityvoting":
			data = vuln.ExtractGlobalStricMajorityVotingPerc(heuristicsList, from, to)
		case "overlapping":
			set := vuln.MajorityFullAnalysis(heuristicsList, from, to)
			renderOverlapTable(set, heuristicsList, "")
			renderOverlapTable(set, heuristicsList, "majority")
			renderOverlapTable(set, heuristicsList, "full")
			return
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
