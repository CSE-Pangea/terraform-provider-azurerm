package client

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/preview/synapse/2020-02-01-preview/accesscontrol"
	"github.com/Azure/azure-sdk-for-go/services/preview/synapse/mgmt/2019-06-01-preview/synapse"
	"github.com/Azure/go-autorest/autorest"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	FirewallRulesClient      *synapse.IPFirewallRulesClient
	WorkspaceClient          *synapse.WorkspacesClient
	WorkspaceAadAdminsClient *synapse.WorkspaceAadAdminsClient

	synapseAuthorizer autorest.Authorizer
}

func NewClient(o *common.ClientOptions) *Client {
	firewallRuleClient := synapse.NewIPFirewallRulesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&firewallRuleClient.Client, o.ResourceManagerAuthorizer)

	workspaceClient := synapse.NewWorkspacesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&workspaceClient.Client, o.ResourceManagerAuthorizer)

	workspaceAadAdminsClient := synapse.NewWorkspaceAadAdminsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&workspaceAadAdminsClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		FirewallRulesClient:      &firewallRuleClient,
		WorkspaceClient:          &workspaceClient,
		WorkspaceAadAdminsClient: &workspaceAadAdminsClient,

		synapseAuthorizer: o.SynapseAuthorizer,
	}
}

func (client Client) AccessControlClient(workspaceName string) *accesscontrol.BaseClient {
	endpoint := getWorkspaceEndpoint("dev.azuresynapse.net", workspaceName)
	accessControlClient := accesscontrol.New(endpoint)
	accessControlClient.Client.Authorizer = client.synapseAuthorizer
	return &accessControlClient
}

// getWorkspaceEndpoint returns the endpoint for API Operations on this workspace
func getWorkspaceEndpoint(baseUri string, workspaceName string) string {
	return fmt.Sprintf("https://%s.%s", workspaceName, baseUri)
}
