package iferr

import "log"

func Log(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func Log2(_ any, err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func Panic(err error) {
	if err != nil {
		panic(err)
	}
}
