package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	debug = flag.Bool("debug", false, "")

	dsn       = flag.String("dsn", "vt_app@tcp(localhost:15306)/customer", "Full DSN spec: `[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]`")
	threads   = flag.Int("threads", 4, "")
	batchSize = flag.Int("batch_size", 100, "")
	sleep     = flag.Duration("sleep-interval", time.Millisecond*10, "sleep `sleep-interval` +/- 0.1*`sleep-interval` between insert batches")

	dictionary = flag.String("dictionary", "/usr/share/dict/words", "")
	words      []string
)

func main() {
	flag.Parse()

	if *threads < 1 {
		log.Fatalf("-threads must be > 0, got: %d", *threads)
	}

	if *batchSize < 1 {
		log.Fatalf("-batch_size must be > 0, got: %d", *batchSize)
	}

	data, err := ioutil.ReadFile(*dictionary)
	if err != nil {
		log.Fatal(err)
	}

	words = strings.Split(string(data), "\n")
	if words[len(words)-1] == "" {
		words = words[:len(words)-1]
	}

	if *debug {
		log.Printf("Loaded %d words from %s", len(words), *dictionary)
	}

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	signals := make(chan os.Signal, 8)
	signal.Notify(signals, os.Interrupt)

	shutdown := make(chan error)
	go func() {
		sig := <-signals
		shutdown <- fmt.Errorf("got signal %v, shutting down", sig)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *debug {
		log.Printf("spawning %d insert threads", *threads)
	}

	for i := 0; i < *threads; i++ {
		go func() {
			insertData(ctx, db)
		}()
	}

	err = <-shutdown
	if err != nil {
		log.Println(err)
	}
}
