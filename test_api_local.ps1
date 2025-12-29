$headers = @{
  "Content-Type"  = "application/json; charset=utf-8"
  "Authorization" = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJpc3MiOiJyYXZlbGwtYXBpIiwiZXhwIjoxNzY3MTAwNTU4LCJpYXQiOjE3NjcwMTQxNTh9.M7mzjfumWgN3VvmpohAmBiSxnrQmEZl95pdRIlANkXU"
}


$body = @{
  user_id = 9
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "https://ravell-backend-1.onrender.com/users/achievements/grant_influential" `
  -Method POST `
  -Headers $headers `
  -Body $body