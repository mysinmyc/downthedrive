package cmdactions

import "github.com/mysinmyc/gocommons/diagnostic"

type ActionAuthenticate struct {
	genericAction
}

func (vSelf *ActionAuthenticate) ParseCmdLine(pArgs []string) error {

	vFlagSet := vSelf.BuildFlagSet()
	vParseError := vFlagSet.Parse(pArgs)

	if vParseError != nil {
		return vParseError
	}

	vMandatoryParametersError := vSelf.CheckMandatoryParameters()

	if vMandatoryParametersError != nil {
		return vMandatoryParametersError
	}

	if vSelf.inLineAuthentication.HasEnoughInformations() == false {
		return diagnostic.NewError("missing required informations to perform the authentication (clientId and clientSecret)", nil)
	}

	return nil
}

func (vSelf *ActionAuthenticate) Execute() error {

        vDownTheDrive, vDownTheDriveError:= vSelf.GetDownTheDrive()
        if vDownTheDriveError != nil {
                return diagnostic.NewError("Error getting downthedrive", vDownTheDriveError)
        }

	vAuthenticationError := vDownTheDrive.PerformAuthentication()

	if vAuthenticationError != nil {
		return diagnostic.NewError("Failed to perform authentication", vAuthenticationError)
	}

	vDownTheDrive.SaveAuthenticationInfo()

	return nil
}

func (vSelf *ActionAuthenticate) GetDescription() string {
	return "Execute the authentication on the drive and store credentials in the database"
}
