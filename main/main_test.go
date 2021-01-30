package main

import (
	"context"
	"github.com/RinesThaix/homeTask/client"
	"github.com/RinesThaix/homeTask/server"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConsistencyParallel(t *testing.T) {
	consistencyParallel(t, 100, 20)
	consistencyParallel(t, 1_000_000, 40)
	consistencyParallel(t, 10_000_000, 20)
}

func consistencyParallel(t *testing.T, initialArraySize, clientsCount int) {
	t.Logf("starting test with %d clients and %d array elements", clientsCount, initialArraySize)
	srv := server.NewServer(initialArraySize)
	srv.Initialize()
	//t.Logf("srv array: %v", srv.Array())

	clients := createClients(srv, clientsCount)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	requests := int32(0)
	t.Logf("initializing clients")
	for _, client := range clients {
		if err := client.Initialize(); err != nil {
			t.Errorf("could not initialize client: %v", err)
		}
	}
	t.Logf("initialized clients")
	for i, client := range clients {
		j := i
		c := client
		random := rand.New(rand.NewSource(int64(j)))
		ticker := time.NewTicker(time.Millisecond * 200)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := performRandomModification(random, c); err != nil {
						t.Errorf("could not perform random modification on client %d: %v", j, err)
						return
					}
					atomic.AddInt32(&requests, 1)
				}
			}
		}()
	}

	for seconds := 1; seconds <= 15; seconds++ {
		time.Sleep(time.Second)
		t.Logf("[%02d:%02d] Handled %d requests", seconds / 60, seconds % 60, requests)
		if seconds == 10 {
			newClient := client.NewClient(srv)
			clients = append(clients, newClient)
			if err := newClient.Initialize(); err != nil {
				t.Errorf("could not initialize new client: %v", err)
				return
			}
		}
	}
	cancel()
	wg.Wait()
	time.Sleep(time.Second)

	//t.Logf("srv array (%d): %v", len(srv.Array()), srv.Array())
	//for i := 0; i < len(clients); i++ {
	//	t.Logf("%dth array (%d): %v", i, len(clients[i].Array()), clients[i].Array())
	//}

	t.Logf("checking values in the array between all the clients")
	size := clients[0].Size()
	for i := 0; i < size; i++ {
		value, err := clients[0].Get(i)
		if err != nil {
			t.Errorf("could not retrieve client 0 value for position %d: %v", i, err)
			return
		}
		for j := 1; j < len(clients); j++ {
			val, err := clients[j].Get(i)
			if err != nil {
				t.Errorf("could not retrieve client %d value for position %d: %v", j, i, err)
				return
			}
			if value != val {
				t.Errorf("mismatching values between clients 0 and %d on position %d: %d and %d", j, i, value, val)
				return
			}
		}
	}
	t.Logf("OK")
}

func createClients(srv *server.Server, size int) []*client.Client {
	result := make([]*client.Client, size)
	for i := 0; i < size; i++ {
		result[i] = client.NewClient(srv)
	}
	return result
}

func performRandomModification(rand *rand.Rand, client *client.Client) error {
	if client.Size() == 0 {
		return client.Insert(0, int32(rand.Uint32()))
	}
	switch rand.Intn(3) {
	case 0:
		return client.Insert(rand.Intn(client.Size() + 1), int32(rand.Uint32()))
	case 1:
		return client.Update(rand.Intn(client.Size()), int32(rand.Uint32()))
	default:
		return client.Delete(rand.Intn(client.Size()))
	}
}

func TestOperations(t *testing.T) {
	srv := server.NewServer(10)
	srv.Initialize()
	client := client.NewClient(srv)
	if err := client.Initialize(); err != nil {
		t.Errorf("could not initialize client: %v", err)
	}
	value, err := client.Get(9)
	if err != nil {
		t.Errorf("could not get client value at pos 9: %v", err)
	}
	if _, err = client.Get(10); err == nil {
		t.Errorf("by some reason could get value at pos 10, whilst the last applicable index is 9")
	}
	if err = client.Update(9, value - 1); err != nil {
		t.Errorf("could not update pos 9: %v", err)
	}
	newValue, err := client.Get(9)
	if err != nil {
		t.Errorf("could not get client value at pos 9: %v", err)
	}
	if newValue != value - 1 {
		t.Errorf("update failed")
	}
}
