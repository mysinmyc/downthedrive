package cmdactions

type ActionInit struct {
	genericAction
	drivePath string
	threads   int
}

func (vSelf *ActionInit) ParseCmdLine(pArgs []string) error {

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

func (vSelf *ActionInit) Execute() error {
	return nil
}

func (vSelf *ActionInit) GetDescription() string {
	return "Initialize down the drive configuration if not present"
}
