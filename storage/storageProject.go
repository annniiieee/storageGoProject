package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
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
	endTime   time.Time
}

type systemState struct {
	totalTransactions  int
	transactionStorage map[int]transaction // [id, transaction]
}

func main() {
	transactionChannel := make(chan transaction, 20) // buffer amount should depend on TPS ?
	systemState := systemState{
		transactionStorage: make(map[int]transaction),
	}
	var mutex sync.Mutex
	var wg sync.WaitGroup

	//simulating transaction enterring the channel using threads
	var senderWg sync.WaitGroup
	for i := 0; i < 20; i++ {
		senderWg.Add(1)
		t := transaction{
			id:        randomString(),
			startTime: time.Now(),
		}

		go func(tr transaction) {
			defer senderWg.Done()
			transactionChannel <- tr
			fmt.Printf("Transaction in channel with ID: %s\n", tr.id)
		}(t)
	}

	wg.Add(1)
	// thread to calculate latency
	go func() {
		//call .Done no matter what happens to goroutine
		defer wg.Done() // placed at the beginning to keep it safe from early return & panic
		for t := range transactionChannel {
			// add to storage
			mutex.Lock()
			t.endTime = time.Now()
			systemState.transactionStorage[systemState.totalTransactions] = t
			systemState.totalTransactions++
			mutex.Unlock()

			fmt.Printf("Stored transaction with ID %s and latency of %s\n", t.id, t.endTime.Sub(t.startTime))
		}
		fmt.Printf("Total transactions in system: %d\n", systemState.totalTransactions)

	}()

	senderWg.Wait() // don't end, wait after waitgroups
	// After sending all transactions
	close(transactionChannel) // let goroutine know there's no more input

	wg.Wait()

	fmt.Println("All transactions with timestamps and latency:")
	var firstEnd time.Time
	var lastEnd time.Time
	first := true

	for _, t := range systemState.transactionStorage {
		latency := t.endTime.Sub(t.startTime)
		fmt.Printf("ID: %s | Start: %s | End: %s | Latency: %s\n",
			t.id,
			t.startTime.Format(time.RFC3339Nano),
			t.endTime.Format(time.RFC3339Nano),
			latency)

		if first {
			firstEnd = t.endTime
			lastEnd = t.endTime
			first = false
		} else {
			if t.endTime.Before(firstEnd) {
				firstEnd = t.endTime
			}
			if t.endTime.After(lastEnd) {
				lastEnd = t.endTime
			}
		}
	}

	// Calculate and print final TPS
	totalTime := lastEnd.Sub(firstEnd).Seconds()
	tps := float64(systemState.totalTransactions) / totalTime
	fmt.Printf("\nEffective TPS (Total: %d, Duration: %.4f seconds): %.2f\n",
		systemState.totalTransactions, totalTime, tps)

}
