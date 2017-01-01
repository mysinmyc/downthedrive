package cmdactions

import (
	"flag"
	"net/http"

	"io"

	"github.com/mysinmyc/downthedrive/httpd"
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient/auth"
)

const (
	STYLE_CSS = `

.topBar {        
        width:100%;
        height:150px;
	font-weight:bold;
	font-style:italic;
	text-align: center;
	line-height:150px;
	text-align:center;
}



@media only screen and (max-width: 699px) {
	
	.topBar{
        	font-size:20px;
	}

	.menuBarNotSelectedItem, .menuBarSelectedItem {
		width:100%;
    		padding: 14px 0px;
	}	

}
@media only screen and (min-width: 700px) {

        .topBar {
                font-size:40px;
        }

	.content {
		margin:40px;
		font-size:20px;
	}
	.menuBarNotSelectedItem, .menuBarSelectedItem {
    		padding: 14px 25px;
	}
}

@media only screen and (max-width: 699px) {

        .topBar{
                font-size:20px;
        }

	.content {
		margin:20px;
	}
}



body {
	margin: 0px;
	color:#27406f;
}


.menuBar {
        background-color: black;
        list-style-type:none;
}

a.menuBarNotSelectedItem, a.menuBarSelectedItem {
   	text-align: center;
    	text-decoration: none;
    	display: inline-block;
}

a.menuBarSelectedItem:hover, a.menuBarSelectedItem:active {
        color:white;
        background:black;
}

a.menuBarSelectedItem:link, a.menuBarSelectedItem:visited {
        background:white;
        color:black;
		font-weight: bold;
    	text-decoration: none;
}

a.menuBarNotSelectedItem:hover, a.menuBarNotSelectedItem:active {
        background:white;
        color:black;
}

a.menuBarNotSelectedItem:link, a.menuBarNotSelectedItem:visited {
        background: none;
        color:white;
    	text-decoration: none;
}

.fileElement {
        list-style-type:none;
}

a.fileElement {
   	text-align: center;
    text-decoration: none;
    display: inline-block;
}

a.fileElement:hover, a.fileElement:active {
        color:white;
        background:#27406f;
}

a.fileElement:link, a.fileElement:visited {
        background:white;
        color:#27406f;
		font-weight: bold;
    	text-decoration: none;
}


`

	PAGE_HOME = `

<html>
<head>
<!--<meta name="viewport" content="width=device-width, initial-scale=1">-->
<link rel="stylesheet" type="text/css" href="/style.css">
<link rel="stylesheet" href="http://fontawesome.io/assets/font-awesome/css/font-awesome.css">
</head>
<body>
<table style="width:100%; height:100%">
<tr><td style="height:50" class="menuBar">
  	<a class="menuBarNotSelectedItem" href="/browse?drivePath=/" target="innerFrame"><i class="fa fa-list" aria-hidden="true"></i>&nbsp;Browse</a>
	<a class="menuBarNotSelectedItem" href="/onedrive/auth/begin" target="innerFrame"><i class="fa fa-key" aria-hidden="true"></i>&nbsp;Authenticate</a>
</td></tr>
<tr><td>
<iframe style="border:0; width:100%; height:100%" name="innerFrame"/>
</td></tr>
</table>
</body>
</html>
`
)

type ActionHttpd struct {
	genericAction
	listenAddress string
	localStorePath string
}

func (vSelf *ActionHttpd) ParseCmdLine(pArgs []string) error {

	vFlagSet := vSelf.BuildFlagSet()

	vParseError := vFlagSet.Parse(pArgs)

	if vParseError != nil {
		return vParseError
	}

	vMandatoryParametersError := vSelf.CheckMandatoryParameters()
	if vMandatoryParametersError != nil {
		return vMandatoryParametersError
	}

	return nil
}

func (vSelf *ActionHttpd) AddCustomFlagsTo(pFlagSet *flag.FlagSet) {
	pFlagSet.StringVar(&vSelf.listenAddress, PARAMETER_LISTEN, "localhost:8080", "Listen address")
}

func (vSelf *ActionHttpd) PreInit() error {
	vSelf.customFlagsFunc = vSelf
	return nil
}

func (vSelf *ActionHttpd) initAuthentication() error {

	vClientId := *vSelf.inLineAuthentication.clientIdParameter
	vClientSecret := *vSelf.inLineAuthentication.clientSecretParameter

	if vClientId == "" || vClientSecret == "" {

		vExistingApplicationInfo, vErrorGetting := vSelf.downTheDrive.GetChachedApplicationInfo()

		if vErrorGetting != nil {
			return diagnostic.NewError("Failed to get cached application info ", vErrorGetting)
		}

		if vExistingApplicationInfo != nil {
			if vClientId == "" {
				vClientId = vExistingApplicationInfo.ClientID
			}

			if vClientSecret == "" {
				vClientSecret = vExistingApplicationInfo.ClientSecret
			}
		}

	}

	vAuthHelper := auth.NewHttpAuthHelper(vSelf.listenAddress, vClientId, vClientSecret, []string{"offline_access", "onedrive.readonly"}, "/")
	vSelf.downTheDrive.DenySynchronousAuthentication = true
	vSelf.downTheDrive.SetAuthenticationHelper(vAuthHelper)

	return nil
}

func (vSelf *ActionHttpd) Execute() error {


        vDownTheDrive, vDownTheDriveError:= vSelf.GetDownTheDrive()
        if vDownTheDriveError != nil {
                return diagnostic.NewError("Error getting downthedrive", vDownTheDriveError)
        }

       	vLocalStore, vLocalStoreError:= vSelf.GetLocalStore()
        if vLocalStoreError != nil {
                return diagnostic.NewError("Error getting local store", vLocalStoreError)
        }



	vAuthenticationError := vSelf.initAuthentication()

	if vAuthenticationError != nil {
		return diagnostic.NewError("Failed to initialize authentication", vAuthenticationError)
	}

	http.HandleFunc("/style.css", func(pResponse http.ResponseWriter, pRequest *http.Request) {
		io.WriteString(pResponse, STYLE_CSS)

	})

	http.HandleFunc("/", func(pResponse http.ResponseWriter, pRequest *http.Request) {
		io.WriteString(pResponse, PAGE_HOME)

	})

	vError := httpd.RegisterBrowseDriveHandler(vDownTheDrive)
	if vError != nil {
		return diagnostic.NewError("Failed to register browsedrivehandler", vError)
	}

	vError = httpd.RegisterDownloadHandler(vDownTheDrive,vLocalStore)
	if vError != nil {
		return diagnostic.NewError("Failed to register downloadhandler", vError)
	}

	diagnostic.LogInfo("ActionHttpd.Execute", "website url is http://%s", vSelf.listenAddress)
	return http.ListenAndServe(vSelf.listenAddress, nil)
}

func (vSelf *ActionHttpd) GetDescription() string {
	return "Start an http to access the drive content"
}
