# ACI DNS Manager: Deploy using an init container

## Introduction

>
> :warning: **Note:**
>
> Using `aci-dns-manager` in an ACI init container isn't working anymore because
> an IP address will be assigned to the ACI instance only **after all** init containers
> finished successfully. Refer to the [S6](../s6/README.md) example for an alternative.
>

This example shows how to configure `aci-dns-manager` as an ACI init container
to manage a DNS A-Record for the deployed ACI (main) container bound to an
virtual private network.

Using an init container is ideal, if you have a pre-configured Docker Image that
is running in an ACI and you don't want to change the Docker Image to install
`aci-dns-manager` in it. It's the easiest way to utilize `aci-dns-manager` to
have a DNS record for your ACI instace managed automatically, because it requires
only additinal and thus non-destructive changes to the current ACI instance.

## How to deploy

1. Create a resource group

    ```shell
    az group create --name myResourceGroup --location eastus
    ```

1. Create a private virtual network

    ```shell
    az network vnet create \
      --resource-group myResourceGroup \
      --location eastus \
      --name myVnet \
      --address-prefix 10.0.0.0/16 \
      --subnet-name mySubnet \
      --subnet-prefix 10.0.0.0/24
    ```

1. Collect ID of virtual subnet, because t's required later on

    ```shell
    az network vnet subnet show -n mySubnet -g myResourceGroup --vnet-name myVNet --query 'id' --output tsv
    ```

1. Delegate the subnet to ACI runtime

    ```shell
    az network vnet subnet update \
      --resource-group myResourceGroup \
      --name mySubnet \
      --vnet-name myVnet \
      --delegations Microsoft.ContainerInstance/containerGroups
    ```

1. Create a private DNS Zone

    ```shell
    az network private-dns zone create --name aci-demo.example.com --resource-group myResourceGroup
    ```

1. Create a Service Principal that is used by `aci-dns-manager` to manage the
   DNS entry for the ACI and Assign RBAC Role to the DNS Zone. For simplicity
   the `Private DNS Zone Contributor` role is used. Refer to the top-level
   [README.md](../../README.md) for a more secure approach.

    ```shell
    dns_zone_id=$(az network private-dns zone show -n aci-demo.example.com -g myResourceGroup --query 'id' --output tsv)
    
    az ad sp create-for-rbac \
      --name 'ACI DNS Manager (Demo)' \
      --role 'Private DNS Zone Contributor' \
      --scopes $dns_zone_id
    ```

    >
    > **Note:**
    >
    > Save the JSON data returned from the command, because they're required to
    > configure various environment variables for the ACI instance
    >

1. Create a YAML file `aci-init.yml` based on the following template. Ensure that all
   placeholders are replaced with information collected before.

    ```yaml
    additional_properties: {}
    apiVersion: '2021-09-01'
    extended_location: null
    identity: null
    location: eastus
    name: aci-dns-manager-demo-init
    properties:
      containers:
        - name: main
          properties:
            image: mcr.microsoft.com/azuredocs/aci-helloworld
            ports:
            - port: 80
              protocol: TCP
            resources:
              requests:
                cpu: 1.0
                memoryInGB: 0.5
      initContainers:
        - name: init
          properties:
            image: ghcr.io/tmeckel/aci-dns-manager:latest
            environmentVariables:
            - name: AZURE_TENANT_ID
              value: <TENANT-ID>
            - name: AZURE_CLIENT_ID
              value: <SERVICE-PRINCIPAL-APPID>
            - name: AZURE_CLIENT_SECRET
              secureValue: <ACI-SUBSCRIPTION-ID>
            - name: ARM_SUBSCRIPTION_ID
              value: <ACI-SUBSCRIPTION-ID>
            - name: ACI_INSTANCE_NAME
              value: aci-dns-manager-demo-init
            - name: ACI_RESOURCE_GROUP_NAME
              value: myResourceGroup
            - name: DNS_ZONE_NAME
              value: aci-demo.example.com

      ipAddress:
        ports:
        - port: 80
          protocol: TCP
        type: Private
      osType: Linux
      restartPolicy: OnFailure
      sku: Standard
      subnetIds:
      - id: <SUBNET-ID>
    tags: {}
    type: Microsoft.ContainerInstance/containerGroups
    ```

1. Create the Azure Container Instance

    ```shell
    az container create --no-wait -g myResourceGroup -f aci-init.yml
    ```

    >
    > **Note:**
    >
    > It is not possible to create an ACI instance in a stopped state. So at the
    > first start `aci-dns-manager` won't be possible to read information about
    > the ACI instance and thus will log an error. In the next step we'll create
    > an Azure RBAC role assignment that will fix this.
    >

    ```shell
    az container logs --container-name init -n aci-dns-manager-demo-init -g myResourceGroup
    ```

    ```logs
    Failed to get to container instance: containerinstance.ContainerGroupsClient#Get: Failure responding to request: StatusCode=403 -- Original Error: autorest/azure: Service returned an error. Status=403 Code="AuthorizationFailed" Message="The client '02c8db02-2e88-4c1f-8c70-2374cbbfd30f' with object id '02c8db02-2e88-4c1f-8c70-2374cbbfd30f' does not have authorization to perform action 'Microsoft.ContainerInstance/containerGroups/read' over scope '/subscriptions/4b661122-c2eb-4d16-b78a-e2cb6b1a464f/resourceGroups/myResourceGroup/providers/Microsoft.ContainerInstance/containerGroups/aci-dns-manager-demo-init' or the scope is invalid. If access was recently granted, please refresh your credentials."
    failed to get container instance
    ```

1. Add `Reader` RBAC Role to Service Principal

    ```shell
    acr_id=$(az container show -g myResourceGroup -n aci-dns-manager-demo-init --query 'id' --output tsv)
    sp_object_id=$(az ad sp list --display-name 'ACI DNS Manager (Demo)' --query '[].objectId' --output tsv)

    az role assignment create --assignee $sp_object_id --role "Reader" --scope $acr_id
    ```

    >
    > **Note:**
    >
    > Per default `aci-dns-manager` will retry to access the ACI instance for 20 seconds.
    > This should be enough time so the RBAC role assignment will in effect and `aci-dns-manager`
    > can read the IP address of the connected ACI instance. If this is not the case `restart` the ACI
    > by using:
    > 
    > `az container restart -n aci-dns-manager-demo-init -g myResourceGroup`
    >

1. Show log of init container and verify that `aci-dns-manager` successfully create a DNS A-Record for the ACI instance

    ```shell
    az container logs --container-name init -n  aci-dns-manager-demo-init -g myResourceGroup
    ```

## References

[YAML reference: Azure Container Instances](https://docs.microsoft.com/en-us/azure/container-instances/container-instances-reference-yaml)
