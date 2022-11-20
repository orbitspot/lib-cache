# Redis GO 

Biblioteca para utilizar cache pelo Redis em seus projetos GO.

Possui estratégia para lidar com Multiplas conexões para servidores/bancos diferentes! 

![Biblioteca para cache em GO](https://assets.orbitspot.com/git/redis-golang.png)

## Como utilizar essa biblioteca em seu projeto

No diretório do seu projeto em GO, obtendo a versão mais recente disponível. Exemplo:
```
go get github.com/orbitspot/lib-cache
```

1 - Dentro de seu arquivo `.env`, adicione as variáveis de ambiente abaixo. Exemplo:
```
ENVIRONMENT=DEV
LOG_LEVEL=DEBUG
APP_NAME=mytest
SENTRY_DSN=

# REDIS_CONNECTION_X=host,port,db,expiration(seconds),name(for connection)
REDIS_CONNECTION_0=localhost,6379,15,  360,default  # DEFAULT connection, 10 minutes TTL
REDIS_CONNECTION_1=localhost,6379,05,    0,store    # Infinite TTL
REDIS_CONNECTION_2=localhost,6379,12,   10,price    # 10 seconds TTL
REDIS_CONNECTION_3=localhost,6379,07,86400,banner   # 1 day TTL
```
**Atencão: Solicitar para o time de DEVOPS preparar as variáveis para producão e configurar o pipeline para o seu projeto**

2 - Para facilitar sua vida, inclua em seu arquivo global de constantes os nomes das **conexões** definidas em seu arquivo `.env`. Exemplo:
```
const (
	redis_default = "default"
	redis_store   = "store"
	redis_price   = "price"
	redis_banner  = "banner"
)
```

3 - Dentro de seu código no arquivo `main.go`, importar a dependência e iniciar o Redis. 

Veja o exemplo/tutorial completo abaixo:
```
import (
    "github.com/orbitspot/lib-cache/pkg/cache"
    "github.com/orbitspot/lib-metrics/pkg/log"
)    

const (
	redis_default = "default"
	redis_store   = "store"
	redis_price   = "price"
	redis_banner  = "banner"
)

func main(){
	// Init logging
	log.Init()

	// Init Cache module, check requirements for `.env`
	cache.Init()

	// Health Check for Default Connection
	if err := cache.Ping(); err != nil {
		log.Error(err)
		panic("Redis Offline")
	}

	// Health Check for a specific Connection
	if err := cache.R[redis_store].Ping(); err != nil {
		log.Error(err)
		panic("Redis Offline")
	}

	// Basic definitions just for this example
	type myStruct struct {
		Code int
		Description string
	}

	// Using DEFAULT database - Single Database Architecture
	
	log.Boxed(log.LInfo,"SINGLE DATABASE ARCHITECTURE - SET/GET a few K/V using DEFAULT connection (with custom TTL example)")

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
	err, found =  cache.Get("key1", &returnedValue1)
	err, found =  cache.Get("key2", &returnedValue2)
	err, found =  cache.Get("key3", &returnedStruct)
	log.Info("Returned values (DELETED BY PATTERN 'key*'): [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)


	// Using Multi Databases Architecture

	log.Boxed(log.LInfo,"MULTI DATABASES ARCHITECTURE - SET/GET a few K/V using STORE connection")
	// Initialize variables and Set in Redis by Key
	initialValue1 = "my-value-2"
	initialValue2 = 7777777
	initialStruct = &myStruct{Code: 5678, Description: "Just Other Test of Structs"}
	err = cache.R[redis_store].Set("key1", &initialValue1)
	err = cache.R[redis_store].SetT("key2", &initialValue2, 60)
	err = cache.R[redis_store].Set("key3", &initialStruct)

	// Declare return variables and Get it in Redis by Key
	returnedValue1 = ""
	returnedValue2 = 0
	returnedStruct = &myStruct{}
	err, found =  cache.R[redis_store].Get("key1", &returnedValue1)
	err, found =  cache.R[redis_store].Get("key2", &returnedValue2)
	err, found =  cache.R[redis_store].Get("key3", &returnedStruct)
	log.Info("Returned values: [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)

	// Example for deleting Keys BY PATTERN
	err = cache.R[redis_store].Del("key*") // Prefix 'key'
	returnedValue2 = 0
	if err, found = cache.R[redis_store].Get("key2", &returnedValue2); err != nil || !found {
		log.Info("DELETE 'key2' - Key Deleted!")
	}
	log.Info("Returned values (DELETE key2): [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)

	// Example for deleting Keys
	err = cache.R[redis_store].Del("key*")
	returnedValue1 = ""
	returnedValue2 = 0
	returnedStruct = &myStruct{}
	err, found =  cache.R[redis_store].Get("key1", &returnedValue1)
	err, found =  cache.R[redis_store].Get("key2", &returnedValue2)
	err, found =  cache.R[redis_store].Get("key3", &returnedStruct)
	log.Info("Returned values (DELETED BY PATTERN 'key*'): [key1: %s, key2: %v, key3: %+v]", returnedValue1, returnedValue2, returnedStruct)

	log.Boxed(log.LInfo,"Tutorial for Finished!")
} 	
```

Resultado esperado do Terminal:
```
2022/06/17 08:51:52 WARN : Sentry WAS NOT STARTED because there is no config >> sentry.go:25
2022/06/17 08:51:52 INFO : Logging STARTED! [app: mytest] >> logger.go:48
2022/06/17 08:51:52 INFO : Redis Connection 'default' STARTED! [[app: mytest, connection: [number: 0, details: localhost,6379,15,  360,default]] >> redis.go:121
2022/06/17 08:51:52 INFO : Redis Connection 'store' STARTED! [[app: mytest, connection: [number: 0, details: localhost,6379,05,    0,store]] >> redis.go:121
2022/06/17 08:51:52 INFO : Redis Connection 'price' STARTED! [[app: mytest, connection: [number: 0, details: localhost,6379,12,   10,price]] >> redis.go:121
2022/06/17 08:51:52 INFO : Redis Connection 'banner' STARTED! [[app: mytest, connection: [number: 0, details: localhost,6379,07,86400,banner]] >> redis.go:121
2022/06/17 08:51:52 INFO : 
2022/06/17 08:51:52 INFO : ***********************************************************************************************************
2022/06/17 08:51:52 INFO : **  SINGLE DATABASE ARCHITECTURE - SET/GET a few K/V using DEFAULT connection (with custom TTL example)  **
2022/06/17 08:51:52 INFO : ***********************************************************************************************************
2022/06/17 08:51:52 INFO : 
2022/06/17 08:51:52 INFO : Checking if 'my-key-x' exists [value: '', found: false, error: <nil>] >> main.go:48
2022/06/17 08:51:52 INFO : Returned values: [key1: my-value-1, key2: 123456, key3: &{Code:1234 Description:Test of Structs}] >> main.go:67
2022/06/17 08:51:52 DEBUG: Deleting key [key2] >> redis.go:188
2022/06/17 08:51:52 INFO : DELETE 'key2' - Key Deleted! >> main.go:73
2022/06/17 08:51:52 INFO : Returned values (DELETE key2): [key1: my-value-1, key2: 0, key3: &{Code:1234 Description:Test of Structs}] >> main.go:75
2022/06/17 08:51:52 DEBUG: Deleting key [key3] >> redis.go:179
2022/06/17 08:51:52 DEBUG: Deleting key [key1] >> redis.go:179
2022/06/17 08:51:52 INFO : Returned values (DELETED BY PATTERN 'key*'): [key1: , key2: 0, key3: &{Code:0 Description:}] >> main.go:85
2022/06/17 08:51:52 INFO : 
2022/06/17 08:51:52 INFO : *******************************************************************************
2022/06/17 08:51:52 INFO : **  MULTI DATABASES ARCHITECTURE - SET/GET a few K/V using STORE connection  **
2022/06/17 08:51:52 INFO : *******************************************************************************
2022/06/17 08:51:52 INFO : 
2022/06/17 08:51:52 INFO : Returned values: [key1: my-value-2, key2: 7777777, key3: &{Code:5678 Description:Just Other Test of Structs}] >> main.go:106
2022/06/17 08:51:52 DEBUG: Deleting key [key3] >> redis.go:179
2022/06/17 08:51:52 DEBUG: Deleting key [key1] >> redis.go:179
2022/06/17 08:51:52 DEBUG: Deleting key [key2] >> redis.go:179
2022/06/17 08:51:52 INFO : DELETE 'key2' - Key Deleted! >> main.go:112
2022/06/17 08:51:52 INFO : Returned values (DELETE key2): [key1: my-value-2, key2: 0, key3: &{Code:5678 Description:Just Other Test of Structs}] >> main.go:114
2022/06/17 08:51:52 INFO : Returned values (DELETED BY PATTERN 'key*'): [key1: , key2: 0, key3: &{Code:0 Description:}] >> main.go:124
2022/06/17 08:51:52 INFO : 
2022/06/17 08:51:52 INFO : ******************************
2022/06/17 08:51:52 INFO : **  Tutorial for Finished!  **
2022/06/17 08:51:52 INFO : ******************************
```


## Notas para evolucão dessa lib

No momento utilizamos servidores Redis na versão 5.0.3
```
go get github.com/go-redis/redis
```

Para evolucões futuras, observar a correta versão da lib principal:

Redis 6:
```
go get github.com/go-redis/redis/v8
```

Redis 7:
```
go get github.com/go-redis/redis/v9
```

