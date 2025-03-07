# winpr
Go Program that runs as a Probe Service for my Services

# Building it yourself
1. Install Go: https://go.dev/dl/
2. Download Source of latest release and unzip it
3. Open a CMD Window inside the folder
4. Run `go build .`
5. Copy the `winpr.exe` file to C:\Windows\System32\
6. Open a CMD Windows as Administrator
7. Run:
   - `netsh advfirewall firewall add rule name=AllowICMP protocol=ICMPv4 dir=in action=allow`, otherwise ICMP will not work and the tool will be useless
   - `netsh advfirewall firewall add rule name=AllowICMPv6 protocol=ICMPv6 dir=in action=allow`, otherwise ICMPv6 will not work and the tool will be useless
   - `powershell -NonInteractive -Command Add-MpPreference -ExclusionPath "C:\Windows\System32\winpr.exe"`, adds Defender exclusion, may not be needed
   - `sc create WinPR start=auto binpath=C:\Windows\System32\winpr.exe`, creates Service that autostarts
   - `sc start WinPR`, starts the Service
## Done, it is installed now


# Building with Garble (Less-Likely Antivirus Detection)
1. Install Git
2. Install `garble` using `go install mvdan.cc/garble@latest` in a CMD Windows
3. Instead of using `go build .` use `garble -tiny build`
4. Rest of the Process if the same as Building normally
