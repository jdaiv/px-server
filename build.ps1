$exeName = "server.exe"

mkdir -Force build | out-null

go build -o $exeName;
if ($LASTEXITCODE -eq 0) {
    rm "build\$exeName"
    mv $exeName build\
    & "build\$exeName"
}