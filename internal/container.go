package internal

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/containerinstance/mgmt/2019-12-01/containerinstance"
	"github.com/Azure/go-autorest/autorest"
	"github.com/pkg/errors"
)

type ContainerInstanceClient struct {
	client             *containerinstance.ContainerGroupsClient
	resourceGroupName  string
	containerGroupName string
}

func NewContainerInstanceClient(auth autorest.Authorizer, subscriptionID, resourceGroupName, containerGroupName string) (*ContainerInstanceClient, error) {
	if auth == nil {
		return nil, fmt.Errorf("parameter auth is nil")
	}
	if subscriptionID == "" {
		return nil, fmt.Errorf("parameter subscriptionId is empty")
	}
	if resourceGroupName == "" {
		return nil, fmt.Errorf("parameter resourceGroupName is empty")
	}
	if containerGroupName == "" {
		return nil, fmt.Errorf("parameter containerGroupName is empty")
	}
	groupsClient := containerinstance.NewContainerGroupsClient(subscriptionID)
	groupsClient.Authorizer = auth

	return &ContainerInstanceClient{
		client:             &groupsClient,
		resourceGroupName:  resourceGroupName,
		containerGroupName: containerGroupName,
	}, nil
}

func (ci *ContainerInstanceClient) Get(ctx context.Context) (*containerinstance.ContainerGroup, error) {
	val, err := ci.client.Get(ctx, ci.resourceGroupName, ci.containerGroupName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get container instance")
	}

	return &val, nil
}
