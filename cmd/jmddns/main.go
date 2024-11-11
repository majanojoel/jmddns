package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/majanojoel/jmddns/internal/externip"
	"github.com/majanojoel/jmddns/internal/reconciler"
)

func main() {
	log.Println("jmddns: Service was started")
	externalIPProvider, err := externip.NewHTTPExternalIPProvider()
	if err != nil {
		log.Fatalln(err)
	}
	reconciler, err := reconciler.NewDNSRecordReconciler(externalIPProvider)
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
