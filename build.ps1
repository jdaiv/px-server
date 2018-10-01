$exeName = "server.exe"

mkdir -Force build | out-null

go build -o "build/$exeName";
if ($LASTEXITCODE -eq 0) {
    Push-Location build
    try {
        & ".\$exeName"
    } finally {
        Pop-Location
    }
}