# AzKube is an Azure k8s bring-up solution !

## Deprecation Notice (!!!)

The supported method for deploying Kubernetes clusters in Azure is now [`kubernetes-anywhere`](https://github.com/kubernetes/kubernetes-anywhere).

Please see [`azure-kubernetes-status`](https://github.com/colemickens/azure-kubernetes-status) for more information.


----------------------------------------------------------------------
----------------------------------------------------------------------
----------------------------------------------------------------------
----------------------------------------------------------------------

## Overview

	Tool used to deploy and bootstrap a Kubernetes cluster in Azure.


## Running azkube deployment

	```
	docker run -it \
		colemickens/azkube:latest /opt/azkube/azkube \
			--flavor="coreos" \
			--node-size="Standard_D1" \
			--master-size="Standard_D1" \
			--node-count="3" \
			--location="westus" \
			--no-cloud-provider \ 
			--subscription-id="{Azure subscription id}"
	```

# Flavors for Azkube

## The "coreos" flavor :

	- CoreOS Stable version
	- No LoadBalancer to server the traffic to the outside
	- A Standard storage
	- Not using the Ephemerials SSDs of the Azure D or DS serie to mount Docker lib

## The "coreos-ssd" :

	- CoreOS Stable version
	- No LoadBalancer to serve the traffic to the outside (you must use your own solution)
	- A Premium Storage
	- Using the Ephemerials SSDs to mount the Docker lib on each nodes

## The "coreos-lb" flavor :

	- CoreOS Stable version
	- A LoadBalancer Azure (L4) linked on each nodes on the TCP:80 port (use this guide to deploy a Ingress/Nginx solution to use properly the Azure Load-Balancer).
	- A Standard Storage
	- Not using Ephemerials SSDs of the Azure D or DS serie to mount Docker lib

## The "coreos-lbssd" flavor :

	- CoreOS Stable version
	- A LoadBalancer Azure (L4) linked on each nodes on the TCP:80 port (use this guide to deploy a Ingress/Nginx solution to use properly the Azure Load-Balancer).
	- A Premium Storage
	- Using the Ephemerials SSDs to mount the Docker lib on each nodes


## Usage

``` 
A Kubernetes deployment helper for Azure

Usage:
  azkube [command]

Available Commands:
  deploy      creates a new kubernetes cluster in Azure
  destroy     destroy a deployment (and its containing resource group)
  scale       scale a deployment's vm scale set
```
```
Usage:
  azkube deploy [flags]

Flags:
      --cluster-domain string              the dns suffix used in the cluster (used as a SAN in the PKI generation) (default "cluster.local")
      --deployment-name string             deployment identifier (used to name output, resource group, and other resources)
      --flavor string                      the flavor of deployment to perform (currently supported: coreos, coreos-ssd, coreos-lb, coreos-lbssd) (default "coreos")
      --kubernetes-hyperkube-spec string   docker spec for hyperkube container to use (default "gcr.io/google_containers/hyperkube-amd64:v1.3.0")
      --location azure location list       location to deploy Azure resource (these can be found by running azure location list with azure-xplat-cli) (default "westus")
      --master-extra-fqdns value           comma delimited list of SANs for the master (default [])
      --master-fqdn string                 fqdn for master (used for PKI). calculated from cloudapp dns for master's public ip
      --master-private-ip string           the internal vnet ip address to use for the master (used as a SAN in the PKI generation) (default "10.0.1.4")
      --master-size string                 size of the master virtual machine (default "Standard_A1")
      --no-cloud-provider                  skip service principal steps entirely. this suppresses creation of a new service principal and prevents passthrough of client_secret credentials
      --node-count int                     initial number of node virtual machines (default 3)
      --node-size string                   size of the node virtual machines (default "Standard_A1")
      --output-directory string            output directory (this is derived from --deployment-name if omitted)
      --resource-group string              resource group to deploy to (this is derived from --deployment-name if omitted)
      --service-principal-passthrough      bypass service principal creation and use deployers credentials for cluster's service principal
      --username string                    username to virtual machines (default "kube")

Global Flags:
      --auth-method device        auth method (default:device, `client_secret`, `client_certificate`) (default "device")
      --certificate-path string   path to client certificate (used with --auth-method=client_certificate)
      --client-id string          client id (used with --auth-method=[client_secret|client_certificate])
      --client-secret string      client secret (used with --auth-mode=client_secret)
      --debug                     debug mode, outputs more logging
      --private-key-path string   path to private key (used with --auth-method=client_certificate)
      --subscription-id string    azure subscription id
      
```

## Motivations

1. Existing shell script was fragile.
2. azure-xplat-cli changes out from underneath of us, and is slow, and doesn't handle errors well.
3. Need a tool and process to create service principals and configure them appropriately.
4. Need a tool to consume scripts, interweave ARM template variables/parameters, and
   output a "deployable" template.

## Future

As Azure lands support for managed service identity and metadata facilities, some of the need for this tool will be alleviated.

## Known Issues
1. Needs tests
2. `scale` is not implemented
3. `destroy` is not implemented


## Potential Future Improvements
1. Add an Ubuntu 16.04 Flavor (lets us avoid `upstart`)
2. Validate `location`, `master-size`, `node-size` against the live services


## Bugs
1. `kube-controller-manager` seems to flat out ignore my `kubeconfig` file.
   The other components all read the `kubeconfig` file and connect correctly.
   Skimming the source, I see a lack of warning about missing `kubeconfig` file,
   and no permission or opening errors, so I'm wondering if there's a bug in the
   `hyperkube:v1.1.8` image.

2. As a result of #1, the insecure port is still open on the apiserver. (Though
   it is firewalled to the outside world.

## Important Notes
1. The user who executes the application must have permission to provision
   additional applications. This is difficult to achieve unless you're using
   the device auth method (`--auth-method=device` which is the default). If you
   wish to automate the use of this tool to remove all interactivity, you must create
   a new Application in your Azure Active Directory Tenant. Then you must use the ADAL
   Powershell Toolkit to grant the service principal associated with the application
   the AD Role "Company User".

2. The resulting "templates" are fully parameterized and generic. They can be uploaded
   and used by others. The values that are entered for the parameters are interpolated
   into the cloud-config scripts which are then interpolated into the ARM Templates.
   These are copied to the deployment directory so that they might be reused (if desired).
