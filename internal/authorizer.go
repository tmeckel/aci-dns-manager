package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest"
	"github.com/pkg/errors"
)

type Authorizer struct {
	ctx        context.Context
	credential *azidentity.ChainedTokenCredential
	scopes     []string
	token      *azcore.AccessToken
}

func NewAzureManagementAuthorizer(ctx context.Context, credential *azidentity.ChainedTokenCredential) (*Authorizer, error) {
	return NewAuthorizer(ctx, credential, &[]string{
		"https://management.azure.com/.default",
	})
}

func NewAuthorizer(ctx context.Context, credential *azidentity.ChainedTokenCredential, scopes *[]string) (*Authorizer, error) {
	if credential == nil {
		return nil, fmt.Errorf("parameter credential is nil")
	}
	if scopes == nil || len(*scopes) == 0 {
		return nil, fmt.Errorf("parameter scopes is nil or empty")
	}

	return &Authorizer{
		ctx:        ctx,
		credential: credential,
		scopes:     *scopes,
		token:      nil,
	}, nil
}

func (a *Authorizer) WithAuthorization() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(request *http.Request) (*http.Request, error) {
			if a.token == nil || a.token.ExpiresOn.Before(time.Now()) {
				accToken, err := a.credential.GetToken(a.ctx, policy.TokenRequestOptions{
					Scopes: a.scopes,
				})
				if err != nil {
					return nil, errors.Wrap(err, "failed to get access token")
				}
				a.token = accToken
			}

			request, err := autorest.Prepare(request, autorest.WithBearerAuthorization(a.token.Token))
			if err != nil {
				return nil, errors.Wrap(err, "failed to create authorization header")
			}

			return request, nil
		})
	}
}
