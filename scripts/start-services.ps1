$services = @("gateway", "orders", "stock", "payments", "kitchen", "cart", "notification")

Write-Host "Syncing .env to all microservice directories..."
foreach ($s in $services) {
    if (Test-Path ".env") {
        Copy-Item ".env" "$s/.env" -Force
    }
}

Write-Host "Starting microservices in individual windows..."
foreach ($s in $services) {
    Write-Host " -> Launching $s..."
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd $s; Write-Host '=== $s Microservice ===' -ForegroundColor Cyan; go run ."
}

Write-Host "All backend microservices initiated."
