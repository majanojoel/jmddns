package reconciler

import (
	"errors"
	"log"
	"time"
)

const (
	defaultReconcileInterval = 5 * time.Second
)

var (
	ErrNilArgument = errors.New("a nil argument was provided")
)

type (
	ExternalIPProvider interface {
		GetExternalIP() (string, error)
	}
	DNSRecordReconciler struct {
		ipProvider ExternalIPProvider
		quitCh     chan struct{}
	}
)

func NewDNSRecordReconciler(ipProvider ExternalIPProvider) (*DNSRecordReconciler, error) {
	if ipProvider == nil {
		return nil, ErrNilArgument
	}
	r := &DNSRecordReconciler{
		ipProvider: ipProvider,
		quitCh:     make(chan struct{}),
	}
	return r, nil
}

// RunIPReconcileLoop runs the reconciliation loop.
// Note: This is a blocking function.
func (r *DNSRecordReconciler) RunIPReconcileLoop() error {
	log.Println("DNSRecordReconciler: Starting the reconciliation loop")
	ticker := time.NewTicker(defaultReconcileInterval)
	for {
		select {
		case <-ticker.C:
			r.handleReconcile()
		case <-r.quitCh:
			log.Println("DNSRecordReconciler: Stopped by the quit channel")
			return nil
		}
	}
}

func (r *DNSRecordReconciler) handleReconcile() error {
	// Get the current DNS record values

	// Get the current IP address
	ipAddress, err := r.ipProvider.GetExternalIP()
	if err != nil {
		log.Printf("(ExternalIPProvider).GetExternalIP: %s\n", err.Error())
		// TODO: Return an error in fatal cases.
		return nil
	}
	log.Println("DNSRecordReconciler: Retrieved external IP address", ipAddress)
	// Compare dns record IP address to the current IP address
	// If not different, exit the loop.

	// Make the changes to the DNS records
	return nil
}

func (r *DNSRecordReconciler) Stop() {
	r.quitCh <- struct{}{}
}
