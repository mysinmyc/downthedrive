package cmdactions

import (
	"flag"

	"github.com/mysinmyc/downthedrive"
	"github.com/mysinmyc/onedriveclient/auth"
)

type inLineAuthentication struct {
	clientIdParameter     *string
	clientSecretParameter *string
}

func (vSelf *inLineAuthentication) AddParametersTo(pFlagSet *flag.FlagSet) {
	vSelf.clientIdParameter = pFlagSet.String(PARAMETER_CLIENTID, "", "Client id for authentication")
	vSelf.clientSecretParameter = pFlagSet.String(PARAMETER_CLIENTSECRET, "", "Client secret for authentication")
}

func (vSelf *inLineAuthentication) HasEnoughInformations() bool {
	vClientId := *vSelf.clientIdParameter
	vClientSecret := *vSelf.clientSecretParameter
	return vClientId != "" && vClientSecret != ""
}

func (vSelf *inLineAuthentication) ConfigureDownTheDrive(pDownTheDrive *downthedrive.DownTheDrive) bool {

	vClientId := *vSelf.clientIdParameter
	vClientSecret := *vSelf.clientSecretParameter

	if vClientId == "" || vClientSecret == "" {
		return false
	}

	pDownTheDrive.SetAuthenticationHelper(auth.NewOfflineAuthHelper(vClientId, vClientSecret, []string{"offline_access", "onedrive.readonly"}))
	return true
}
