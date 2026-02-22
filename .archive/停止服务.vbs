' Launcher script - Runs stop.bat and auto-closes when done
' The main window will be visible, and will close automatically after stop.bat finishes

Set WshShell = CreateObject("WScript.Shell")
Set fso = CreateObject("Scripting.FileSystemObject")

' Get the directory where this script is located
scriptDir = fso.GetParentFolderName(WScript.ScriptFullName)

' Run stop.bat with visible window (1), wait for it to complete (True)
' When stop.bat exits, this script will also exit, closing the window
WshShell.Run chr(34) & scriptDir & "\stop.bat" & Chr(34), 1, True
