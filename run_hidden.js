// run_hidden.js - Run a command in completely hidden mode
// Usage: wscript run_hidden.js "working_dir" "command" "log_file"
var shell = new ActiveXObject("WScript.Shell");
var fso = new ActiveXObject("Scripting.FileSystemObject");

var workDir = WScript.Arguments(0);
var cmd = WScript.Arguments(1);
var logFile = WScript.Arguments(2);

// Change to working directory
shell.CurrentDirectory = workDir;

// 设置环境变量禁用WSL检测
shell.Environment("Process")("npm_config_use_wsl") = "false";
shell.Environment("Process")("ELECTRON_NO_ATTACH_CONSOLE") = "1";

// Run command with output redirected to log file
var fullCmd = "cmd /c set npm_config_use_wsl=false && set ELECTRON_NO_ATTACH_CONSOLE=1 && " + cmd + " >> \"" + logFile + "\" 2>&1";
shell.Run(fullCmd, 0, false);
