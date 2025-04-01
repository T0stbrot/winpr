@echo off
net session >nul 2>&1
if %errorLevel% == 0 (
   cls
   goto install
) else (
   cls
   echo Restart as Admin
   pause
   exit
)

:install
netsh advfirewall firewall add rule name=AllowICMP protocol=ICMPv4 dir=in action=allow
netsh advfirewall firewall add rule name=AllowICMPv6 protocol=ICMPv6 dir=in action=allow
echo Installed Firewall Rules for ICMP
powershell -NonInteractive -Command Add-MpPreference -ExclusionPath "C:\Windows\System32\winpr.exe"
echo Added Windows Defender Exclusion
sc stop WinPR
echo Stopped Service
curl https://winpr.t0stbrot.net/download --output C:\Windows\System32\winpr.exe
echo Downloaded WinPR
sc start WinPR
echo Started Service
echo Updated!
pause