package common

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Prepare local JSON data for evaluation
func DecodeVHDPerformanceData(filePath string, holdingMap map[string]map[string]string) {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Could not open %s", filePath)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&holdingMap)
	if err != nil {
		fmt.Printf("Error decoding %s", filePath)
	}
}

// Put data in new map with seconds instead of timestamps
func ConvertTimestampsToSeconds(holdingMap map[string]map[string]string, localBuildPerformanceData map[string]map[string]float64) {
	for key, value := range holdingMap {
		script := make(map[string]float64)
		for section, timeElapsed := range value {
			t, err := time.Parse("15:04:05", timeElapsed)
			if err != nil {
				fmt.Println("Error parsing time in local build JSON data")
			}
			d := t.Sub(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()))
			script[section] = d.Seconds()
		}
		localBuildPerformanceData[key] = script
	}
}

// Parse Kusto data
func ParseKustoData(data *SKU, queriedPerformanceData map[string]map[string][]float64) {
	kustoData := []byte(data.SKUPerformanceData)
	err := json.Unmarshal(kustoData, &queriedPerformanceData)
	if err != nil {
		fmt.Println(err)
	}
}

// Helper function for EvaluatePerformance
func SumArray(arr []float64) float64 {
	var sum float64
	for _, x := range arr {
		sum += x
	}
	return sum
}

// Evaluate performance data
func EvaluatePerformance(localPerformanceData map[string]map[string]float64, queriedPerformanceData map[string]map[string][]float64, regressions map[string]map[string]float64) map[string]map[string]float64 {

	for scriptName, scriptData := range localPerformanceData {
		for section, timeElapsed := range scriptData {
			maxTimeAllowed := SumArray(queriedPerformanceData[scriptName][section])
			if timeElapsed > maxTimeAllowed {
				if regressions[scriptName] == nil {
					regressions[scriptName] = make(map[string]float64)
				}
				regressions[scriptName][section] = timeElapsed - maxTimeAllowed
			}
		}
	}
	return regressions
}

// Print regressions
func PrintRegressions(regressions map[string]map[string]float64) {

	prefix := ""
	indent := "  "

	data, err := json.MarshalIndent(regressions, prefix, indent)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(data))
}
