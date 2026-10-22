Write-Host "Stopping all running microservice processes..." -ForegroundColor Yellow

# 1. Kill any processes listening on service ports
$ports = @(8080, 2000, 2001, 2002, 2003, 2005, 2006)
foreach ($port in $ports) {
    $conns = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    if ($conns) {
        $pids = $conns | Select-Object -ExpandProperty OwningProcess -Unique
        foreach ($pidToKill in $pids) {
            if ($pidToKill -gt 0) {
                Write-Host " -> Stopping process $pidToKill listening on port $port..."
                Stop-Process -Id $pidToKill -Force -ErrorAction SilentlyContinue
            }
        }
    }
}

# 2. Kill any PowerShell windows opened by start-services.ps1
$windows = Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -like "*Microservice ===*" }
foreach ($win in $windows) {
    Write-Host " -> Closing service terminal window (PID $($win.ProcessId))..."
    Stop-Process -Id $win.ProcessId -Force -ErrorAction SilentlyContinue
}

# 3. Kill any lingering go.exe or microservice binaries
$services = @("gateway", "orders", "stock", "payments", "kitchen", "cart", "notification")
foreach ($s in $services) {
    Get-Process -Name $s -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
}
Get-Process -Name "omsv2-*" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

Write-Host "All microservices stopped and ports freed." -ForegroundColor Green
