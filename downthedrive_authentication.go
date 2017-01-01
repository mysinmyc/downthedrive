package downthedrive

import (
	"time"

	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/gocommons/persistency"
	"github.com/mysinmyc/onedriveclient/auth"
)

type DownTheDriveAuthenticationInfo struct {
	auth.StaticAuthenticationInfo
}

func (vSelf *DownTheDriveAuthenticationInfo) GetIdInDb() string {
	return "OneDriveAuthenticationInfo"
}

func (vSelf *DownTheDrive) SetAuthenticationHelper(pAuthenticationHelper auth.AuthenticationHelper) {
	vSelf.authenticationHelper = pAuthenticationHelper
	pAuthenticationHelper.SetAuthenticationHandler(vSelf.onAuthentication)
}

func (vSelf *DownTheDrive) onAuthentication(pAuthenticationToken *auth.AuthenticationToken, pApplicationInfo auth.ApplicationInfo) {

	diagnostic.LogInfo("DownTheDrive.onAuthentication", "authentication token %v", pAuthenticationToken)

	vSelf.authenticationInfo.StaticAuthenticationInfo.AuthenticationToken = pAuthenticationToken
	vSelf.authenticationInfo.StaticAuthenticationInfo.ApplicationInfo = &pApplicationInfo
}

func (vSelf *DownTheDrive) GetChachedApplicationInfo() (*auth.ApplicationInfo, error) {

	if vSelf.authenticationInfo.StaticAuthenticationInfo.ApplicationInfo != nil {
		return vSelf.authenticationInfo.ApplicationInfo, nil
	}

	vAuthenticationInfo := &DownTheDriveAuthenticationInfo{}

	vContext, vContextError:= vSelf.GetDefaultContext()
	if vContextError != nil {
		return nil, diagnostic.NewError("Error while getting default context", vContextError)
	}

	vDb, vDbError := vContext.GetDb()
	if vDbError != nil {
		return nil, diagnostic.NewError("Error while getting db from context", vDbError)
	}
	vLoadError := vDb.LoadBean(vAuthenticationInfo)

	if vLoadError != nil {
		if _, vIsNotFound := diagnostic.GetMainError(vLoadError, false).(*persistency.BeanNotFoundError); vIsNotFound == false {
			return nil, diagnostic.NewError("Error loading authentication info ", vLoadError)
		}

	}

	if vAuthenticationInfo == nil {
		return nil, nil
	}
	return vAuthenticationInfo.ApplicationInfo, nil

}

func (vSelf *DownTheDrive) getAuthenticationInfo() (vRisAuthenticationInfo *DownTheDriveAuthenticationInfo, vRisError error) {

	if vSelf.authenticationInfo != nil && vSelf.authenticationInfo.StaticAuthenticationInfo.AuthenticationToken != nil {
		return vSelf.authenticationInfo, nil
	}

	diagnostic.LogDebug("DownTheDrive.getAuthenticationInfo", "Loading Authentication info...")
	vAuthenticationInfo := &DownTheDriveAuthenticationInfo{}

	vContext, vContextError:= vSelf.GetDefaultContext()
	if vContextError != nil {
		return nil, diagnostic.NewError("Error while getting default context", vContextError)
	}

	vDb, vDbError := vContext.GetDb()
	if vDbError != nil {
		return nil, diagnostic.NewError("Error while getting db from context", vDbError)
	}

	vLoadError := vDb.LoadBean(vAuthenticationInfo)

	if vLoadError != nil {
		if _, vIsNotFound := vLoadError.(*persistency.BeanNotFoundError); vIsNotFound == false {
			return nil, diagnostic.NewError("Error loading authentication info ", vLoadError)
		}

	}

	if vAuthenticationInfo.AuthenticationToken != nil && vAuthenticationInfo.AuthenticationToken.IsValid() {
		diagnostic.LogDebug("DownTheDrive.getAuthenticationInfo", "Already authenticated in db")
		vSelf.authenticationInfo.StaticAuthenticationInfo.AuthenticationToken = vAuthenticationInfo.StaticAuthenticationInfo.AuthenticationToken
		vSelf.authenticationInfo.StaticAuthenticationInfo.ApplicationInfo = vAuthenticationInfo.StaticAuthenticationInfo.ApplicationInfo
		return vAuthenticationInfo, nil
	}

	vError := vSelf.PerformAuthentication()

	if vError != nil {
		return nil, diagnostic.NewError("Authentication failed", vError)
	}
	return vSelf.authenticationInfo, nil

}

func (vSelf *DownTheDrive) PerformAuthentication() (vRisError error) {

	if vSelf.DenySynchronousAuthentication {
		return diagnostic.NewError("Synchronous authentication rejected", nil)
	}

	if vSelf.authenticationHelper == nil {
		return diagnostic.NewError("No Authentication helper defined to obtain an authentication token", nil)
	}
	vAuthenticationToken, vAuthenticationError := vSelf.authenticationHelper.WaitAuthenticationToken(time.Second * 120)

	if vAuthenticationError != nil {
		return diagnostic.NewError("An error occurred during authentication", vAuthenticationError)
	}

	vApplicationInfo := vSelf.authenticationHelper.GetApplicationInfo()
	vSelf.authenticationInfo.StaticAuthenticationInfo.AuthenticationToken = vAuthenticationToken
	vSelf.authenticationInfo.StaticAuthenticationInfo.ApplicationInfo = &vApplicationInfo

	return nil
}

func (vSelf *DownTheDrive) SaveAuthenticationInfo() error {

	if vSelf.authenticationInfo == nil {
		return diagnostic.NewError("Missing authentication info", nil)
	}

	vContext, vContextError:= vSelf.GetDefaultContext()
	if vContextError != nil {
		return diagnostic.NewError("Error while getting default context", vContextError)
	}

	vDb, vDbError := vContext.GetDb()
	if vDbError != nil {
		return diagnostic.NewError("Error while getting db from context", vDbError)
	}
	vDb.SaveBean(vSelf.authenticationInfo)
	return nil
}

func (vSelf *DownTheDrive) GetAuthenticationToken() (*auth.AuthenticationToken, error) {

	vAuthenticatioInfo, vAuthenticatioInfoError := vSelf.getAuthenticationInfo()

	if vAuthenticatioInfoError != nil {
		return nil, diagnostic.NewError("failed to get authentication info ", vAuthenticatioInfoError)
	}
	return vAuthenticatioInfo.AuthenticationToken, nil
}

func (vSelf *DownTheDrive) GetApplicationInfo() (auth.ApplicationInfo, error) {
	return *vSelf.authenticationInfo.StaticAuthenticationInfo.ApplicationInfo, nil
}
