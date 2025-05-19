package main

import (
	"fmt"       // sys print
	"math/rand" // generate random strings
	"sync"      // locks
	"time"      // latency calculation
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString() string {
	b := make([]byte, rand.Intn(100)) // different legnths from 0 - 99
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type transaction struct {
	id        string // generate a random string
	startTime time.Time
}

func newTransaction() transaction {
	return transaction{
		id:        randomString(),
		startTime: time.Now(),
	}
}

type systemState struct {
	totalTransactions int
	//latencies          []time.Duration
	transactionStorage map[string]time.Duration // [transaction.id, transaction.latency]
}

func main() {
	transactionChannel := make(chan transaction, 100) // buffer amount should depend on TPS
	systemState := systemState{
		transactionStorage: make(map[string]time.Duration),
	}
	var mutex sync.Mutex
	var wg sync.WaitGroup

	wg.Add(1)
	// thread to calculate latency
	go func() {
		//call .Done no matter what happens to goroutine
		defer wg.Done() // placed at the beginning to keep it safe from early return & panic
		for t := range transactionChannel {
			latency := time.Since(t.startTime)
			// add to storage
			mutex.Lock()
			systemState.transactionStorage[t.id] = latency
			systemState.totalTransactions++
			mutex.Unlock()

			fmt.Printf("Stored transaction with ID %s and latency of %s\n", t.id, latency)
		}
		fmt.Printf("Total transactions in system: %d\n", systemState.totalTransactions)

	}()

	wg.Add(1)
	// TPS calculation thread
	go func() {
		defer wg.Done()
		// timers = countdown, so use tickers
		// tickers for repeatedly events at reg intervals in future
		ticker := time.NewTicker(1 * time.Second) // bc tpS
		defer ticker.Stop()                       // calls .Stop no matter what
		var lock sync.Mutex
		var lastTotal int

		// tps = current - last
		for range ticker.C {
			lock.Lock()
			tps := (systemState.totalTransactions - lastTotal)
			lastTotal = systemState.totalTransactions
			lock.Unlock()

			fmt.Printf("TPS %d\n", tps)
			if tps == 0 {
				break
			}
		}
	}()

	//simulating transaction enterring the channel using threads
	var senderWg sync.WaitGroup
	for i := 0; i < 50; i++ {
		senderWg.Add(1)
		var t transaction
		t = newTransaction()

		go func() {
			defer senderWg.Done()
			transactionChannel <- t
			fmt.Printf("Transaction in channel with ID: %s\n", t.id)
		}()
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond) // random simulating "wait time" bc transactions were saved within milliseconds making TPS = max range
	}

	senderWg.Wait()
	// After sending all transactions
	close(transactionChannel) // let goroutine know there's no more input
	// time.Sleep(time.Second) ///// blindly delaying the goroutines is so bad use waitgroup instead
	wg.Wait()
}
