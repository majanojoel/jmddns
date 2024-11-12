package reconciler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cloudflare/cloudflare-go"
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
		// TODO: This should be an interface, but for now just use the API directly.
		zoneID string
		cfAPI  *cloudflare.API
		quitCh chan struct{}
	}
)

func NewDNSRecordReconciler(ipProvider ExternalIPProvider, cfAPI *cloudflare.API, zoneID string) (*DNSRecordReconciler, error) {
	if ipProvider == nil {
		return nil, ErrNilArgument
	}
	r := &DNSRecordReconciler{
		ipProvider:      ipProvider,
		cfAPI:           cfAPI,
		quitCh:          make(chan struct{}),
		cachedIPAddress: "",
		zoneID:          zoneID,
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

type (
	dnsRecordInfo struct {
		name string
		id   string
	}
)

func (r *DNSRecordReconciler) handleReconcile() error {
	ctx := context.Background()
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
	rc := cloudflare.ZoneIdentifier(r.zoneID)
	params := cloudflare.ListDNSRecordsParams{}
	dnsRecords, _, err := r.cfAPI.ListDNSRecords(ctx, rc, params)
	if err != nil {
		return fmt.Errorf("(*cloudflare.API).ListDNSRecords: %w", err)
	}
	needToUpdateRecordInfos := make([]dnsRecordInfo, 0)
	// Compare dns record IP address to the current IP address
	for _, dnsr := range dnsRecords {
		// We will only reconcile A records.
		if dnsr.Type != "A" {
			continue
		}
		fmt.Println("DNSRecordReconciler: Handling record with ID", dnsr.ID)
		if dnsr.Content != ipAddress {
			info := dnsRecordInfo{
				name: dnsr.Name,
				id:   dnsr.ID,
			}
			needToUpdateRecordInfos = append(needToUpdateRecordInfos, info)
		}
	}
	log.Printf("DNSRecordReconciler: Found %d records to update\n", len(needToUpdateRecordInfos))
	// Make the changes to the DNS records
	allUpdated := true
	for _, dnsrToUpdate := range needToUpdateRecordInfos {
		comment := fmt.Sprintf("This record was updated by the jmddns service at %s",
			time.Now().UTC().String())
		rc := cloudflare.ZoneIdentifier(r.zoneID)
		updateParams := cloudflare.UpdateDNSRecordParams{
			Type:    "A",
			Name:    dnsrToUpdate.name,
			ID:      dnsrToUpdate.id,
			Content: ipAddress,
			Comment: &comment,
		}
		if _, err := r.cfAPI.UpdateDNSRecord(ctx, rc, updateParams); err != nil {
			log.Println("DNSRecordReconciler: Failed to update record", dnsrToUpdate)
			if !allUpdated {
				allUpdated = false
			}
			continue
		}
		log.Println("DNSRecordReconciler: Updated record with ID", dnsrToUpdate.id)
	}
	// Update our cached value ONLY if the DNS record was updated successfully.
	if allUpdated {
		r.cachedIPAddress = ipAddress
	}
	return nil
}

func (r *DNSRecordReconciler) Stop() {
	r.quitCh <- struct{}{}
}
