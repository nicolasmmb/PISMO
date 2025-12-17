package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type Config struct {
	BaseURL       string
	Rate          int
	Duration      time.Duration
	PrometheusURL string
}

type EndpointMetrics struct {
	Name           string
	TotalRequests  uint64
	SuccessRate    float64
	MeanLatency    time.Duration
	P50Latency     time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
	MaxLatency     time.Duration
	MinLatency     time.Duration
	RequestsPerSec float64
	StatusCodes    map[string]int
}

type PrometheusMetrics struct {
	TotalRequests float64
	P50Latency    float64
	P95Latency    float64
	P99Latency    float64
}

type AuditResult struct {
	Endpoint      string
	Vegeta        EndpointMetrics
	Prometheus    PrometheusMetrics
	Discrepancies []string
	Passed        bool
}

type BenchmarkResult struct {
	Name    string
	Metrics EndpointMetrics
}

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "Base URL of the API")
	rate := flag.Int("rate", 100, "Requests per second (total, split between endpoints)")
	duration := flag.Duration("duration", 60*time.Second, "Test duration")
	prometheusURL := flag.String("prometheus", "http://localhost:9090", "Prometheus URL")
	flag.Parse()

	config := Config{
		BaseURL:       *baseURL,
		Rate:          *rate,
		Duration:      *duration,
		PrometheusURL: *prometheusURL,
	}

	fmt.Println("🚀 Pismo API Benchmark Tool (Concurrent)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📍 Target: %s\n", config.BaseURL)
	fmt.Printf("⚡ Rate: %d req/s (split between endpoints)\n", config.Rate)
	fmt.Printf("⏱️  Duration: %s\n", config.Duration)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create test accounts for transactions
	accountID, err := createTestAccount(config.BaseURL)
	if err != nil {
		fmt.Printf("❌ Failed to create test account: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Test account created: ID %d\n", accountID)

	fmt.Println("\n🔥 Running Load Tests CONCURRENTLY...")

	// Run all benchmarks concurrently
	results := runConcurrentBenchmarks(config, accountID)

	fmt.Println("\n⏳ Waiting 15s for Prometheus to scrape metrics...")
	time.Sleep(15 * time.Second)

	// Fetch Prometheus metrics and audit
	var audits []AuditResult
	for _, r := range results {
		prom := fetchPrometheusMetrics(config.PrometheusURL, "/"+r.Name)
		audit := auditMetrics(r.Name, r.Metrics, prom)
		audits = append(audits, audit)
	}

	// Print results
	for _, audit := range audits {
		printEndpointResults(audit)
	}

	// Summary
	allPassed := true
	for _, a := range audits {
		if !a.Passed {
			allPassed = false
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if allPassed {
		fmt.Println("✅ ALL BENCHMARKS PASSED")
	} else {
		fmt.Println("❌ SOME BENCHMARKS FAILED")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if !allPassed {
		os.Exit(1)
	}
}

func runConcurrentBenchmarks(config Config, accountID int) []BenchmarkResult {
	var wg sync.WaitGroup
	resultsChan := make(chan BenchmarkResult, 4)

	endpoints := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"healthz", "GET", "/healthz", ""},
		{"accounts", "GET", fmt.Sprintf("/accounts/%d", accountID), ""},
		{"transactions", "POST", "/transactions", fmt.Sprintf(`{"account_id": %d, "operation_type_id": 4, "amount": 100.00}`, accountID)},
	}

	ratePerEndpoint := config.Rate / len(endpoints)
	if ratePerEndpoint < 1 {
		ratePerEndpoint = 1
	}

	for _, ep := range endpoints {
		wg.Add(1)
		go func(name, method, path, body string) {
			defer wg.Done()
			fmt.Printf("   📌 Starting %s %s @ %d req/s...\n", method, path, ratePerEndpoint)

			metrics := runBenchmark(config, method, path, body, ratePerEndpoint)
			resultsChan <- BenchmarkResult{Name: name, Metrics: metrics}

			fmt.Printf("   ✅ Finished %s %s\n", method, path)
		}(ep.name, ep.method, ep.path, ep.body)
	}

	// Wait for all benchmarks to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var results []BenchmarkResult
	for r := range resultsChan {
		results = append(results, r)
	}

	// Sort by name for consistent output
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

func runBenchmark(config Config, method, path, body string, rate int) EndpointMetrics {
	target := vegeta.Target{
		Method: method,
		URL:    config.BaseURL + path,
	}

	if body != "" {
		target.Body = []byte(body)
		target.Header = http.Header{
			"Content-Type": []string{"application/json"},
		}
	}

	targeter := vegeta.NewStaticTargeter(target)
	attacker := vegeta.NewAttacker()
	vegetaRate := vegeta.Rate{Freq: rate, Per: time.Second}

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, vegetaRate, config.Duration, path) {
		metrics.Add(res)
	}
	metrics.Close()

	return toEndpointMetrics(path, metrics)
}

func createTestAccount(baseURL string) (int, error) {
	payload := fmt.Sprintf(`{"document_number":"%d"}`, time.Now().UnixNano())
	resp, err := http.Post(baseURL+"/accounts", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		AccountID int `json:"account_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 1, nil // fallback to account 1
	}
	return result.AccountID, nil
}

func toEndpointMetrics(name string, m vegeta.Metrics) EndpointMetrics {
	statusCodes := make(map[string]int)
	for code, count := range m.StatusCodes {
		statusCodes[code] = int(count)
	}

	return EndpointMetrics{
		Name:           name,
		TotalRequests:  m.Requests,
		SuccessRate:    m.Success * 100,
		MeanLatency:    m.Latencies.Mean,
		P50Latency:     m.Latencies.P50,
		P95Latency:     m.Latencies.P95,
		P99Latency:     m.Latencies.P99,
		MaxLatency:     m.Latencies.Max,
		MinLatency:     m.Latencies.Min,
		RequestsPerSec: m.Rate,
		StatusCodes:    statusCodes,
	}
}

func fetchPrometheusMetrics(prometheusURL, path string) PrometheusMetrics {
	var pm PrometheusMetrics

	query := fmt.Sprintf(`sum(http_requests_total{path="%s"})`, path)
	totalReq, err := queryPrometheus(prometheusURL, query)
	if err == nil {
		pm.TotalRequests = totalReq
	}

	query = fmt.Sprintf(`histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{path="%s"}[5m])) by (le)) * 1000000`, path)
	p50, err := queryPrometheus(prometheusURL, query)
	if err == nil {
		pm.P50Latency = p50
	}

	query = fmt.Sprintf(`histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{path="%s"}[5m])) by (le)) * 1000000`, path)
	p95, err := queryPrometheus(prometheusURL, query)
	if err == nil {
		pm.P95Latency = p95
	}

	query = fmt.Sprintf(`histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{path="%s"}[5m])) by (le)) * 1000000`, path)
	p99, err := queryPrometheus(prometheusURL, query)
	if err == nil {
		pm.P99Latency = p99
	}

	return pm
}

func queryPrometheus(baseURL, query string) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", baseURL, query)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	if len(result.Data.Result) == 0 || len(result.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("no data")
	}

	valueStr, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("invalid value type")
	}

	var value float64
	fmt.Sscanf(valueStr, "%f", &value)
	return value, nil
}

func auditMetrics(name string, vegeta EndpointMetrics, prom PrometheusMetrics) AuditResult {
	audit := AuditResult{
		Endpoint:   name,
		Vegeta:     vegeta,
		Prometheus: prom,
		Passed:     true,
	}

	if prom.TotalRequests > 0 {
		diff := math.Abs(float64(vegeta.TotalRequests) - prom.TotalRequests)
		tolerance := float64(vegeta.TotalRequests) * 0.15 // 15% tolerance for concurrent tests
		if diff > tolerance {
			audit.Discrepancies = append(audit.Discrepancies,
				fmt.Sprintf("Request count: Vegeta=%d, Prometheus=%.0f (diff=%.0f)",
					vegeta.TotalRequests, prom.TotalRequests, diff))
			audit.Passed = false
		}
	}

	vegetaP50us := float64(vegeta.P50Latency.Microseconds())
	if prom.P50Latency > 0 && vegetaP50us > 0 {
		diff := math.Abs(vegetaP50us - prom.P50Latency)
		tolerance := vegetaP50us * 0.30
		if diff > tolerance {
			audit.Discrepancies = append(audit.Discrepancies,
				fmt.Sprintf("P50: Vegeta=%.0fµs, Prometheus=%.0fµs",
					vegetaP50us, prom.P50Latency))
		}
	}

	vegetaP95us := float64(vegeta.P95Latency.Microseconds())
	if prom.P95Latency > 0 && vegetaP95us > 0 {
		diff := math.Abs(vegetaP95us - prom.P95Latency)
		tolerance := vegetaP95us * 0.30
		if diff > tolerance {
			audit.Discrepancies = append(audit.Discrepancies,
				fmt.Sprintf("P95: Vegeta=%.0fµs, Prometheus=%.0fµs",
					vegetaP95us, prom.P95Latency))
		}
	}

	return audit
}

func printEndpointResults(audit AuditResult) {
	v := audit.Vegeta
	p := audit.Prometheus

	icon := "🩺"
	switch audit.Endpoint {
	case "transactions":
		icon = "💳"
	case "accounts":
		icon = "👤"
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📊 %s /%s\n", icon, audit.Endpoint)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println("\n🎯 Vegeta Results:")
	fmt.Printf("   Total Requests:  %d\n", v.TotalRequests)
	fmt.Printf("   Success Rate:    %.2f%%\n", v.SuccessRate)
	fmt.Printf("   Requests/sec:    %.2f\n", v.RequestsPerSec)
	fmt.Printf("   Mean Latency:    %s\n", v.MeanLatency)
	fmt.Printf("   P50 Latency:     %s\n", v.P50Latency)
	fmt.Printf("   P95 Latency:     %s\n", v.P95Latency)
	fmt.Printf("   P99 Latency:     %s\n", v.P99Latency)

	fmt.Println("\n   Status Codes:")
	codes := make([]string, 0, len(v.StatusCodes))
	for code := range v.StatusCodes {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	for _, code := range codes {
		fmt.Printf("     %s: %d\n", code, v.StatusCodes[code])
	}

	fmt.Println("\n📈 Prometheus Metrics:")
	fmt.Printf("   Total Requests:  %.0f\n", p.TotalRequests)
	fmt.Printf("   P50 Latency:     %.0f µs\n", p.P50Latency)
	fmt.Printf("   P95 Latency:     %.0f µs\n", p.P95Latency)
	fmt.Printf("   P99 Latency:     %.0f µs\n", p.P99Latency)

	fmt.Println("\n🔍 Audit:")
	if len(audit.Discrepancies) == 0 {
		fmt.Println("   ✅ All metrics match within tolerance")
	} else {
		fmt.Println("   ⚠️  Discrepancies:")
		for _, d := range audit.Discrepancies {
			fmt.Printf("      • %s\n", d)
		}
	}

	if audit.Passed {
		fmt.Println("   ✅ PASSED")
	} else {
		fmt.Println("   ❌ FAILED")
	}
}
