package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// CONTEXT WITH TIMEOUT

// --- Manually cancelled without context timeout ---

func operationWithoutTimeout(cancelChan <-chan bool) {
	fmt.Println("Simulate long running operation (10 seconds) that will be cancelled manually after 5 seconds running")
	for i := range 10 {
		select {
		case <-cancelChan:
			// If timeout signal received
			fmt.Printf("[%s] : Operation %d cancelled\n", time.Now().Format(time.RFC3339), i)
			return
		default:
			// Normal operation before timeout received
			fmt.Printf("[%s] : Operation %d running\n", time.Now().Format(time.RFC3339), i)
			time.Sleep(1 * time.Second) // Simulate sequence operation every 1 second
		}
	}
	fmt.Printf("[%s] : Simulation complete\n", time.Now().Format(time.RFC3339)) // This line will never be printed out
}

// --- Cancelled using context.WithTimeout() ---

func operationWithTimeout(ctx context.Context) {
	fmt.Println("\nSimulate long running operation (10 seconds) that will be cancelled via context.WithTimeout() in 5 seconds")
	for i := range 10 {
		select {
		case <-ctx.Done():
			// If timeout signal received
			return
		default:
			// Normal operation before timeout signal received
			fmt.Printf("[%s] : Operation %d running\n", time.Now().Format(time.RFC3339), i)
			time.Sleep(1 * time.Second) // Simulate sequence operation every 1 second
		}
	}
	fmt.Printf("[%s] : Simulation complete\n", time.Now().Format(time.RFC3339)) // This line will never be printed out
}

func simulateTimeout() {
	// Simulate without context timeout
	cancelChan := make(chan bool)
	go operationWithoutTimeout(cancelChan)
	time.Sleep(5 * time.Second) // Simulate 5 seconds running operation
	cancelChan <- true
	close(cancelChan) // Closing channel means nothing getting in/out after the last value is sent/received

	// Simulate with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Important to close all resources after cancel signal is sent
	go operationWithTimeout(ctx)
	<-ctx.Done()
}

// CONTEXT WITH CANCEL

// --- Without cancel, operation keeps running eventhough parameter is found ---

func operationWithoutCancel(wg *sync.WaitGroup, id int, resultChan chan<- int) {
	defer wg.Done()

	fmt.Printf("Worker %d start\n", id)
	keyFound := rand.Intn(5) + 1                                // To simulate random result found
	workDuration := time.Duration(rand.Intn(5)+1) * time.Second // To simulate random working time
	time.Sleep(workDuration)

	if keyFound == id {
		fmt.Printf("Worker %d found the key\n", id)
		resultChan <- id
	}
	fmt.Printf("Worker %d finish\n", id)
}

func simulateWithoutCancel() {
	fmt.Println("\nSimulate work without cancel")
	var wg sync.WaitGroup
	resultChan := make(chan int, 1)

	// Start goroutine
	for i := range 10 {
		wg.Add(1)
		go operationWithoutCancel(&wg, i, resultChan)
	}

	foundWorker := <-resultChan
	fmt.Printf("Got result from worker %d, other goroutine still running\n", foundWorker)

	wg.Wait()
	fmt.Println("Simulation finishes")
}

// --- With cancel, operation stops when parameter is found ---

func operationWithCancel(ctx context.Context, id int, resultChan chan<- int) {
	// We don't use wait group because once the cancel signal received we don't need to wait for other goroutine to finish their work
	fmt.Printf("Worker %d start\n", id)
	keyFound := rand.Intn(5) + 1                                // To simulate random result found
	workDuration := time.Duration(rand.Intn(5)+1) * time.Second // To simulate random working time

	select {
	case <-time.After(workDuration):
		// If cancel signal is not received during working time
		if keyFound == id {
			fmt.Printf("Worker %d found the key\n", id)
			resultChan <- id
		}
		fmt.Printf("Worker %d finish\n", id)
	case <-ctx.Done():
		// If cancel signal received before parameter found
		fmt.Printf("Worker %d cancelled\n", id)
	}
}

func simulateWithCancel() {
	fmt.Println("\nSimulate work with cancel")
	resultChan := make(chan int, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := range 10 {
		go operationWithCancel(ctx, i, resultChan)
	}

	foundWorker := <-resultChan
	fmt.Printf("Got result from worker %d, other goroutine cancelled\n", foundWorker)
	cancel()                           // Call cancel() to signal all other goroutine to stop, otherwise there will be deadlock
	time.Sleep(100 * time.Millisecond) // Give time to other goroutine to print out cancel message
	fmt.Println("Simulation finishes")
}

// CONTEXT WITH DEADLINE

// --- Without context deadline, normal operation runs without interuption ---

func operationWithoutDeadline() {
	fmt.Println("This operation will run exactly for 5 seconds without interuption")
	fmt.Printf("[%s] Operation starts\n", time.Now().Format(time.RFC3339))
	time.Sleep(5 * time.Second)
	fmt.Printf("[%s] Operation finishes\n", time.Now().Format(time.RFC3339))
}

func simulateWithoutDeadline() {
	fmt.Println("\nSimulate operation without context deadline")
	start := time.Now()
	operationWithoutDeadline()
	fmt.Printf("Elapsed time: %v\n", time.Since(start))
}

// --- With context deadline, we can define when the cancel signal will be activated ---

func operationWithDeadline(ctx context.Context) {
	fmt.Println("This operation is designed to run for 5 seconds, but will be interupted in 3 seconds")
	fmt.Printf("[%s] Operation starts\n", time.Now().Format(time.RFC3339))

	select {
	case <-time.After(5 * time.Second):
		// If cancel signal is not received
		fmt.Printf("[%s] Operation finishes\n", time.Now().Format(time.RFC3339))
	case <-ctx.Done():
		fmt.Printf("[%s] Operation cancelled: %v\n", time.Now().Format(time.RFC3339), ctx.Err())
	}
}

func simulateWithDeadline() {
	// Set the deadline where cancel signal will be activated
	deadline := time.Now().Add(3 * time.Second)
	fmt.Println("\nSimulate operation with context deadline")

	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	start := time.Now()
	operationWithDeadline(ctx)
	fmt.Printf("Elapsed time: %v\n", time.Since(start))
}

// CONTEXT WITH REQUEST-SCOPE VALUE

// --- Without context value, parameter will be passed as parameter argument ---

func operationWithoutValue(username, password string) {
	fmt.Println("\n[Without context] Start processing")

	// Manually passing argument
	validateData(username, password)
	saveData(username, password)

	fmt.Println("[Without context] Finish")
}

func validateData(username, password string) {
	fmt.Printf("[Without context] Username: %s, password: %s is valid\n", username, password)
}

func saveData(username, password string) {
	fmt.Printf("[Without context] Username: %s, password: %s is saved\n", username, password)
}

// --- With context value, we embed request value via context.WithValue() ---

type ctxKey string

func operationWithValue(ctx context.Context) {
	fmt.Println("\n[With context] Start processing")

	validateDataWithContext(ctx)
	saveDataWithContext(ctx)

	fmt.Println("[With context] Finish")
}

func validateDataWithContext(ctx context.Context) {
	username := ctx.Value(ctxKey("username")).(string)
	password := ctx.Value(ctxKey("password")).(string)
	fmt.Printf("[With context] Username: %s, password: %s is valid\n", username, password)
}

func saveDataWithContext(ctx context.Context) {
	username := ctx.Value(ctxKey("username")).(string)
	password := ctx.Value(ctxKey("password")).(string)
	fmt.Printf("[With context] Username: %s, password: %s is saved\n", username, password)
}

func main() {
	// == Context timeout ==
	// simulateTimeout()

	// == Context cancel ==
	// simulateWithoutCancel()
	// simulateWithCancel()

	// == Context deadline ==
	// simulateWithoutDeadline()
	// simulateWithDeadline()

	// == Context value ==
	username := "boy123"
	password := "password456"
	operationWithoutValue(username, password)

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxKey("username"), username)
	ctx = context.WithValue(ctx, ctxKey("password"), password)

	operationWithValue(ctx)

}
