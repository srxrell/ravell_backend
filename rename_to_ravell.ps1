$body = @{
    key = "first_story"
    title = "Первооткрыватель"
    description = "Опубликуй первую историю"
    icon = "https://.../icon.png"
    condition = @{ type="story_count"; operator=">="; value=1 }
} | ConvertTo-Json

Invoke-RestMethod -Uri "https://ravell-backend-1.onrender.com/achievements/create" `
                   -Method POST `
                   -ContentType "application/json" `
                   -Body $body