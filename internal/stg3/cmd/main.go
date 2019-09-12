package main

import "github.com/fpawel/goanalit/stg3alit/stg3/internal/comportworker"

func main() {
	if err := comportworker.Run(); err != nil {
		panic(err)
	}
}
