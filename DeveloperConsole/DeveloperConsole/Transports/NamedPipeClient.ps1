param(
    [string]$PipeName = "MyDevConsole"
)

Write-Host "Connecting to \\.\pipe\$PipeName..." -ForegroundColor Cyan

try
{
    $pipe = New-Object System.IO.Pipes.NamedPipeClientStream(".", $PipeName, [System.IO.Pipes.PipeDirection]::Out)
    $pipe.Connect(3000)

    $writer = New-Object System.IO.StreamWriter($pipe)
    $writer.AutoFlush = $true

    Write-Host "Connected. Type 'exit' to quit.`n" -ForegroundColor Green

    while ($true)
    {
        $input = Read-Host ">"

        if ($input -eq "exit") { break }

        $writer.Write($input)
    }
}
catch [TimeoutException]
{
    Write-Host "Error: Could not connect to pipe (timeout). Is the server running?" -ForegroundColor Red
}
catch
{
    Write-Host "Error: $_" -ForegroundColor Red
}
finally
{
    if ($pipe) { $pipe.Dispose() }
    Write-Host "Disconnected." -ForegroundColor Cyan
}
