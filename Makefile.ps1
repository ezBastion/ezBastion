

$r = $(split-path -parent $MyInvocation.MyCommand.Definition)
cd $r


Register-ArgumentCompleter -CommandName do-gobuild -ParameterName repo -ScriptBlock {
    $( Get-ChildItem -Directory .\cmd -Filter "ezb_*" ).Name | ForEach-Object {
        $Text = $_

        [System.Management.Automation.CompletionResult]::new(
                $Text,
                $_,
                'ParameterValue',
                "$_"
        )
    }
}


function do-gobuild  {
    param(
        [Parameter(Position = 0, Mandatory = $false )]
        #[ValidateSet({[String[]] $( Get-ChildItem -Directory .\cmd -Filter "ezb_*" ).Name })]
        [String]
        $repo
    )
    if ($repo) {
        upgrade-semver -file "./cmd/$($repo)" -appname $repo
        go fmt "./cmd/$($repo)"
        go generate "./cmd/$($repo)"
        go build -o "./bin" "./cmd/$($repo)"
    } else {
        foreach ($f in $(Get-ChildItem "./cmd/ezb_*" -Directory )) {
            upgrade-semver -file "./cmd/$($f.Name)" -appname $f.Name
            go fmt "./cmd/$($f.Name)"
            go generate "./cmd/$($f.Name)"
            go build -o "./bin" "./cmd/$($f.Name)"
        }
    }
}

function do-gozip {
    $allver = Get-Content "./bin/allver.json" -Raw | ConvertFrom-Json
    foreach ($f in $(Get-ChildItem "./bin/ezb_*" -Filter *.exe)) {
        Compress-Archive -Path ".\bin\$($f.Name)" -DestinationPath ".\bin\$($f.BaseName)-$($allver.$($f.BaseName)).zip" -CompressionLevel Optimal -Force
    }
    $(Get-ChildItem "./bin/ezb_*" -Filter *.zip).Name
}

function upgrade-semver {
    param ($file, $appname)
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
    $allver = New-Object -TypeName PSCustomObject
    if (Test-Path -Path "./bin/allver.json" ) {
        $allver = Get-Content "./bin/allver.json" -Raw | ConvertFrom-Json
    }
    $allver | Add-Member -Name $appname -Value "$($major).$($minor).$($patch).$($info.FixedFileInfo.FileVersion.Build)" -MemberType NoteProperty -Force
    $allver | ConvertTo-json -depth 10 | Out-File "./bin/allver.json" -Encoding ascii
}
if (!($MyInvocation.InvocationName -eq '.' -or $MyInvocation.Line -eq ''))
{
    Write-Host -ForegroundColor Green "dot sourceing this file: [. $( $MyInvocation.MyCommand.Definition )]"
    "build all: [do-gobuild]"
    "build one: [do-gobuild ezb_*]"
    "zip all ezBastion binary: [do-gozip]"
}
