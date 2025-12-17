package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type Config struct {
	BaseURL  string
	Rate     int
	Duration time.Duration
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

type BenchmarkResult struct {
	Name    string
	Metrics EndpointMetrics
}

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "Base URL of the API")
	rate := flag.Int("rate", 100, "Total requests per second (split between endpoints)")
	duration := flag.Duration("duration", 60*time.Second, "Test duration")
	flag.Parse()

	config := Config{
		BaseURL:  *baseURL,
		Rate:     *rate,
		Duration: *duration,
	}

	fmt.Println("🚀 Pismo API Benchmark")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📍 Target: %s\n", config.BaseURL)
	fmt.Printf("⚡ Rate: %d req/s total\n", config.Rate)
	fmt.Printf("⏱️  Duration: %s\n", config.Duration)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create test account
	accountID, err := createTestAccount(config.BaseURL)
	if err != nil {
		fmt.Printf("❌ Failed to create test account: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Test account created: ID %d\n", accountID)

	fmt.Println("\n🔥 Running Benchmarks Concurrently...\n")

	// Run all benchmarks concurrently
	results := runConcurrentBenchmarks(config, accountID)

	// Print results
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 RESULTS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	allSuccess := true
	for _, r := range results {
		printResult(r)
		if r.Metrics.SuccessRate < 100 {
			allSuccess = false
		}
	}

	// Summary
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if allSuccess {
		fmt.Println("✅ ALL BENCHMARKS COMPLETED SUCCESSFULLY")
	} else {
		fmt.Println("⚠️  SOME REQUESTS FAILED")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func runConcurrentBenchmarks(config Config, accountID int) []BenchmarkResult {
	var wg sync.WaitGroup
	resultsChan := make(chan BenchmarkResult, 3)

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
			fmt.Printf("   📌 %s %s @ %d req/s\n", method, path, ratePerEndpoint)

			metrics := runBenchmark(config, method, path, body, ratePerEndpoint)
			resultsChan <- BenchmarkResult{Name: name, Metrics: metrics}
		}(ep.name, ep.method, ep.path, ep.body)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var results []BenchmarkResult
	for r := range resultsChan {
		results = append(results, r)
	}

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

	statusCodes := make(map[string]int)
	for code, count := range metrics.StatusCodes {
		statusCodes[code] = int(count)
	}

	return EndpointMetrics{
		Name:           path,
		TotalRequests:  metrics.Requests,
		SuccessRate:    metrics.Success * 100,
		MeanLatency:    metrics.Latencies.Mean,
		P50Latency:     metrics.Latencies.P50,
		P95Latency:     metrics.Latencies.P95,
		P99Latency:     metrics.Latencies.P99,
		MaxLatency:     metrics.Latencies.Max,
		MinLatency:     metrics.Latencies.Min,
		RequestsPerSec: metrics.Rate,
		StatusCodes:    statusCodes,
	}
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
		return 1, nil
	}
	return result.AccountID, nil
}

func printResult(r BenchmarkResult) {
	m := r.Metrics

	icon := "🩺"
	switch r.Name {
	case "transactions":
		icon = "💳"
	case "accounts":
		icon = "👤"
	}

	fmt.Printf("\n%s %s\n", icon, r.Name)
	fmt.Printf("   Requests:    %d (%.2f req/s)\n", m.TotalRequests, m.RequestsPerSec)
	fmt.Printf("   Success:     %.2f%%\n", m.SuccessRate)
	fmt.Printf("   Latency:     P50=%s  P95=%s  P99=%s\n", m.P50Latency, m.P95Latency, m.P99Latency)

	// Status codes
	codes := make([]string, 0, len(m.StatusCodes))
	for code := range m.StatusCodes {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	codeStr := ""
	for _, code := range codes {
		codeStr += fmt.Sprintf("%s:%d ", code, m.StatusCodes[code])
	}
	fmt.Printf("   Status:      %s\n", codeStr)
}
