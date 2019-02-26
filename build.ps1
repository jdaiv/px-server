$exeName = "server.exe"

mkdir -Force build | out-null

go build -o "build/$exeName";
if ($LASTEXITCODE -eq 0) {
    if (Test-Path .\build\resources) {
        Remove-Item -Recurse -Force .\build\resources
    }
    Copy-Item -Recurse .\resources .\build
    Push-Location build
    try {
        & ".\$exeName"
    } finally {
        Pop-Location
    }
}