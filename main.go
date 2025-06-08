package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Result struct {
	StatusCode int
	Error      error
}

func worker(url string, requests int, wg *sync.WaitGroup, results chan<- Result) {
	defer wg.Done()
	for i := 0; i < requests; i++ {
		resp, err := http.Get(url)
		if err != nil {
			results <- Result{Error: err}
			continue
		}
		results <- Result{StatusCode: resp.StatusCode}
		resp.Body.Close()
	}
}

func main() {
	// Parse flags
	url := flag.String("url", "", "URL do serviço a ser testado")
	requests := flag.Int("requests", 1, "Número total de requests")
	concurrency := flag.Int("concurrency", 1, "Número de chamadas simultâneas")
	flag.Parse()

	if *url == "" || *requests < 1 || *concurrency < 1 {
		fmt.Println("Uso: --url=<url> --requests=<total> --concurrency=<n>")
		os.Exit(1)
	}

	fmt.Printf("Iniciando teste de carga em %s com %d requisições e %d concorrências\n", *url, *requests, *concurrency)

	// Execução
	start := time.Now()
	results := make(chan Result, *requests)
	var wg sync.WaitGroup

	reqsPerWorker := *requests / *concurrency
	extra := *requests % *concurrency

	for i := 0; i < *concurrency; i++ {
		n := reqsPerWorker
		if i < extra {
			n++
		}
		wg.Add(1)
		go worker(*url, n, &wg, results)
	}

	wg.Wait()
	close(results)
	duration := time.Since(start)

	// Análise de resultados
	total := 0
	success := 0
	statusCount := make(map[int]int)
	failures := 0

	for r := range results {
		total++
		if r.Error != nil {
			failures++
			continue
		}
		if r.StatusCode == 200 {
			success++
		}
		statusCount[r.StatusCode]++
	}

	// Relatório
	fmt.Println("\n=== Relatório de Teste de Carga ===")
	fmt.Printf("Tempo total: %v\n", duration)
	fmt.Printf("Requests totais: %d\n", total)
	fmt.Printf("Sucesso (HTTP 200): %d\n", success)
	fmt.Printf("Falhas: %d\n", failures)
	fmt.Println("Distribuição de códigos HTTP:")
	for code, count := range statusCount {
		fmt.Printf("  %d: %d\n", code, count)
	}
}
