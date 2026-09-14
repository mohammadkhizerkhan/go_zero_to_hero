Here is the complete, comprehensive technical manual of how we recovered and built your optimized environment, including every command with inline comments explaining exactly what it does.
Phase 1: Force-Killing & Resetting the Frozen WSL Engine
When your background WSL processes hung completely and blocked standard utilities, we manually targeted the task manager processes directly to force a full environment reset.
powershell

# Run this inside Windows PowerShell as Administrator:

# 1. Forcefully terminate the hanging core WSL background service host
Stop-Process -Name "wslservice" -Force -ErrorAction SilentlyContinue

# 2. Search for the legacy Linux Instance Manager service host and force-kill its process ID
Get-WmiObject Win32_Service | Where-Object {$_.Name -eq "LxssManager"} | ForEach-Object { Stop-Process -Id $_.ProcessId -Force }

# 3. Completely unregister, wipe, and delete the corrupted Ubuntu distribution data
wsl --unregister Ubuntu

# 4. Force a clean, official update to download the absolute latest WSL kernel binaries
wsl --update
Use code with caution.



Phase 2: Installing the Core Linux Toolchain (Natively in Ubuntu)
After opening your clean Orange Ubuntu Profile terminal, we updated the packages and compiled your runtime environments inside the isolated Linux filesystem for max performance.
bash

# Run these inside your Ubuntu Terminal:

# 1. Update your Linux package listings and install core C/C++ compilation tools and Git
sudo apt update && sudo apt upgrade -y
sudo apt install git curl build-essential -y

# 2. Download the official, native Linux Go compiler version 1.27.1
wget https://go.dev

# 3. Erase older Go instances and extract the clean binaries directly into your system files
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.27.1.linux-amd64.tar.gz

# 4. Permanently append the Go application binary folder to your user shell environment PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# 5. Download the installation script for Node Version Manager (NVM)
curl -o- https://githubusercontent.com | bash

# 6. Force Ubuntu to reload its profile settings to activate NVM and Go pathways immediately
source ~/.bashrc

# 7. Use NVM to download and manage the absolute latest Long Term Support version of Node.js and NPM
nvm install --lts

# 8. Clean up and delete the heavy Go archive package file to free up Linux hard drive space
rm go1.27.1.linux-amd64.tar.gz
Use code with caution.



Phase 3: Fixing Docker Desktop Integration & Permission Errors
We resolved the WSL_E_DISTRO_NOT_FOUND registration engine glitch, linked the Windows Docker Desktop UI to the Ubuntu engine, and eliminated permissions errors.
powershell

# Run these inside Windows PowerShell as Administrator (to clean up the old cache):
wsl --shutdown
wsl --unregister docker-desktop
wsl --unregister docker-desktop-data

# [Action Step]: Deleted %AppData%\Docker\settings.json and launched Docker Desktop as Administrator.
# [Action Step]: Navigated to Settings > Resources > WSL Integration > Toggled "Ubuntu" ON.
Use code with caution.



bash

# Run these inside your Ubuntu Terminal to eliminate the "Permission Denied" socket error:

# 1. Create a system-recognized security group named 'docker' (if it doesn't already exist)
sudo groupadd docker

# 2. Permanently attach your current active user account ($USER) to the secure docker system group
sudo usermod -aG docker $USER

# 3. Fully log out and log back in to your active user session to refresh active group permissions
exec su - $USER

# 4. Pull and run a lightweight template container container to verify Docker functions cleanly 
docker run hello-world
Use code with caution.



Phase 4: Constructing the Cryptographic GitHub SSH Tunnel
We generated safe modern keys and mapped them into your GitHub account so you could securely execute repository adjustments without typing out account credentials.
bash

# Run these inside your Ubuntu Terminal:

# 1. Generate an enterprise-grade Ed25519 cryptographic authentication key pair tagged to your email
ssh-keygen -t ed25519 -C "your.email@example.com"

# 2. Boot up the local SSH keys manager daemon tracking engine inside your background processes
eval "$(ssh-agent -s)"

# 3. Register your newly minted highly private identity key directly inside the background manager
ssh-add ~/.ssh/id_ed25519

# 4. View and display your PUBLIC key on screen so you can copy-paste it cleanly into github.com
cat ~/.ssh/id_ed25519.pub

# 5. Handshake test to verify if GitHub safely recognizes your encryption keys
ssh -T git@github.com
Use code with caution.



Phase 5: Bridging VS Code & The Linux File System Rule
To avoid path issues where your terminal throws command not found: code, we mapped the path explicitly and deployed the workspace configuration.
bash

# Run this inside your Ubuntu Terminal:

# 1. Map the specific Windows location of the VS Code execution engine directly into Ubuntu's workspace environment
echo 'export PATH=$PATH:"/mnt/c/Users/khizerkhan/AppData/Local/Programs/Microsoft VS Code/bin"' >> ~/.bashrc
source ~/.bashrc

# 2. Move directly into your high-speed Linux ecosystem directory (Avoid coding inside /mnt/c/ or /mnt/e/!)
cd ~/go_zero_to_hero/calculator

# 3. Launch VS Code windows directly, linking it natively via the green "WSL: Ubuntu" remote badge
code .
Use code with caution.



Everything is completely set up and fully working. If you'd like, let me know:
Do you want to run go mod init to start building your Go calculator?
Are there any other programming libraries or tools you need to add to this environment?





