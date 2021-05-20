package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type Customer struct {
	CustomerID int
	Email      string
}

func email() string {
	name := words[rand.Intn(len(words))]
	domain := words[rand.Intn(len(words))]
	tld := words[rand.Intn(len(words))]

	return fmt.Sprintf("%s@%s.%s", name, domain, tld)
}

func insertData(ctx context.Context, db *sql.DB) {
	buf := bytes.NewBuffer(nil)

	doInsert := func() (int64, error) {
		defer buf.Reset()

		buf.WriteString("INSERT INTO customer (customer_id, email) VALUES ")

		binds := make([]interface{}, 0, *batchSize*2)
		for i := 0; i < *batchSize; i++ {
			c := &Customer{
				CustomerID: rand.Int(),
				Email:      email(),
			}
			buf.WriteString("(?, ?)")
			binds = append(binds, c.CustomerID, c.Email)
			if i < *batchSize-1 {
				buf.WriteString(", ")
			}
		}

		buf.WriteString(" ON DUPLICATE KEY UPDATE email=VALUES(email);")
		q := buf.String()

		if *debug {
			log.Printf("Executing query: %q", q)
		}

		r, err := db.Exec(q, binds...)
		if err != nil {
			log.Printf("query execute error: %s", err)
			return -1, err
		}

		if *debug {
			log.Printf("%+v", r)
		}

		return r.RowsAffected()
	}

	var (
		rowsAffected int64
		err          error
	)

	for {
		select {
		case <-ctx.Done():
			log.Printf("context finished: %v, stopping insert thread", ctx.Err())
			return
		default:
			rowsAffected, err = doInsert()
			switch err {
			case nil:
				Stats.RowsInserted.Inc(int(rowsAffected))
			default:
				Stats.Errors.Inc(1)
			}
		}

		delay := *sleep + 0 // TODO: add jitter
		if *debug {
			msg := "sleeping for %s between insert batches"
			params := []interface{}{delay}
			if err == nil {
				msg = "rows affected: %d; " + msg
				params = append([]interface{}{rowsAffected}, params...)
			}

			log.Printf(msg, params...)
		}
		time.Sleep(delay)
	}
}
