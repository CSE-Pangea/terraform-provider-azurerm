# private branch for Spring Cloud Service Virtual Network Integration

## Instruction
This feature is based on azurerm provider 2.17.0

to build a private binary, you should 
- checkout this repo and switch to current branch
- `make build`

If you want to cross compile for other OS platform, you could change the file: `GNUmakefile`. In the line 23, add cross compile parameters

## how to use private build binary
please See https://www.terraform.io/docs/extend/how-terraform-works.html#plugin-locations for more details

## verify
The way to verify you are using the private build version is to check the debug log of terraform

- you shall see: `[DEBUG] plugin: starting plugin: path=<plugin location>/terraform-provider-azurerm`
- debug log should contain following information
```
[INFO] *************************************************************************************
[INFO] ***** Using Terraform Azure provider with Spring Cloud Service Vnet Integration *****
[INFO] *************************************************************************************
```

## sample usage
```hcl
provider "azurerm" {
  features {}
}

data "azurerm_resource_group" "test" {
  name = "cz"
}

resource "azurerm_virtual_network" "test" {
  name                = "testvnet"
  address_space       = ["10.1.0.0/16"]
  location            = data.azurerm_resource_group.test.location
  resource_group_name = data.azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test1" {
  name                 = "internal1"
  resource_group_name  = data.azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.1.0.0/24"
}

resource "azurerm_subnet" "test2" {
  name                 = "internal2"
  resource_group_name  = data.azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.1.1.0/24"
}

resource "azurerm_role_assignment" "test" {
  scope                = azurerm_virtual_network.test.id
  role_definition_name = "Owner"
  principal_id         = "d2531223-68f9-459e-b225-5592f90d145e"
}

resource "azurerm_spring_cloud_service" "test" {
  name                = "sc-cz"
  resource_group_name = data.azurerm_resource_group.test.name
  location            = data.azurerm_resource_group.test.location
  
  network {
    app_subnet_id             = azurerm_subnet.test1.id
    service_runtime_subnet_id = azurerm_subnet.test2.id
    cidr                      = ["10.4.0.0/16", "10.5.0.0/16", "10.3.0.1/16"]
  }
}
```

# Terraform Provider for Azure (Resource Manager)

Version 2.0 of the AzureRM Provider requires Terraform 0.12.x and later.

* [Terraform Website](https://www.terraform.io)
* [AzureRM Provider Documentation](https://www.terraform.io/docs/providers/azurerm/index.html)
* [AzureRM Provider Usage Examples](https://github.com/terraform-providers/terraform-provider-azurerm/tree/master/examples)
* [Slack Workspace for Contributors](https://terraform-azure.slack.com) ([Request Invite](https://join.slack.com/t/terraform-azure/shared_invite/enQtNDMzNjQ5NzcxMDc3LWNiY2ZhNThhNDgzNmY0MTM0N2MwZjE4ZGU0MjcxYjUyMzRmN2E5NjZhZmQ0ZTA1OTExMGNjYzA4ZDkwZDYxNDE))

## Usage Example

```
# Configure the Microsoft Azure Provider
provider "azurerm" {
  # We recommend pinning to the specific version of the Azure Provider you're using
  # since new versions are released frequently
  version = "=2.20.0"

  features {}

  # More information on the authentication methods supported by
  # the AzureRM Provider can be found here:
  # http://terraform.io/docs/providers/azurerm/index.html

  # subscription_id = "..."
  # client_id       = "..."
  # client_secret   = "..."
  # tenant_id       = "..."
}

# Create a resource group
resource "azurerm_resource_group" "example" {
  name     = "production-resources"
  location = "West US"
}

# Create a virtual network in the production-resources resource group
resource "azurerm_virtual_network" "test" {
  name                = "production-network"
  resource_group_name = "${azurerm_resource_group.example.name}"
  location            = "${azurerm_resource_group.example.location}"
  address_space       = ["10.0.0.0/16"]
}
```

Further [usage documentation is available on the Terraform website](https://www.terraform.io/docs/providers/azurerm/index.html).

## Developer Requirements

* [Terraform](https://www.terraform.io/downloads.html) version 0.12.x +
* [Go](https://golang.org/doc/install) version 1.14.x (to build the provider plugin)

### On Windows 

If you're on Windows you'll also need:
* [Git Bash for Windows](https://git-scm.com/download/win)
* [Make for Windows](http://gnuwin32.sourceforge.net/packages/make.htm)

For *GNU32 Make*, make sure its bin path is added to PATH environment variable.*

For *Git Bash for Windows*, at the step of "Adjusting your PATH environment", please choose "Use Git and optional Unix tools from Windows Command Prompt".*

Or install via [Chocolatey](https://chocolatey.org/install) (`Git Bash for Windows` must be installed per steps above)
```powershell
choco install make golang terraform -y
refreshenv
```

You must run  `Developing the Provider` commands in `bash` because `sh` scrips are invoked as part of these.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.13+ is **required**). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

First clone the repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-azurerm`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-azurerm
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-azurerm
```

Once inside the provider directory, you can run `make tools` to install the dependent tooling required to compile the provider.

At this point you can compile the provider by running `make build`, which will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-azurerm
...
```

You can also cross-compile if necessary:

```sh
GOOS=windows GOARCH=amd64 make build
```

In order to run the `Unit Tests` for the provider, you can run:

```sh
$ make test
```

The majority of tests in the provider are `Acceptance Tests` - which provisions real resources in Azure. It's possible to run the entire acceptance test suite by running `make testacc` - however it's likely you'll want to run a subset, which you can do using a prefix, by running:

```sh
make acctests SERVICE='resource' TESTARGS='-run=TestAccAzureRMResourceGroup' TESTTIMEOUT='60m'
```

The following Environment Variables must be set in your shell prior to running acceptance tests:

- `ARM_CLIENT_ID`
- `ARM_CLIENT_SECRET`
- `ARM_SUBSCRIPTION_ID`
- `ARM_TENANT_ID`
- `ARM_ENVIRONMENT`
- `ARM_METADATA_HOST`
- `ARM_TEST_LOCATION`
- `ARM_TEST_LOCATION_ALT`
- `ARM_TEST_LOCATION_ALT2`

**Note:** Acceptance tests create real resources in Azure which often cost money to run.

---

## Developer: Scaffolding the Website Documentation

You can scaffold the documentation for a Data Source by running:

```sh
$ make scaffold-website BRAND_NAME="Resource Group" RESOURCE_NAME="azurerm_resource_group" RESOURCE_TYPE="data"
```

You can scaffold the documentation for a Resource by running:

```sh
$ make scaffold-website BRAND_NAME="Resource Group" RESOURCE_NAME="azurerm_resource_group" RESOURCE_TYPE="resource" RESOURCE_ID="/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1"
```
