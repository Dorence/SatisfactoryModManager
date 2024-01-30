Unicode true

!addplugindir /x86-ansi ".\NsisMultiUser\Plugins\x86-ansi\"
!addplugindir /x86-unicode ".\NsisMultiUser\Plugins\x86-unicode\"
!addincludedir ".\NsisMultiUser\Include\"

!define UNINST_KEY_NAME "${INFO_PRODUCTNAME}"
!define REQUEST_EXECUTION_LEVEL "user"
!include "wails_tools.nsh"

# Convert wails defines to common names
!define PRODUCT_NAME "${INFO_PRODUCTNAME}"
!define VERSION "${INFO_PRODUCTVERSION}"
!define PROGEXE "${PRODUCT_EXECUTABLE}"
!define COMPANY_NAME "${INFO_COMPANYNAME}"
!define URL_INFO_ABOUT "https://github.com/satisfactorymodding/SatisfactoryModManager"
!define UNINSTALL_FILENAME "uninstall.exe"

!define MULTIUSER_INSTALLMODE_DISPLAYNAME "${PRODUCT_NAME}"
!define MULTIUSER_INSTALLMODE_ALLOW_ELEVATION_IF_SILENT 1 # We use silent mode when updating
!define MULTIUSER_INSTALLMODE_64_BIT 1
!include "NsisMultiUser.nsh"

BrandingText "${INFO_PRODUCTNAME} ${INFO_PRODUCTVERSION}"

# The version information for this two must consist of 4 parts
!include "vi_version.nsh"
VIProductVersion "${VI_VERSION}"
VIFileVersion    "${VI_VERSION}"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

!include "MUI2.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

!include "smm2_uninstall.nsh"
!include "utils.nsh"

# Install pages
!define MUI_PAGE_CUSTOMFUNCTION_PRE WelcomePagePre
!insertmacro MUI_PAGE_WELCOME

!insertmacro MULTIUSER_PAGE_INSTALLMODE

!define MUI_PAGE_CUSTOMFUNCTION_PRE DirectoryPagePre
!insertmacro MUI_PAGE_DIRECTORY

!insertmacro MUI_PAGE_INSTFILES

!define MUI_FINISHPAGE_SHOWREADME ""
!define MUI_FINISHPAGE_SHOWREADME_NOTCHECKED
!define MUI_FINISHPAGE_SHOWREADME_TEXT "Create Desktop Shortcut"
!define MUI_FINISHPAGE_SHOWREADME_FUNCTION desktopShortcut
!insertmacro MUI_PAGE_FINISH

# Uninstall pages

!define MUI_UNABORTWARNING ; Show a confirmation when cancelling the installation

; Pages
!insertmacro MULTIUSER_UNPAGE_INSTALLMODE

!insertmacro MUI_UNPAGE_INSTFILES

# Language config
!insertmacro MUI_LANGUAGE "English"
!insertmacro MULTIUSER_LANGUAGE_INIT

# Reserve files
!insertmacro MUI_RESERVEFILE_LANGDLL

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\Satisfactory-Mod-Manager-Setup.exe"
ShowInstDetails show

Function .onInit
    SetRegView 64
    ; The original wails.checkArchitecture macro adds an unnecessary requirement on Windows 10
    ; !insertmacro wails.checkArchitecture
    !insertmacro MULTIUSER_INIT
FunctionEnd

Function un.onInit
  SetRegView 64
  !insertmacro MULTIUSER_UNINIT
FunctionEnd

Section
    !insertmacro wails.webview2runtime

    !insertmacro smm2Uninst

    Call EnsureEmptyFolder

    SetOutPath $INSTDIR
    
    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    WriteUninstaller "$INSTDIR\${UNINSTALL_FILENAME}"

    !insertmacro MULTIUSER_RegistryAddInstallInfo
    !insertmacro MULTIUSER_RegistryAddInstallSizeInfo
SectionEnd

Section "-Post"
	${GetParameters} $R0
	${GetOptions} $R0 "/ForceRun" $R1
    ${IfNot} ${errors}
        Exec '"$INSTDIR\${PRODUCT_EXECUTABLE}"'
    ${EndIf}
SectionEnd

Section "uninstall"
    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro MULTIUSER_RegistryRemoveInstallInfo

    Delete "$INSTDIR\${UNINSTALL_FILENAME}"
    RMDir "$INSTDIR"
SectionEnd

Function EnsureEmptyFolder
    ${If} $MultiUser.InstallMode == "AllUsers"
    ${AndIf} $HasPerMachineInstallation = 1
        Return # Already installed, so just use existing install dir
    ${EndIf}
    ${If} $MultiUser.InstallMode == "CurrentUser"
    ${AndIf} $HasPerUserInstallation = 1
        Return # Already installed, so just use existing install dir
    ${EndIf}
    ${If} ${FileExists} "$InstDir\*"
        Push $INSTDIR
        Call isEmptyDir
        Pop $0
        StrCmp $0 0 0 +2
        StrCpy $InstDir "$INSTDIR\${MULTIUSER_INSTALLMODE_INSTDIR}"
    ${EndIf}
FunctionEnd

Function desktopShortcut
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
FunctionEnd

Function WelcomePagePre
    ${if} $InstallShowPagesBeforeComponents == 0
        Abort ; don't display the Welcome page for the inner instance
    ${endif}
FunctionEnd

Function DirectoryPagePre
    ${If} $MultiUser.InstallMode == "AllUsers"
    ${AndIf} $HasPerMachineInstallation = 1
        Abort # Already installed, so just use existing install dir
    ${EndIf}
    ${If} $MultiUser.InstallMode == "CurrentUser"
    ${AndIf} $HasPerUserInstallation = 1
        Abort # Already installed, so just use existing install dir
    ${EndIf}
FunctionEnd

Function .onVerifyInstDir    
    var /GLOBAL currentDir
    StrCpy $currentDir $INSTDIR
    
    Check:
    IfFileExists $currentDir\FactoryGame.exe GameExists
    IfFileExists $currentDir\FactoryServer.exe GameExists
    IfFileExists $currentDir\FactoryServer.sh GameExists
    ${GetParent} $currentDir $currentDir
    StrCmp $currentDir "" 0 Check
    
    Return
    
    GameExists:
    Abort "SatisfactoryModManager should not be installed in the Satisfactory directory."
FunctionEnd