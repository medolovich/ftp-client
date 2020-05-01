package ftpclient

const (
	ftpResponseCodePartEndIndex      = 3
	ftpResponseMessagePartStartIndex = 4

	newLineChar = "\n" // cross-platforming will die here
	tabChar     = "\t"

	statusCodeAboutToOpenConnection    = "150"
	statusCodeActionSuccess            = "200"
	statusCodeReadyForNewUser          = "220"
	statusCodeClosingControlConnection = "221"
	statusCodeClosingDataConnection    = "226"
	statusCodeEnterPassiveMode         = "227"
	statusCodeLoginSuccess             = "230"
	statusCodeRequestFileActionOK      = "250"
	statusCodePathnameCreated          = "257"
	statusCodeNeedPassword             = "331"
	statusCodeRequestFilePending       = "350"

	packetSize = 0x20

	newLineLinuxASCII   = 0x0a
	newLineWindowsASCII = 0x85
)
