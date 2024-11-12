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
		cachedIPAddress string
		ipProvider      ExternalIPProvider
		quitCh          chan struct{}
	}
)

func NewDNSRecordReconciler(ipProvider ExternalIPProvider) (*DNSRecordReconciler, error) {
	if ipProvider == nil {
		return nil, ErrNilArgument
	}
	r := &DNSRecordReconciler{
		ipProvider:      ipProvider,
		quitCh:          make(chan struct{}),
		cachedIPAddress: "",
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
	// Get the current IP address
	ipAddress, err := r.ipProvider.GetExternalIP()
	if err != nil {
		log.Printf("(ExternalIPProvider).GetExternalIP: %s\n", err.Error())
		// TODO: Return an error in fatal cases.
		return nil
	}
	log.Println("DNSRecordReconciler: Retrieved external IP address", ipAddress)
	if ipAddress == r.cachedIPAddress {
		log.Println("DNSRecordReconciler: Cached IP address match, nothing to do.")
		return nil
	}
	// Get the current DNS record values
	// Compare dns record IP address to the current IP address

	// Make the changes to the DNS records
	// Update our cached value ONLY if the DNS record was updated successfully.
	r.cachedIPAddress = ipAddress
	return nil
}

func (r *DNSRecordReconciler) Stop() {
	r.quitCh <- struct{}{}
}
