package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/services/containerinstance/mgmt/2019-12-01/containerinstance"
	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"
	"github.com/tmeckel/aci-dns-updater/internal"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	ctx := context.Background()

	if len(os.Args) >= 2 && strings.EqualFold(os.Args[1], "version") {
		fmt.Fprintf(os.Stdout, "aci-dns-manager version: %s, commit: %s\n", version, commit)
		os.Exit(0)
	}

	var doDelete bool
	flag.BoolVar(&doDelete, "delete", false, "delete existing DNS A record")
	flag.Parse()

	if traceAz, err := internal.GetenvBool("ARM_TRACE", false); err == nil && traceAz {
		azlog.SetListener(func(c azlog.Classification, value string) {
			if traceAz {
				log.Println(value)
			}
		})
	}

	subscriptionID := internal.GetenvStr("ARM_SUBSCRIPTION_ID", "")
	if subscriptionID == "" {
		fmt.Fprint(os.Stderr, "Environment variable ARM_SUBSCRIPTION_ID is not defined or empty\n")
		os.Exit(1)
	}

	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{ExcludeAzureCLICredential: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to authenticate: %+v\n", err)
		os.Exit(1)
	}

	auth, err := internal.NewAzureManagementAuthorizer(ctx, cred)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Azure Management authorizer: %+v\n", err)
		os.Exit(1)
	}

	containerName := internal.GetenvStr("ACI_INSTANCE_NAME", "")
	if containerName == "" {
		containerName = internal.GetenvStr("Fabric_CodePackageName", "")
		if containerName == "" {
			fmt.Fprint(os.Stderr, "Unable to determine container name. Both variables ACI_INSTANCE_NAME, Fabric_CodePackageName are unset or empty\n")
			os.Exit(1)
		}
	}

	cic, err := internal.NewContainerInstanceClient(auth,
		internal.GetenvStr("ARM_SUBSCRIPTION_ID", ""),
		internal.GetenvStr("ACI_RESOURCE_GROUP_NAME", ""),
		containerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to container instance: %+v\n", err)
		os.Exit(1)
	}

	var container *containerinstance.ContainerGroup
	maxRetry, _ := internal.GetenvIntRange("ACI_MAX_RETRY", 4, 1, math.MaxInt)
	timeout, _ := internal.GetenvIntRange("ACI_TIMEOUT", 5, 1, math.MaxInt)

	for i := 0; i < maxRetry; i++ {
		_container, err := cic.Get(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get to container instance: %+v\n", err)
		} else if _container.ContainerGroupProperties != nil && _container.ContainerGroupProperties.IPAddress != nil && _container.ContainerGroupProperties.IPAddress.IP != nil {
			container = _container

			break
		}
		time.Sleep(time.Duration(timeout) * time.Second)
	}

	if container == nil {
		fmt.Fprintf(os.Stderr, "Failed to get IP address for container instance")
		os.Exit(1)
	}

	log.Printf("Container instance with id %s has ipv4 address %s\n", *container.ID, *container.ContainerGroupProperties.IPAddress.IP)

	dnsZoneSubscriptionID := os.Getenv("DNS_ZONE_RESOURCE_SUBSCRIPTION_ID")
	if dnsZoneSubscriptionID == "" {
		dnsZoneSubscriptionID = os.Getenv("ARM_SUBSCRIPTION_ID")
	}
	dnsZoneResourceGroup := os.Getenv("DNS_ZONE_RESOURCE_GROUP_NAME")
	if dnsZoneResourceGroup == "" {
		dnsZoneResourceGroup = os.Getenv("ACI_RESOURCE_GROUP_NAME")
	}
	dnsRecordsClient, err := internal.NewPrivateDNSRecordSetsClient(auth, dnsZoneSubscriptionID, dnsZoneResourceGroup, os.Getenv("DNS_ZONE_NAME"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create private DNS record set client: %+v\n", err)
		os.Exit(1)
	}

	aRecordName := os.Getenv("DNS_A_RECORD_NAME")
	if aRecordName == "" {
		aRecordName = *container.Name
	}

	if doDelete {
		err := dnsRecordsClient.Delete(ctx, privatedns.A, aRecordName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete private DNS A record: %+v\n", err)
			os.Exit(1)
		}

		log.Printf("Successfully deleted A record in zone %s\n", os.Getenv("DNS_ZONE_NAME"))
	} else {
		ttl, _ := internal.GetenvInt64Range("DNS_A_RECORD_TTL", 3600, 1, math.MaxInt64)

		record, err := dnsRecordsClient.CreateOrUpdate(ctx,
			privatedns.A,
			aRecordName,
			privatedns.RecordSet{
				RecordSetProperties: &privatedns.RecordSetProperties{
					TTL: &ttl,
					ARecords: &[]privatedns.ARecord{
						{
							Ipv4Address: container.ContainerGroupProperties.IPAddress.IP,
						},
					},
				},
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create private DNS A record: %+v\n", err)
			os.Exit(1)
		}

		log.Printf("Successfully created A record with ETag %s, FQDN %s and TTL %d\n", *record.Etag, *record.RecordSetProperties.Fqdn, *record.RecordSetProperties.TTL)
	}
}
