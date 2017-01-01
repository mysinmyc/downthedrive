package downthedrive

import (
	"github.com/mysinmyc/onedriveclient/auth"
)

//DownTheDrive: onedrive utiliy to index the drive into a database and access the content
type DownTheDrive struct {
	config                        DownTheDriveConfig
	authenticationHelper          auth.AuthenticationHelper
	authenticationInfo            *DownTheDriveAuthenticationInfo
	DenySynchronousAuthentication bool
	defaultContext		      *DownTheDriveContext
}

//NewDownTheDrive: create a new instance of downthedrive
//Parameters:
//	DownTheDriveConfig = configuration
//Returns:
//	*DownTheDrive = DownTheDrive instance
//	error = nil if succeded otherwise the error occurred
func NewDownTheDrive(pConfig DownTheDriveConfig) (*DownTheDrive, error) {

	return &DownTheDrive{config: pConfig, authenticationInfo: &DownTheDriveAuthenticationInfo{}}, nil
}


func (vSelf *DownTheDrive) Close() error {
	if vSelf.defaultContext != nil {
		return vSelf.defaultContext.Close()
	}
	return nil
}
