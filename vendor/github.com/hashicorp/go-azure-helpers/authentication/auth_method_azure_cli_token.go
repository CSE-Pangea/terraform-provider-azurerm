package authentication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure/cli"
	"github.com/hashicorp/go-multierror"
)

type azureCliTokenAuth struct {
	profile                      *azureCLIProfile
	servicePrincipalAuthDocsLink string
}

func (a azureCliTokenAuth) build(b Builder) (authMethod, error) {
	auth := azureCliTokenAuth{
		profile: &azureCLIProfile{
			clientId:       b.ClientID,
			environment:    b.Environment,
			subscriptionId: b.SubscriptionID,
			tenantId:       b.TenantID,
		},
		servicePrincipalAuthDocsLink: b.ClientSecretDocsLink,
	}
	profilePath, err := cli.ProfilePath()
	if err != nil {
		return nil, fmt.Errorf("Error loading the Profile Path from the Azure CLI: %+v", err)
	}

	profile, err := cli.LoadProfile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("Azure CLI Authorization Profile was not found. Please ensure the Azure CLI is installed and then log-in with `az login`.")
	}

	auth.profile.profile = profile

	// Authenticating as a Service Principal doesn't return all of the information we need for authentication purposes
	// as such Service Principal authentication is supported using the specific auth method
	if authenticatedAsAUser := auth.profile.verifyAuthenticatedAsAUser(); !authenticatedAsAUser {
		return nil, fmt.Errorf(`Authenticating using the Azure CLI is only supported as a User (not a Service Principal).

To authenticate to Azure using a Service Principal, you can use the separate 'Authenticate using a Service Principal'
auth method - instructions for which can be found here: %s

Alternatively you can authenticate using the Azure CLI by using a User Account.`, auth.servicePrincipalAuthDocsLink)
	}

	err = auth.profile.populateFields()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving the Profile from the Azure CLI: %s Please re-authenticate using `az login`.", err)
	}

	err = auth.profile.populateClientId()
	if err != nil {
		return nil, fmt.Errorf("Error populating Client ID from the Azure CLI: %+v", err)
	}

	return auth, nil
}

func (a azureCliTokenAuth) isApplicable(b Builder) bool {
	return b.SupportsAzureCliToken
}

func (a azureCliTokenAuth) getAuthorizationToken(sender autorest.Sender, oauth *OAuthConfig, endpoint string) (autorest.Authorizer, error) {
	if oauth.OAuth == nil {
		return nil, fmt.Errorf("Error getting Authorization Token for cli auth: an OAuth token wasn't configured correctly; please file a bug with more details")
	}

	// the Azure CLI appears to cache these, so to maintain compatibility with the interface this method is intentionally not on the pointer
	token, err := obtainAuthorizationToken(endpoint, a.profile.subscriptionId)
	if err != nil {
		return nil, fmt.Errorf("Error obtaining Authorization Token from the Azure CLI: %s", err)
	}

	adalToken, err := token.ToADALToken()
	if err != nil {
		return nil, fmt.Errorf("Error converting Authorization Token to an ADAL Token: %s", err)
	}

	spt, err := adal.NewServicePrincipalTokenFromManualToken(*oauth.OAuth, a.profile.clientId, endpoint, adalToken)
	if err != nil {
		return nil, err
	}

	auth := autorest.NewBearerAuthorizer(spt)
	return auth, nil
}

func (a azureCliTokenAuth) name() string {
	return "Obtaining a token from the Azure CLI"
}

func (a azureCliTokenAuth) populateConfig(c *Config) error {
	c.ClientID = a.profile.clientId
	c.Environment = a.profile.environment
	c.SubscriptionID = a.profile.subscriptionId
	c.TenantID = a.profile.tenantId

	//get object ID by shelling out
	return nil
}

func (a azureCliTokenAuth) validate() error {
	var err *multierror.Error

	errorMessageFmt := "A %s was not found in your Azure CLI Credentials.\n\nPlease login to the Azure CLI again via `az login`"

	if a.profile == nil {
		return fmt.Errorf("Azure CLI Profile is nil - this is an internal error and should be reported.")
	}

	if a.profile.clientId == "" {
		err = multierror.Append(err, fmt.Errorf(errorMessageFmt, "Client ID"))
	}

	if a.profile.subscriptionId == "" {
		err = multierror.Append(err, fmt.Errorf(errorMessageFmt, "Subscription ID"))
	}

	if a.profile.tenantId == "" {
		err = multierror.Append(err, fmt.Errorf(errorMessageFmt, "Tenant ID"))
	}

	return err.ErrorOrNil()
}

func obtainAuthorizationToken(endpoint string, subscriptionId string) (*cli.Token, error) {

	azCmdJsonUnmarshal"account", "get-access-token", "--resource", endpoint, "--subscription", subscriptionId, "-o=json"
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.Command("az", "account", "get-access-token", "--resource", endpoint, "--subscription", subscriptionId, "-o=json")

	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("Error launching Azure CLI: %+v", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("Error waiting for the Azure CLI: %+v", err)
	}

	stdOutStr := stdout.String()
	stdErrStr := stderr.String()

	if stdErrStr != "" {
		return nil, fmt.Errorf("Error retrieving access token from Azure CLI: %s", strings.TrimSpace(stdErrStr))
	}

	var token *cli.Token
	err := json.Unmarshal([]byte(stdOutStr), &token)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling Access Token from the Azure CLI: %s", err)
	}

	return token, nil
}

func azCmdJsonUnmarshal(endpoint string, arg ...string) (*cli.Token, error) {
	cmd := exec.Command("az", arg)

	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("Error launching Azure CLI: %+v", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("Error waiting for the Azure CLI: %+v", err)
	}

	stdOutStr := stdout.String()
	stdErrStr := stderr.String()

	if stdErrStr != "" {
		return nil, fmt.Errorf("Error retrieving running Azure CLI: %s", strings.TrimSpace(stdErrStr))
	}

	err := json.Unmarshal([]byte(stdOutStr), &token)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling the result of Azure CLI: %s", err)
	}


}