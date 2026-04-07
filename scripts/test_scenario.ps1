# TaskManager - Business Scenario Test Script
# Scenario: E-commerce Order Processing System
#
# Pipeline:
# 1. Order confirmation emails (send_email) - succeed
# 2. Shipping notification webhooks (send_webhook) - succeed
# 3. Payment receipt emails (send_email) - succeed
# 4. Inventory sync webhooks to warehouse API - FAIL (localhost:9999 is down)
#
# Monitoring insight: Grafana reveals spike in failed tasks,
# all send_webhook with priority 1. Error: "connection refused" to warehouse endpoint.

$API = "http://localhost:8081/api/v1/tasks"

function Send-TaskRequest {
    param([string]$Title, [string]$TaskType, [string]$Payload, [string]$Cron, [int]$Priority)
    $obj = [ordered]@{ title = $Title; type = $TaskType; payload = $Payload; cron_expr = $Cron; priority = $Priority }
    $json = $obj | ConvertTo-Json -Compress
    try {
        $r = Invoke-RestMethod -Uri $API -Method POST -Body $json -ContentType "application/json"
        Write-Host "[OK] $Title (id=$($r.id))" -ForegroundColor Green
    } catch {
        Write-Host "[FAIL] $Title - $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host " E-Commerce Order Processing - Test Scenario" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

# --- Phase 1: Order confirmation emails (will succeed) ---
Write-Host "--- Phase 1: Order Confirmation Emails ---" -ForegroundColor Yellow
Send-TaskRequest -Title "Order #1001 Confirmation" -TaskType "send_email" -Payload '{"to":"customer1@shop.com","from":"orders@myshop.com","body":"Order #1001 confirmed! Total: 59.99"}' -Cron "*/2 * * * *" -Priority 2
Send-TaskRequest -Title "Order #1002 Confirmation" -TaskType "send_email" -Payload '{"to":"customer2@shop.com","from":"orders@myshop.com","body":"Order #1002 confirmed! Total: 124.50"}' -Cron "*/2 * * * *" -Priority 2
Send-TaskRequest -Title "Order #1003 Confirmation" -TaskType "send_email" -Payload '{"to":"customer3@shop.com","from":"orders@myshop.com","body":"Order #1003 confirmed! Total: 34.99"}' -Cron "*/2 * * * *" -Priority 2
Send-TaskRequest -Title "VIP Order #1004 Confirmation" -TaskType "send_email" -Payload '{"to":"vip@shop.com","from":"orders@myshop.com","body":"VIP order #1004 confirmed! Total: 899.00"}' -Cron "*/2 * * * *" -Priority 1
Write-Host ""

# --- Phase 2: Shipping notifications (will succeed via httpbin) ---
Write-Host "--- Phase 2: Shipping Notifications ---" -ForegroundColor Yellow
Send-TaskRequest -Title "Shipping: Order #1001 DHL" -TaskType "send_webhook" -Payload '{"url":"https://httpbin.org/post","method":"POST","headers":{"X-Api-Key":"logistics-123"},"body":{"order_id":"1001","carrier":"DHL","tracking":"DHL-98765"}}' -Cron "*/3 * * * *" -Priority 3
Send-TaskRequest -Title "Shipping: Order #1002 FedEx" -TaskType "send_webhook" -Payload '{"url":"https://httpbin.org/post","method":"POST","headers":{"X-Api-Key":"logistics-123"},"body":{"order_id":"1002","carrier":"FedEx","tracking":"FDX-54321"}}' -Cron "*/3 * * * *" -Priority 3
Send-TaskRequest -Title "Shipping: Order #1003 UPS" -TaskType "send_webhook" -Payload '{"url":"https://httpbin.org/post","method":"POST","headers":{"X-Api-Key":"logistics-123"},"body":{"order_id":"1003","carrier":"UPS","tracking":"UPS-11223"}}' -Cron "*/3 * * * *" -Priority 3
Write-Host ""

# --- Phase 3: Payment receipts (will succeed) ---
Write-Host "--- Phase 3: Payment Receipts ---" -ForegroundColor Yellow
Send-TaskRequest -Title "Payment Receipt: Order #1001" -TaskType "send_email" -Payload '{"to":"customer1@shop.com","from":"billing@myshop.com","body":"Payment of 59.99 received for order #1001. Receipt #R-5001"}' -Cron "*/2 * * * *" -Priority 2
Send-TaskRequest -Title "Payment Receipt: Order #1002" -TaskType "send_email" -Payload '{"to":"customer2@shop.com","from":"billing@myshop.com","body":"Payment of 124.50 received for order #1002. Receipt #R-5002"}' -Cron "*/2 * * * *" -Priority 2
Send-TaskRequest -Title "Payment Receipt: VIP #1004" -TaskType "send_email" -Payload '{"to":"vip@shop.com","from":"billing@myshop.com","body":"Payment of 899.00 received for VIP order #1004. Receipt #R-5004"}' -Cron "*/2 * * * *" -Priority 1
Write-Host ""

# --- Phase 4: THE BUG - Warehouse inventory sync (WILL FAIL!) ---
# Warehouse API at localhost:9999 is DOWN - simulates microservice outage
Write-Host "--- Phase 4: Inventory Sync - WAREHOUSE DOWN! (will fail) ---" -ForegroundColor Red
Send-TaskRequest -Title "[CRITICAL] Inventory Sync: Order #1001" -TaskType "send_webhook" -Payload '{"url":"http://localhost:9999/warehouse/sync","method":"POST","headers":{"X-Api-Key":"warehouse-secret"},"body":{"order_id":"1001","items":[{"sku":"SHIRT-BLU-M","qty":-1}]}}' -Cron "*/1 * * * *" -Priority 1
Send-TaskRequest -Title "[CRITICAL] Inventory Sync: Order #1002" -TaskType "send_webhook" -Payload '{"url":"http://localhost:9999/warehouse/sync","method":"POST","headers":{"X-Api-Key":"warehouse-secret"},"body":{"order_id":"1002","items":[{"sku":"LAPTOP-PRO-15","qty":-1}]}}' -Cron "*/1 * * * *" -Priority 1
Send-TaskRequest -Title "[CRITICAL] Inventory Sync: Order #1003" -TaskType "send_webhook" -Payload '{"url":"http://localhost:9999/warehouse/sync","method":"POST","headers":{"X-Api-Key":"warehouse-secret"},"body":{"order_id":"1003","items":[{"sku":"HEADPHONES-WL","qty":-2}]}}' -Cron "*/1 * * * *" -Priority 1
Send-TaskRequest -Title "[CRITICAL] Inventory Sync: VIP #1004" -TaskType "send_webhook" -Payload '{"url":"http://localhost:9999/warehouse/sync","method":"POST","headers":{"X-Api-Key":"warehouse-secret"},"body":{"order_id":"1004","items":[{"sku":"WATCH-GOLD","qty":-1},{"sku":"CASE-LEATHER","qty":-1}]}}' -Cron "*/1 * * * *" -Priority 1
Send-TaskRequest -Title "[CRITICAL] Daily Inventory Reconciliation" -TaskType "send_webhook" -Payload '{"url":"http://localhost:9999/warehouse/sync","method":"POST","headers":{"X-Api-Key":"warehouse-secret"},"body":{"order_id":"BATCH","action":"daily_reconciliation"}}' -Cron "*/1 * * * *" -Priority 1
Write-Host ""

# --- Phase 5: Routine operations (will succeed) ---
Write-Host "--- Phase 5: Routine Operations ---" -ForegroundColor Yellow
Send-TaskRequest -Title "Daily Sales Report" -TaskType "send_email" -Payload '{"to":"admin@myshop.com","from":"system@myshop.com","body":"Daily sales report: 4 orders, revenue 1118.48"}' -Cron "*/5 * * * *" -Priority 5
Send-TaskRequest -Title "Customer Feedback Digest" -TaskType "send_email" -Payload '{"to":"support@myshop.com","from":"system@myshop.com","body":"Feedback summary: 12 new reviews, avg 4.5/5"}' -Cron "*/5 * * * *" -Priority 7
Send-TaskRequest -Title "Analytics: Daily Metrics" -TaskType "send_webhook" -Payload '{"url":"https://httpbin.org/post","method":"POST","headers":{"Authorization":"Bearer analytics-token"},"body":{"event":"daily_metrics","orders":4,"revenue":1118.48}}' -Cron "*/5 * * * *" -Priority 6
Write-Host ""

Write-Host "=============================================" -ForegroundColor Cyan
Write-Host " Scenario loaded!" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Healthy tasks (will complete):       13" -ForegroundColor Green
Write-Host "    - 7x send_email  (confirmations, receipts, reports)"
Write-Host "    - 6x send_webhook (shipping, analytics via httpbin.org)"
Write-Host ""
Write-Host "  Failing tasks (warehouse outage):     5" -ForegroundColor Red
Write-Host "    - 5x send_webhook to localhost:9999 (connection refused)"
Write-Host "    - All priority 1 (critical) - will retry and fail"
Write-Host ""
Write-Host "  What to look for in monitoring:" -ForegroundColor Magenta
Write-Host "    1. Grafana: spike in failed tasks, all send_webhook type"
Write-Host "    2. Grafana: tasks_processed_by_priority shows priority 1 failures"
Write-Host "    3. Web GUI: filter by 'failed' status - all are inventory sync"
Write-Host "    4. Worker logs: 'connection refused' to localhost:9999"
Write-Host "    5. Root cause: warehouse microservice is DOWN"
Write-Host ""
