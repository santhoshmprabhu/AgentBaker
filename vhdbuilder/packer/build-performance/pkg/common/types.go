package common

type Config struct {
	KustoTable                string
	KustoEndpoint             string
	KustoDatabase             string
	KustoClientID             string
	SigImageName              string
	LocalBuildPerformanceFile string
	SourceBranch              string
}

type DataMaps struct {
	LocalPerformanceDataMap   map[string]map[string]float64
	QueriedPerformanceDataMap map[string]map[string][]float64
	RegressionMap             map[string]map[string]float64
}

type SKU struct {
	Name               string `kusto:"SIG_IMAGE_NAME"`
	SKUPerformanceData string `kusto:"BUILD_PERFORMANCE"`
}
