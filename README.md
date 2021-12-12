# Azure Container Instances DNS Record Manager

## Introduction

Azure Container Instances (ACI) are a great way to run Docker Images in a
managed environment with little configuration effort, without the need to deploy
a Container runtime like Docker or, even bigger, a Kubernetes Cluster. Despite the
ease-of-use ACI has some annoying issues, especially when bound to a private
virtual network.

One of these annoyances is the fact that an ACI instance, which is connected to
private virtual network, does not propagate a DNS name (FQDN). This is even more
problematic because ACI does not support a static IP address and thus an ACI
might be bound to a different IP address after a restart.

This repository contains a program (`aci-dns-manager`) which will ensure that a
DNS A-Record is created for an ACI instance, in a startup container, side car or
during the startup code (script) of the deployed container.

>
> Note
> Currently only Private DNS Zones are supported
>

## How to use

### Authentication

With ACI (theoretically) two types of principals can be used with
`aci-dns-manager` to authenticate towards the Azure REST API:

- Managed Identity
- Service Principal

Unfortunately ACI only supports Managed Identities only when the ACI is **not**
connected to a private virtual network.

Ref: https://docs.microsoft.com/en-us/azure/container-instances/container-instances-managed-identity#limitations

Due to this, the only way that `aci-dns-manager` can authenticate at the Azure
REST API is using a service principal. The following environment variables must
be specified to enable `aci-dns-manager` to use a service principal for
authentication:

- `AZURE_CLIENT_ID`
- `AZURE_CLIENT_SECRET`
- `AZURE_TENANT_ID`

For additional details refer to chapter [configuration](#configuration)

### Security considerations

Even if an own ACI deployment is trustworthy, the permissions of the
`aci-dns-manager` should be narrowed down to the permissions really required to
create (update, delete respectively) the DNS record for the connected ACI.

The ideal way to do this is to use a custom RBAC role definition in Azure, which
will then assigned to the principal the `aci-dns-manager` will use to authenticate
towards the Azure REST API.

An example of such a custom role definition could be the following:

```json
{
    "Name": "Private DNS A-Record Writer",
    "Description": "Can read, write DNS A and AAAA records",
    "AssignableScopes": [
        "/subscriptions/{subscriptionId1}"
    ],
    "Actions": [
        "Microsoft.Network/privateDnsZones/A/read",
        "Microsoft.Network/privateDnsZones/A/write",
        "Microsoft.Network/privateDnsZones/read",
        "Microsoft.Network/privateDnsZones/AAAA/read",
        "Microsoft.Network/privateDnsZones/AAAA/write",
        "Microsoft.Network/privateDnsOperationResults/read",
        "Microsoft.Network/privateDnsOperationStatuses/read"
    ],
    "NotActions": [],
    "DataActions": [],
    "NotDataActions": []
}
```

To allow the deletion of A and AAAA records, e.g. while deprovisioning an ACI, the
following `Actions` must be added to the above shown Role Definition

```text
"Microsoft.Network/privateDnsZones/A/delete",
"Microsoft.Network/privateDnsZones/AAAA/delete"
```

If the above Role is assigned to a Azure DNS Private Zone resource, the
`aci-dns-manager` could change, or even delete, arbitrary DNS records therein. To
mitigate this risk of inadvertently changing DNS entries by `aci-dns-manager`,
the above role should be assigned to a pre-created DNS entry. In that case
`aci-dns-manager` can only change it's "own" DNS record. What is definitely the
preferred way to go.

### Configuration

The program is configured via various environment variables

| Variable                            | Description                                                                                                                                                                                                                                                                                                                    | Required                                      |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------- |
| `ARM_SUBSCRIPTION_ID`               | The Azure subscription of the ACI                                                                                                                                                                                                                                                                                              | yes                                           |
| `ACI_RESOURCE_GROUP_NAME`           | The resource group in which the ACI is deployed                                                                                                                                                                                                                                                                                | yes                                           |
| `DNS_ZONE_NAME`                     | The name of the DNS zone in which the A record shall be created                                                                                                                                                                                                                                                                | yes                                           |
| `AZURE_TENANT_ID`                   | The id (guid) of the Azure Tenant in which the service principal is located used for authentication                                                                                                                                                                                                                            | no, required if: AZURE_CLIENT_ID is specified |
| `AZURE_CLIENT_ID`                   | The id (guid) of the service principal used for authentication                                                                                                                                                                                                                                                                 | no                                            |
| `AZURE_CLIENT_SECRET`               | The secret (password) of the service principal used for authentication                                                                                                                                                                                                                                                         | no, required if: AZURE_CLIENT_ID is specified |
| `ACI_INSTANCE_NAME`                 | The name of the ACI for which a a DNS A record shall be created. <br /> If not specified the environment variable `Fabric_CodePackageName` will be used. <br /> However, it was found that the ACI runtime does not reliably set this variable. So to prevent any errors the `ACI_INSTANCE_NAME` variable should always be set | no, if `Fabric_CodePackageName` is used       |
| `ACI_MAX_RETRY`                     | Number of retries to get deployment information (IP address) of the configured ACI                                                                                                                                                                                                                                             | no, default: `10`                             |
| `ACI_TIMEOUT`                       | How long to wait between retries                                                                                                                                                                                                                                                                                               | no, default: `10` seconds                     |
| `DNS_ZONE_RESOURCE_GROUP_NAME`      | The resource group of the private DNS Zone                                                                                                                                                                                                                                                                                     | no, default: `ACI_RESOURCE_GROUP_NAME`        |
| `DNS_ZONE_RESOURCE_SUBSCRIPTION_ID` | If the private DNS zone is located in a different subscription than the ACI instance, this variable contains the Id (guid) of this subscription.                                                                                                                                                                               | no, default `ARM_SUBSCRIPTION_ID`             |
| `DNS_A_RECORD_NAME`                 | The name of the DNS A record to create.                                                                                                                                                                                                                                                                                        | no, default `container name`                  |
| `DNS_A_RECORD_TTL`                  | The TTL value of the DNS A record                                                                                                                                                                                                                                                                                              | no, default `3600`                            |
| `ARM_TRACE`                         | If this environment variable is set (any value), trace messages will be created during communication with the Azure REST API. Valuable for troubleshooting.                                                                                                                                                                    | no                                            |

### Examples

The following examples are provided to show the usage of `aci-dns-manager`:

1. Scenario: `S6-Init`
    The `S6-Init` system will execute `aci-dns-manager` inside an init script.
    Details can be found in [README](example/s6/README.md)

2. Scenario: ACI Init Container
   `aci-dns-manager` will be run in an Init container defined in the YAML file which describes the ACI.
   Details can be found in [README](example/init/README.md)

## Resources

The repository will provider the following type of ways to download `aci-dns-manager`

- `Docker Image`
- `TAR archive`

Please refer to the [release section](releases) to find the latest versions.

## Links

https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#section-readme  
https://docs.microsoft.com/en-us/azure/developer/go/manage-resource-groups?tabs=bash%2Cazure-portal  
https://docs.microsoft.com/en-us/azure/developer/go/management-libraries#long-running-operations  
https://kreuzwerker.de/en/post/managing-multi-process-applications-in-containers-using-s6  
https://tutumcloud.wordpress.com/2015/05/20/s6-made-easy-with-the-s6-overlay/  
https://docs.microsoft.com/ms-my/azure/service-fabric/service-fabric-environment-variables-reference  
