package internal

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/privatedns/mgmt/privatedns"
	"github.com/Azure/go-autorest/autorest"
)

type PrivateDnsRecordSetsClient struct {
	client            *privatedns.RecordSetsClient
	resourceGroupName string
	privateZoneName   string
}

func NewPrivateDnsRecordSetsClient(auth autorest.Authorizer, subscriptionId string, resourceGroupName string, privateZoneName string) (*PrivateDnsRecordSetsClient, error) {
	if auth == nil {
		return nil, fmt.Errorf("parameter auth is nil")
	}
	if subscriptionId == "" {
		return nil, fmt.Errorf("parameter subscriptionId is empty")
	}
	if resourceGroupName == "" {
		return nil, fmt.Errorf("parameter resourceGroupName is empty")
	}
	if privateZoneName == "" {
		return nil, fmt.Errorf("parameter privateZoneName is empty")
	}

	privateDnsClient := privatedns.NewRecordSetsClient(subscriptionId)
	privateDnsClient.Authorizer = auth
	return &PrivateDnsRecordSetsClient{
		client:            &privateDnsClient,
		resourceGroupName: resourceGroupName,
		privateZoneName:   privateZoneName,
	}, nil
}

func (c *PrivateDnsRecordSetsClient) CreateOrUpdate(ctx context.Context, recordType privatedns.RecordType, relativeRecordSetName string, parameters privatedns.RecordSet) (privatedns.RecordSet, error) {
	return c.client.CreateOrUpdate(ctx,
		c.resourceGroupName,
		c.privateZoneName,
		recordType,
		relativeRecordSetName,
		parameters,
		"",
		"")
}

func (c *PrivateDnsRecordSetsClient) Get(ctx context.Context, recordType privatedns.RecordType, relativeRecordSetName string) (privatedns.RecordSet, error) {
	return c.client.Get(ctx,
		c.resourceGroupName,
		c.privateZoneName,
		recordType,
		relativeRecordSetName)
}

func (c *PrivateDnsRecordSetsClient) Delete(ctx context.Context, recordType privatedns.RecordType, relativeRecordSetName string) error {
	_, err := c.client.Delete(ctx,
		c.resourceGroupName,
		c.privateZoneName,
		recordType,
		relativeRecordSetName,
		"")

	return err
}
