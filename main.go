package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ignore this for now
// func connectToPostgresql() {
// 	fmt.Println("connecting to postgresql..")
// 	postgresInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "postgres", 5432, "user", "mypassword", "mydb")

// 	dbpool, err := pgxpool.New(context.Background(), postgresInfo)
// 	if err != nil {
// 		fmt.Printf("Unable to create connection pool: %v\n", err)
// 		os.Exit(1)
// 	}
// 	fmt.Printf("Unable to create connection pool: %v\n", dbpool)	

// 	defer dbpool.Close()
// }

const maxConnections = 5

var (
	pool	*pgxpool.Pool
	mu 		sync.RWMutex
)
func main() {
	// connectToPostgresql()

	poolConfig, err := pgxpool.ParseConfig("host=postgres port=5432 user=om password=password dbname=mydb sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}

	poolConfig.MaxConns = maxConnections

	pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal(err)
	}
	
	defer pool.Close()

	var wg sync.WaitGroup

	wg.Add(3)
	go connectToDB("myserver", &wg)
	go connectToDB("worker1", &wg)
	go connectToDB("worker2", &wg)

	wg.Wait()
}

func connectToDB(serviceName string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// Try to acquire a connection from the pool
		mu.RLock()
		conn, err := pool.Acquire(context.Background())
		mu.RUnlock()

		if err != nil {
			fmt.Printf("[%s] connection pool limit reached: %v\n", serviceName, err)
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("[%s] Successfully connected to DB\n", serviceName)

		// worker is doing some task for 3 seconds
		time.Sleep(3 * time.Second)

		// After task is complete destroy the connection
		conn.Release()
	}
}
