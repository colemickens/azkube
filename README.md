# azkube


## Overview
Tool used to deploy and bootstrap a Kubernetes cluster in Azure.


## Running

### From source
```
go get github.com/colemickens/azkube
$GOPATH/bin/azkube --tenant-id="{your tenant id}" --subscription-id="{subscription id}"
```

### Docker
```
docker run -it colemickens/azkube:latest /azkube --tenant-id="{your tenant id}" --subscription-id="{subscription id}"
```

## Usage

[ insert generated output from cobra's command facilities ]


## Motivations
1. Existing shell script was fragile.
2. azure-xplat-cli changes out from underneath of us, and is slow, and doesn't handle errors well.
3. Need a tool and process to create service principals and configure them appropriately.
4. Need a tool to consume scripts, interweave ARM template variables/parameters, and output a "deployable" template.

## Future

As Azure lands support for managed service identity and metadata facilities, much of the need for this tool will be alleviated.
Further, in lieu of those, Terraform is much more flexible and powerful than ARM Templates. Particularly when the Azurerm provider gains the ability to create ServicePrincipals, then a single terraform file could do everything this tool does. This is due to the power of HCL, and the fact that Terraform can express more concepts than ARM Templates.

## Future Improvements
1. Introduce concept of "flavors" with a defined interface that we hand off
2. Introduce an ubuntu 16.04 LTS (beta) flavor


## Notes
1. The user who executes the application must have permission to provision additional applications.
2. The resulting "templates" are fully parameterized and generic. They can be uploaded and used by others.
