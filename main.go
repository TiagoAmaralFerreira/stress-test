package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

type Report struct {
	TotalRequests      int
	TotalTime          time.Duration
	SuccessCount       int
	StatusDistribution map[int]int
}

func main() {
	url := flag.String("url", "", "URL do serviço a ser testado")
	requests := flag.Int("requests", 100, "Número total de requests")
	concurrency := flag.Int("concurrency", 10, "Número de chamadas simultâneas")

	flag.Parse()

	if *url == "" {
		fmt.Println("Erro: URL é obrigatória")
		flag.Usage()
		return
	}

	fmt.Printf("Iniciando teste de carga:\n")
	fmt.Printf("URL: %s\n", *url)
	fmt.Printf("Total de Requests: %d\n", *requests)
	fmt.Printf("Concorrência: %d\n\n", *concurrency)

	startTime := time.Now()
	results := make(chan Result, *requests)
	var wg sync.WaitGroup

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go worker(*url, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	report := processResults(results)
	report.TotalTime = time.Since(startTime)
	printReport(report)
}

func worker(url string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{}

	for {
		start := time.Now()
		resp, err := client.Get(url)
		duration := time.Since(start)

		if err != nil {
			results <- Result{Error: err, Duration: duration}
			continue
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		results <- Result{StatusCode: resp.StatusCode, Duration: duration}
	}
}

func processResults(results <-chan Result) Report {
	report := Report{
		StatusDistribution: make(map[int]int),
	}

	for result := range results {
		report.TotalRequests++

		if result.Error != nil {
			report.StatusDistribution[0]++
			continue
		}

		report.StatusDistribution[result.StatusCode]++

		if result.StatusCode == http.StatusOK {
			report.SuccessCount++
		}
	}

	return report
}

func printReport(report Report) {
	fmt.Println("\nRelatório do Teste de Carga")
	fmt.Println("==========================")
	fmt.Printf("Tempo Total: %v\n", report.TotalTime)
	fmt.Printf("Total de Requests: %d\n", report.TotalRequests)
	fmt.Printf("Requests com Status 200: %d\n", report.SuccessCount)

	fmt.Println("\nDistribuição de Status HTTP:")
	for status, count := range report.StatusDistribution {
		if status == 0 {
			fmt.Printf("Erros: %d\n", count)
		} else {
			fmt.Printf("Status %d: %d\n", status, count)
		}
	}

	rps := float64(report.TotalRequests) / report.TotalTime.Seconds()
	fmt.Printf("\nRequests por Segundo: %.2f\n", rps)
}
