package internal

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/privatedns/mgmt/privatedns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/pkg/errors"
)

type PrivateDNSRecordSetsClient struct {
	client            *privatedns.RecordSetsClient
	resourceGroupName string
	privateZoneName   string
}

func NewPrivateDNSRecordSetsClient(auth autorest.Authorizer, subscriptionID, resourceGroupName, privateZoneName string) (*PrivateDNSRecordSetsClient, error) {
	if auth == nil {
		return nil, fmt.Errorf("parameter auth is nil")
	}
	if subscriptionID == "" {
		return nil, fmt.Errorf("parameter subscriptionID is empty")
	}
	if resourceGroupName == "" {
		return nil, fmt.Errorf("parameter resourceGroupName is empty")
	}
	if privateZoneName == "" {
		return nil, fmt.Errorf("parameter privateZoneName is empty")
	}

	privateDNSClient := privatedns.NewRecordSetsClient(subscriptionID)
	privateDNSClient.Authorizer = auth

	return &PrivateDNSRecordSetsClient{
		client:            &privateDNSClient,
		resourceGroupName: resourceGroupName,
		privateZoneName:   privateZoneName,
	}, nil
}

func (c *PrivateDNSRecordSetsClient) CreateOrUpdate(ctx context.Context, recordType privatedns.RecordType, relativeRecordSetName string, parameters privatedns.RecordSet) (privatedns.RecordSet, error) {
	recordSet, err := c.client.CreateOrUpdate(ctx,
		c.resourceGroupName,
		c.privateZoneName,
		recordType,
		relativeRecordSetName,
		parameters,
		"",
		"")
	if err != nil {
		return recordSet, errors.Wrap(err, "failed to create or update record set")
	}

	return recordSet, nil
}

func (c *PrivateDNSRecordSetsClient) Get(ctx context.Context, recordType privatedns.RecordType, relativeRecordSetName string) (privatedns.RecordSet, error) {
	recordSet, err := c.client.Get(ctx,
		c.resourceGroupName,
		c.privateZoneName,
		recordType,
		relativeRecordSetName)
	if err != nil {
		return recordSet, errors.Wrap(err, "failed to get record set")
	}

	return recordSet, nil
}

func (c *PrivateDNSRecordSetsClient) Delete(ctx context.Context, recordType privatedns.RecordType, relativeRecordSetName string) error {
	_, err := c.client.Delete(ctx,
		c.resourceGroupName,
		c.privateZoneName,
		recordType,
		relativeRecordSetName,
		"")
	if err != nil {
		return errors.Wrap(err, "failed to delete record set")
	}

	return nil
}
