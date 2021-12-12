# ACI DNS Manager: Using S6-overlay

## Introduction

This example shows how to configure `aci-dns-manager` using [S6-Overlay](https://github.com/just-containers/s6-overlay)
to manage a DNS A-Record for the deployed ACI (main) container bound to an
virtual private network.

S6-Overlay is a Container init system based on the Unix S6 init system. The major advantages of S6-Overlay are:

* supports a list of scripts that are executed before the main entrypoint of the
  container is executed and a list of scripts that executed on shutdown
* monitors the main entry point and collects process zombies
* handles i.e. forwards signals for the main entrypoint
* allows the restriction of environment variables i.e. what variables are
  accessible by init, shutdown scripts and the main entrypoint

As stated this example will show the usage of `aci-dns-manager` executing inside
a S6-Overlay init script and in a shutdown script which will delete the created
DNS Records when the container goes offline.

## How to deploy

1. Create a resource group

    ```shell
    az group create --name myResourceGroup --location eastus
    ```

1. Create a Azure Container Registry

    ```shell
    az acr create --resource-group myResourceGroup \
      --name mycontainerregistry907 --sku Basic --admin-enabled true
    ```

1. Login into created Registry

    >
    > **Note:**
    >
    > If you're creating AZure Resources and building the DOcker Image on
    > different machines/installations, get the credentails from the ACR and do a
    > Docker Login.
    >
    > Otherwise use: `az acr login -n mycontainerregistry907`
    >

    Get credentials

    ```shell
    az acr credential show -n mycontainerregistry907
    ```

    Docker Login

    ```shell
    docker login -u mycontainerregistry907 mycontainerregistry907.azurecr.io
    ```

1. Build Docker Image

    ```shell
    docker build -t mycontainerregistry907.azurecr.io/aci-dns-manager-demo:latest -f Dockerfile .
    ```

1. Push Docker Image

    ```shell
    docker push mycontainerregistry907.azurecr.io/aci-dns-manager-demo:latest
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

1. Create the Azure Container Instance

    ```shell
    az container create  \
      --resource-group myResourceGroup \
      --image mycontainerregistry907.azurecr.io/aci-dns-manager-demo:latest \
      --ip-address Private \
      --location eastus \
      --name aci-dns-manager-demo \
      --os-type Linux \
      --ports 80 \
      --protocol TCP \
      --registry-login-server mycontainerregistry907.azurecr.io \
      --registry-username myContainerRegistry907 \
      --registry-password <REGISTRY_PASSWORD> \
      --subnet mySubnet \
      --vnet myVnet \
      --environment-variables AZURE_TENANT_ID=<TENANT-ID> AZURE_CLIENT_ID=<SERVICE-PRINCIPAL-APPID> ARM_SUBSCRIPTION_ID=<ACI-SUBSCRIPTION-ID> ACI_INSTANCE_NAME=aci-dns-manager-demo ACI_RESOURCE_GROUP_NAME=myResourceGroup DNS_ZONE_NAME=aci-demo.example.com \
      --secure-environment-variables AZURE_CLIENT_SECRET=<service-principal-secret>
    ```

    >
    > **Note:**
    >
    > It is not possible to create an ACI instance in a stopped state. So at the
    > first start `aci-dns-manager` won't be possible to read information about
    > the ACI instance and thus will log an error. In the next step we'll create
    > an Azure RBAC role assignment that will fix this.
    >

    ```log
    [s6-init] making user provided files available at /var/run/s6/etc...exited 0.
    [s6-init] ensuring user provided files have correct perms...exited 0.
    [fix-attrs.d] applying ownership & permissions fixes...
    [fix-attrs.d] done.
    [cont-init.d] executing container initialization scripts...
    [cont-init.d] update-dns.sh: executing...
    Failed to get to container instance: containerinstance.ContainerGroupsClient#Get: Failure responding to request: StatusCode=403 -- Original Error: autorest/azure: Service returned an error. Status=403 Code="AuthorizationFailed" Message="The client '02c8db02-2e88-4c1f-8c70-2374cbbfd30f' with object id '02c8db02-2e88-4c1f-8c70-2374cbbfd30f' does not have authorization to perform action 'Microsoft.ContainerInstance/containerGroups/read' over scope '/subscriptions/4b661122-c2eb-4d16-b78a-e2cb6b1a464f/resourceGroups/myResourceGroup/providers/Microsoft.ContainerInstance/containerGroups/aci-dns-manager-demo' or the scope is invalid. If access was recently granted, please refresh your credentials."
    failed to get container instance    
    ```

1. Add `Reader` RBAC Role to Service Principal

    ```shell
    acr_id=$(az container show -g myResourceGroup -n aci-dns-manager-demo --query 'id' --output tsv)
    sp_object_id=$(az ad sp list --display-name 'ACI DNS Manager (Demo)' --query '[].objectId' --output tsv)

    az role assignment create --assignee $sp_object_id --role "Reader" --scope $acr_id
    ```

1. Restart ACI instance, so that the new RBAC role assignment got picked up by the container

    ```shell
    az container restart -g myResourceGroup -n aci-dns-manager-demo
    ```

1. Show log and verify that `aci-dns-manager` successfully create a DNS A-Record for the ACI instance

    ```shell
    az container logs -g myResourceGroup -n aci-dns-manager-demo
    ```

    ```log
    [s6-init] making user provided files available at /var/run/s6/etc...exited 0.
    [s6-init] ensuring user provided files have correct perms...exited 0.
    [fix-attrs.d] applying ownership & permissions fixes...
    [fix-attrs.d] done.
    [cont-init.d] executing container initialization scripts...
    [cont-init.d] update-dns.sh: executing...
    2021/12/12 11:54:51 Container instance with id /subscriptions/4b661122-c2eb-4d16-b78a-e2cb6b1a464f/resourceGroups/myResourceGroup/providers/Microsoft.ContainerInstance/containerGroups/aci-dns-manager-demo has ipv4 address 10.0.0.4
    2021/12/12 11:54:51 Successfully created A record with ETag 9a70d3b9-6803-46cc-b9b8-9f9e232916ba, FQDN aci-dns-manager-demo.aci-demo.example.com. and TTL 3600
    [cont-init.d] update-dns.sh: exited 0.
    [cont-init.d] done.
    [services.d] starting services
    [services.d] done.
    listening on port 80    
    ```
