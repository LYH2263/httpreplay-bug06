package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/LYH2263/go-httpreplay"
	"github.com/LYH2263/go-httpreplay/studio"
)

func main() {
	addr := flag.String("addr", ":8223", "listen address")
	store := flag.String("store", "", "cassette store path")
	name := flag.String("name", "default", "cassette name")
	flag.Parse()
	opts := httpreplay.Options{Name: *name}
	if *store != "" {
		opts.StorePath = *store
	}
	c, err := httpreplay.OpenCassette(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	api := &studio.API{Cassette: c}
	srv := studio.New(api)
	log.Printf("httpreplay studio on %s", *addr)
	if err := http.ListenAndServe(*addr, srv.Handler); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
