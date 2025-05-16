package main

import (
	"fmt"       // sys print
	"math/rand" // generate random strings
	"sync"      // locks
	"time"      // latency calculation
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString() string {
	b := make([]byte, rand.Intn(100))
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type transaction struct {
	id        string // generate a random string
	startTime time.Time
	endTime   time.Time
	latency   time.Duration // (endTime.Sub(startTime))
}

func newTransaction() transaction {
	return transaction{
		id:        randomString(),
		startTime: time.Now(),
	}
}

func main() {
	transactionChannel := make(chan transaction, 100) // buffer amount should depend on TPS
	transactionStorage := make(map[string]transaction)
	var mutex sync.Mutex
	var wg sync.WaitGroup

	wg.Add(1)
	go func() { //anonymous function using goroutine
		defer wg.Done() // placed at the beginning to keep it safe from early return & panic
		for t := range transactionChannel {
			mutex.Lock()
			transactionStorage[t.id] = t
			t.endTime = time.Now()
			t.latency = t.endTime.Sub(t.startTime)
			mutex.Unlock()

			fmt.Printf("Stored transaction with ID %s and latency of %s\n", t.id, t.latency)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	//simulating transaction enterring the channel
	for i := 0; i < 10; i++ {
		var t transaction
		t = newTransaction()

		transactionChannel <- t
		fmt.Printf("Transaction in channel with ID: %s\n", t.id)

	}

	// After sending all transactions
	close(transactionChannel) // let goroutine know there's no more input
	// time.Sleep(time.Second) ///// blindly delaying the goroutines is so bad
	wg.Wait()

}
