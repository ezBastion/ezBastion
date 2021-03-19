param(
    [Parameter(Position = 0, Mandatory = $true )]
$action,
    [Parameter(Position = 1, Mandatory = $false )]
$repo
)

$RepoPath  = split-path -parent $MyInvocation.MyCommand.Definition
cd $RepoPath
$r = "..."
if ($repo) {
    $r = "$repo"
}

function invoke-generate {
    if ($repo) {
        upgrade-semver -file "./cmd/$($repo)"
        go generate "./cmd/$($repo)"  
    } else {
        foreach ($f in $(Get-ChildItem "./cmd" -Directory )) {
            upgrade-semver -file "./cmd/$($f.Name)"         
            go generate "./cmd/$($f.Name)"        
        }
    }
}

function invoke-build  {
    invoke-generate
    go build -o "./bin" "./cmd/$r"
}

function invoke-zip {
    foreach ($f in $(Get-ChildItem "./bin" -Filter *.exe)) {
        Compress-Archive -Path ".\bin\$($f.Name)" -DestinationPath ".\bin\$($f.BaseName).zip" -CompressionLevel Optimal -Force
    }

}

function show-help {
    "update version & build all => powershell Makefile.ps1 build"
    "update version & build one => powershell Makefile.ps1 build ezb_srv"
    "update all binary version  => powershell Makefile.ps1 generate"
    "update one binary version  => powershell Makefile.ps1 generate ezb_srv"
    "make zip                   => powershell Makefile.ps1 compress"

}

function upgrade-semver {
    param ($file)
    $ver = $(Select-String -Path "$file\main.go" -Pattern "VERSION.*""(\d\.\d.\d)"".*").Matches.Groups[1].Value
    $v = $ver.split(".")
    [int]$major = $v[0]
    [int]$minor = $v[1]
    [int]$patch = $v[2]
    $info = get-content "$file/versioninfo.json" | ConvertFrom-Json
    $info.FixedFileInfo.FileVersion.Major = $major
    $info.FixedFileInfo.FileVersion.Minor = $minor
    $info.FixedFileInfo.FileVersion.Patch = $patch
    $info.FixedFileInfo.FileVersion.Build ++
    $info.FixedFileInfo.ProductVersion = $info.FixedFileInfo.FileVersion
    $info.StringFileInfo.FileVersion = "v$($major).$($minor).$($patch).$($info.FixedFileInfo.FileVersion.Build)"
    $info.StringFileInfo.ProductVersion = $info.StringFileInfo.FileVersion
    $info | ConvertTo-json -depth 100 | Out-File "$file/versioninfo.json" -Encoding ascii
}

Switch ($action)
{
    generate { invoke-generate }
    build { invoke-build }
    compress { invoke-zip }
    default { show-help  }
}