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

func newTransaction() transaction {
	return transaction{
		id:        randomString(),
		startTime: time.Now(),
	}
}

type systemState struct {
	sync.Mutex
	totalTransactions  int
	transactionStorage map[int]transaction // [transaction, transaction.latency]
	tpsValues          []int
}

func main() {
	transactionChannel := make(chan transaction, 10) // buffer amount should depend on TPS ?
	systemState := systemState{
		transactionStorage: make(map[int]transaction),
	}
	var wg sync.WaitGroup

	wg.Add(1)
	// thread to calculate latency
	go func() {
		//call .Done no matter what happens to goroutine
		defer wg.Done() // placed at the beginning to keep it safe from early return & panic
		for t := range transactionChannel {
			systemState.Lock()
			t.endTime = time.Now()
			systemState.transactionStorage[systemState.totalTransactions] = t
			systemState.totalTransactions++
			systemState.Unlock()

			fmt.Printf("Stored transaction with ID %s and latency of %s\n", t.id, t.endTime.Sub(t.startTime))
		}
		fmt.Printf("Total transactions in system: %d\n", systemState.totalTransactions)

	}()

	wg.Add(1)
	// TPS calculation thread
	go func() {
		defer wg.Done()
		// timers = countdown, so use tickers
		// tickers for repeatedly events at reg intervals in future, 1 sec in our case
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop() // calls .Stop no matter what
		var lastTotal int

		// tps = current - last
		for range ticker.C {
			systemState.Lock()
			tps := (systemState.totalTransactions - lastTotal)
			lastTotal = systemState.totalTransactions
			systemState.tpsValues = append(systemState.tpsValues, tps)
			systemState.Unlock()

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

	senderWg.Wait() // don't end, wait after waitgroups
	// After sending all transactions
	close(transactionChannel) // let goroutine know there's no more input

	wg.Wait()

	fmt.Println("All latency values:")
	for _, val := range systemState.transactionStorage {
		latency := val.endTime.Sub(val.startTime)
		fmt.Printf("Transaction ID: %s, latency: %s\n", val.id, latency)
	}

	fmt.Println("All TPS values:")
	for i, val := range systemState.tpsValues {
		fmt.Printf("Second %d: TPS = %d\n", i+1, val)
	}

}
