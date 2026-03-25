[Setup]
; 应用基础信息
AppName=TraeProxy
AppVersion=1.1.0
AppPublisher=TraeProxy Team
AppPublisherURL=https://trae.agentlab.click/
AppSupportURL=https://trae.agentlab.click/

; 默认安装路径，自动识别 x86/x64，安装至 Program Files
DefaultDirName={autopf}\TraeProxy
DefaultGroupName=TraeProxy

; 输出设置
OutputDir=release
OutputBaseFilename=TraeProxy_Setup_v1.1.0
Compression=lzma2/ultra64
SolidCompression=yes

; 挂载桌面和运行相关图标
SetupIconFile=build\windows\icon.ico
UninstallDisplayIcon={app}\TraeProxy.exe

; 强制只允许安装 64 位 （如果 Wails build 出的是 64 位）
ArchitecturesInstallIn64BitMode=x64

[Languages]
Name: "chinesesimp"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"

[Tasks]
Name: "desktopicon"; Description: "在桌面创建 TraeProxy 快捷方式"; GroupDescription: "附加快捷方式:"

[Files]
; 将 Wails 编译的核心可执行文件压包
Source: "build\bin\TraeProxy.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
; 生成开始菜单
Name: "{group}\TraeProxy"; Filename: "{app}\TraeProxy.exe"
Name: "{group}\卸载 TraeProxy"; Filename: "{uninstallexe}"
; 生成桌面图标
Name: "{autodesktop}\TraeProxy"; Filename: "{app}\TraeProxy.exe"; Tasks: desktopicon

[Run]
; 安装完成后是否自动运行
Filename: "{app}\TraeProxy.exe"; Description: "立刻运行 TraeProxy"; Flags: nowait postinstall skipifsilent
