package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func main() {
	printStep("Starting exporter")
	printStep("Reading file...")
	file, err := os.Open("./list.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	printStep("Scanning file...")
	rows := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		rows = append(rows, convert(text))
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	printStep("Exporting rows...")
	if err := export(rows); err != nil {
		log.Fatal(err)
	}

	printStep("Export completed")
}

func convert(mask string) (converted string) {
	numbers := strings.Split(mask, "")
	for i, n := range numbers {
		if n != "0" {
			converted = fmt.Sprintf("%s%s", converted, maskMap(i))
		}
	}

	if converted == "" {
		converted = "-"
	}

	tabs := 8 - len(converted)
	for i := 0; i < tabs; i++ {
		converted = fmt.Sprintf(" %s", converted)
	}

	return converted
}

func maskMap(index int) string {
	return map[int]string{
		0: "B",
		1: "S",
		2: "R",
		3: "T",
		4: "O",
		5: "P",
		6: "C",
		7: "L",
	}[index]
}

func export(list []string) (err error) {
	file, err := os.Create("./exported.txt")
	if err != nil {
		return
	}
	writer := bufio.NewWriter(file)
	for _, line := range list {
		_, err = writer.WriteString(line + "\n")
		if err != nil {
			err = errors.Wrapf(err, "Got error while writing to a file. Err: %s", err.Error())
			return
		}
		fmt.Println(line)
	}
	writer.Flush()
	return
}

func printStep(text string) {
	fmt.Println("-----------------------------")
	fmt.Println(text)
	fmt.Println("-----------------------------")
}
