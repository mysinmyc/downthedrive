package cmdactions

import (
	"flag"
	"os"

	"github.com/mysinmyc/downthedrive"
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/gocommons/persistency"
)

type CmdLineAction interface {
	ParseCmdLine([]string) error
	GetDescription() string

	Execute() error

	PrintHelp()
}

type PreInitializable interface {
	PreInit() error
}

type Initializable interface {
	Init() error
}

type PostExecution interface {
	PostExecution() error
}

type CustomFlags interface {
       AddCustomFlagsTo(*flag.FlagSet)
}


type genericAction struct {
	configFile           string
	inLineAuthentication inLineAuthentication
	downTheDrive         *downthedrive.DownTheDrive
	indexStrategy        int
	logLevel             int
	localStore	     *downthedrive.LocalStore
	localStorePath	     string
	drivePath	     string
	customFlagsFunc      CustomFlags
}

func getHomeDirectory() string {

	vHome := os.Getenv("HOME")

	if vHome != "" {
		return vHome
	}

	vHome = os.Getenv("APPDATA")

	if vHome != "" {
		return vHome
	}

	return ""
}

type BuildFlagSetOptions struct {
	LocalStore bool	
}

func (vSelf *genericAction) BuildFlagSet() *flag.FlagSet {
	vFlagSet := flag.NewFlagSet(os.Args[1], flag.ContinueOnError)
	vSelf.inLineAuthentication.AddParametersTo(vFlagSet)

	if vSelf.customFlagsFunc != nil {
		vSelf.customFlagsFunc.AddCustomFlagsTo(vFlagSet)
	}

	vFlagSet.StringVar(&vSelf.configFile, PARAMETER_CONFIGFILE, getHomeDirectory()+"/.downthedrive/config.json", "Configuration file")
	vFlagSet.IntVar(&vSelf.indexStrategy, PARAMETER_INDEXSTRATEGY, 0, "index strategy")
        vFlagSet.StringVar(&vSelf.drivePath, PARAMETER_DRIVEPATH, "/", "Base path for drive operations")
	vFlagSet.IntVar(&vSelf.logLevel, PARAMETER_LOGLEVEL, diagnostic.LogLevel_Info, "log level")

       	vFlagSet.StringVar(&vSelf.localStorePath, PARAMETER_LOCALSTOREPATH, getHomeDirectory()+"/.downthedrive/localStore", "Path of localStore")

	return vFlagSet    
}

func (vSelf *genericAction) CheckMandatoryParameters() error {
	return nil
}


func (vSelf *genericAction) Init() error {

	diagnostic.SetLogLevel(diagnostic.LogLevel(vSelf.logLevel))

	return nil
}

func (vSelf *genericAction) PostExecute() error {
	return nil
}

func (vSelf *genericAction) GetDownTheDrive() (*downthedrive.DownTheDrive,error) {

	if (vSelf.downTheDrive != nil ) {
		return vSelf.downTheDrive,nil
	}

	os.MkdirAll(getHomeDirectory()+"/.downthedrive", 0700)

	vConfig := &downthedrive.DownTheDriveConfig{}
	vConfigError := persistency.LoadBeanFromFile(vSelf.configFile, vConfig)

	if vConfigError != nil {
		if persistency.IsBeanNotFound(vConfigError) {
			diagnostic.LogWarning("genericAction.buildDownTheDrive", "Configuration file %s not found, initializing it with the default configuration...", nil, vSelf.configFile)

			vConfig = &downthedrive.DownTheDriveConfig{Db: downthedrive.DownTheDriveDbConfig{Driver: "sqlite3", DataSource: getHomeDirectory() + "/.downthedrive/downthedrive.db"}}

			vSaveError := persistency.SaveBeanIntoFile(vConfig, vSelf.configFile)

			if vSaveError != nil {
				return nil,diagnostic.NewError("Failed to save configuration file %s", vSaveError, vSelf.configFile)
			}
		} else {
			return nil,diagnostic.NewError("Failed to load downthedrive configuration file %s", vConfigError, vSelf.configFile)
		}
	}

	vDownTheDrive, vDownTheDriveError := downthedrive.NewDownTheDrive(*vConfig)

	if vDownTheDriveError != nil {
		return nil,diagnostic.NewError("Failed to initialize downthedrive", vDownTheDriveError)
	}

	vSelf.inLineAuthentication.ConfigureDownTheDrive(vDownTheDrive)

	vDownTheDrive.SetDefaultIndexStrategy(downthedrive.IndexStrategy(vSelf.indexStrategy))
	vSelf.downTheDrive = vDownTheDrive

	return vSelf.downTheDrive,nil
}

func (vSelf *genericAction) GetLocalStore() (*downthedrive.LocalStore,error) {

	if vSelf.localStore !=nil {
		return vSelf.localStore,nil
	}

	if vSelf.localStorePath == "" {
		return nil,diagnostic.NewError("Missing %s parameter",nil,PARAMETER_LOCALSTOREPATH)
	}
	
	if vSelf.drivePath == "" {
		return nil,diagnostic.NewError("Missing %s parameter",nil,PARAMETER_DRIVEPATH)
	}

	vDownTheDrive, vDownTheDriveError := vSelf.GetDownTheDrive()
	if vDownTheDriveError != nil {
		return nil,diagnostic.NewError("An error occurred while getting downthedrive", vDownTheDriveError)
	}

	vDriveBaseItem, vDriveBaseItemError := vDownTheDrive.GetItem(vSelf.drivePath,false,downthedrive.IndexStrategy_Default,nil)
	if vDriveBaseItemError != nil {
		return nil, diagnostic.NewError("Error getting base path %s from drive", vDriveBaseItemError, vSelf.drivePath)
	}
	
	vLocalStore,vLocalStoreError := downthedrive.NewLocalStore(vSelf.localStorePath,vDriveBaseItem,nil)
	if vLocalStoreError != nil {
		return nil, diagnostic.NewError("Failed to create local store",vLocalStoreError)
	}

	return vLocalStore,nil
}


func (vSelf *genericAction) PrintHelp() {
	vFlagSet := vSelf.BuildFlagSet()
	vFlagSet.PrintDefaults()
}

func (vSelf *genericAction) Close() error {

	var vError error
	if vSelf.localStore != nil {
	 	vError = vSelf.localStore.Close()	
		if vError !=nil {
			return diagnostic.NewError("An error occurred while closing localstore",vError)
		}
	}
	if vSelf.downTheDrive != nil {
	 	vError = vSelf.downTheDrive.Close()	
		if vError !=nil {
			return diagnostic.NewError("An error occurred while closing downthedrive",vError)
		}

	}

	return nil
}
