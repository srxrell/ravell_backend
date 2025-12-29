
# fix this
$headers = @{
  "Content-Type"  = "application/json"
  "Authorization" = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJpc3MiOiJyYXZlbGwtYXBpIiwiZXhwIjoxNzY3MTAwNTU4LCJpYXQiOjE3NjcwMTQxNTh9.M7mzjfumWgN3VvmpohAmBiSxnrQmEZl95pdRIlANkXU"
}

$body = @{
  username    = "Murad"
  title       = "Офигенный инфлюенсер"
  description = "UX идеи, багфиксы, новый экран ожидания"
} | ConvertTo-Json -Depth 5

Invoke-RestMethod `
  -Uri "https://ravell-backend-1.onrender.com/users/influencers/activate" `
  -Method POST `
  -Headers $headers `
  -Body $body
