package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cloudflare/cloudflare-go"
	"github.com/majanojoel/jmddns/internal/externip"
	"github.com/majanojoel/jmddns/internal/reconciler"
)

func main() {
	log.Println("jmddns: Service was started")
	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	if zoneID == "" {
		log.Fatal("CLOUDFLARE_ZONE_ID must be set")
	}
	externalIPProvider, err := externip.NewHTTPExternalIPProvider()
	if err != nil {
		log.Fatalln(err)
	}
	reconciler, err := reconciler.NewDNSRecordReconciler(externalIPProvider, api, zoneID)
	if err != nil {
		log.Fatalln(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := reconciler.RunIPReconcileLoop(); err != nil {
			log.Println("(*reconciler.DNSRecordReconciler).RunIPReconcileLoop exited due to error", err)
			return
		}
	}()
	// Setup the clean up functions.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigs
		reconciler.Stop()
	}()
	wg.Wait()
	log.Println("jmddns: Service was stopped")
}
