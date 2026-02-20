' Launcher script - Runs start.bat and auto-closes when done
' The main window will be visible, and will close automatically after start.bat finishes

Set WshShell = CreateObject("WScript.Shell")
Set fso = CreateObject("Scripting.FileSystemObject")

' Get the directory where this script is located
scriptDir = fso.GetParentFolderName(WScript.ScriptFullName)

' Run start.bat with visible window (1), wait for it to complete (True)
' When start.bat exits, this script will also exit, closing the window
WshShell.Run chr(34) & scriptDir & "\start.bat" & Chr(34), 1, True
