// run_hidden.js - Run a command in completely hidden mode
// Usage: wscript run_hidden.js "working_dir" "command" "log_file"
var shell = new ActiveXObject("WScript.Shell");
var fso = new ActiveXObject("Scripting.FileSystemObject");

var workDir = WScript.Arguments(0);
var cmd = WScript.Arguments(1);
var logFile = WScript.Arguments(2);

// Change to working directory
shell.CurrentDirectory = workDir;

// Run command with output redirected to log file
var fullCmd = "cmd /c " + cmd + " >> \"" + logFile + "\" 2>&1";
shell.Run(fullCmd, 0, false);
