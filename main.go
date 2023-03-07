package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/orbitspot/lib-cache/pkg/cache"
	"github.com/orbitspot/lib-metrics/pkg/log"
)

const (
	redis_default = "default"
	redis_test_a  = "test_a"
	redis_test_b  = "test_b"
	redis_test_c  = "test_c"
)

func main() {
	// Init logging
	log.Init()

	// Init Cache module, check requirements for `.env`
	cache.Init()

	// Health Check for Default Connection
	if err := cache.Ping(); err != nil {
		log.Fatalf(err, "Redis Offline")
	}

	// Health Check for a specific Connection
	if err := cache.R[redis_test_a].Ping(); err != nil {
		log.Fatalf(err, "Redis Offline")
	}

	// Basic definitions just for this example
	type myStruct struct {
		Code        int
		Description string
	}

	// Using DEFAULT database - Single Database Architecture

	log.Boxed(log.LInfo, "SINGLE DATABASE ARCHITECTURE - SET/GET a few K/V using DEFAULT connection (with custom TTL example)")

	// Simple example for Check & Handle if a value exists in cache
	var err error
	var found bool
	var myValue string
	if err, found := cache.Get("my-key-x", &myValue); err != nil || !found {
		log.Info("Checking if 'my-key-x' exists [value: '%s', found: %v, error: %v]", myValue, found, err)
	}

	// Initialize variables and Set in Redis by Key
	initialValue1 := "my-value-1"
	initialValue2 := 123456
	initialStruct := &myStruct{Code: 1234, Description: "Test of Structs"}
	err = cache.Set("key1", &initialValue1)
	err = cache.SetT("key2", &initialValue2, 60)
	err = cache.Set("key3", &initialStruct)

	// Declare return variables and Get it in Redis by Key
	var returnedValue1 string
	var returnedValue2 int
	var returnedStruct = &myStruct{}
	err, found = cache.Get("key1", &returnedValue1)
	err, found = cache.Get("key2", &returnedValue2)
	err, found = cache.Get("key3", &returnedStruct)
	log.Info("Returned values: [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)

	// Example for deleting Keys
	err = cache.Del("key2")
	returnedValue2 = 0
	if err, found = cache.Get("key2", &returnedValue2); err != nil || !found {
		log.Info("DELETE 'key2' - Key Deleted!")
	}
	log.Info("Returned values (DELETE key2): [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)

	// Example for deleting Keys BY PATTERN
	err = cache.Del("key*") // Prefix 'key'
	returnedValue1 = ""
	returnedValue2 = 0
	returnedStruct = &myStruct{}
	err, found = cache.Get("key1", &returnedValue1)
	err, found = cache.Get("key2", &returnedValue2)
	err, found = cache.Get("key3", &returnedStruct)
	log.Info("Returned values (DELETED BY PATTERN 'key*'): [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)

	// Using Multi Databases Architecture

	log.Boxed(log.LInfo, "MULTI DATABASES ARCHITECTURE - SET/GET a few K/V using test_a connection")
	// Initialize variables and Set in Redis by Key
	initialValue1 = "my-value-2"
	initialValue2 = 7777777
	initialStruct = &myStruct{Code: 5678, Description: "Just Other Test of Structs"}
	err = cache.R[redis_test_a].Set("key1", &initialValue1)
	err = cache.R[redis_test_a].SetT("key2", &initialValue2, 60)
	err = cache.R[redis_test_a].Set("key3", &initialStruct)

	// Declare return variables and Get it in Redis by Key
	returnedValue1 = ""
	returnedValue2 = 0
	returnedStruct = &myStruct{}
	err, found = cache.R[redis_test_a].Get("key1", &returnedValue1)
	err, found = cache.R[redis_test_a].Get("key2", &returnedValue2)
	err, found = cache.R[redis_test_a].Get("key3", &returnedStruct)
	log.Info("Returned values: [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)

	// Example for deleting Keys BY PATTERN
	err = cache.R[redis_test_a].Del("key*") // Prefix 'key'
	returnedValue2 = 0
	if err, found = cache.R[redis_test_a].Get("key2", &returnedValue2); err != nil || !found {
		log.Info("DELETE 'key2' - Key Deleted!")
	}
	log.Info("Returned values (DELETE key2): [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)

	// Example for deleting Keys
	err = cache.R[redis_test_a].Del("key*")
	returnedValue1 = ""
	returnedValue2 = 0
	returnedStruct = &myStruct{}
	err, found = cache.R[redis_test_a].Get("key1", &returnedValue1)
	err, found = cache.R[redis_test_a].Get("key2", &returnedValue2)
	err, found = cache.R[redis_test_a].Get("key3", &returnedStruct)
	log.Info("Returned values (DELETED BY PATTERN 'key*'): [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)

	log.Boxed(log.LInfo, "Tutorial for Finished!")
}
