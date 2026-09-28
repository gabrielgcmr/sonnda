# scripts/extract-lab-summary-picker.ps1
$ErrorActionPreference = "Stop"

Add-Type -AssemblyName System.Windows.Forms

$picker = New-Object System.Windows.Forms.OpenFileDialog
$picker.Title = "Selecione um laudo laboratorial"
$picker.Filter = "Laudos laboratoriais|*.pdf;*.jpg;*.jpeg;*.png|Todos os arquivos|*.*"
$picker.Multiselect = $false

if ($picker.ShowDialog() -ne [System.Windows.Forms.DialogResult]::OK) {
	exit 0
}

$inputPath = $picker.FileName
$fileName = [System.IO.Path]::GetFileNameWithoutExtension($inputPath)
$outputDirectory = Split-Path -Parent $inputPath
$outputPath = Join-Path $outputDirectory "${fileName}_resumo.txt"
$suffix = 2
while (Test-Path -LiteralPath $outputPath) {
	$outputPath = Join-Path $outputDirectory "${fileName}_resumo_${suffix}.txt"
	$suffix++
}

$temporaryDirectory = Join-Path ([System.IO.Path]::GetTempPath()) "sonnda-lab-summary-$([guid]::NewGuid().ToString('N'))"

try {
	New-Item -ItemType Directory -Path $temporaryDirectory | Out-Null
	& go run ./cmd/extract-text -input $inputPath -documentai -output $temporaryDirectory
	if ($LASTEXITCODE -ne 0) {
		throw "Document AI OCR failed with exit code $LASTEXITCODE"
	}

	$ocrTextPath = Join-Path $temporaryDirectory "${fileName}.txt"
	if (-not (Test-Path -LiteralPath $ocrTextPath)) {
		throw "OCR output was not created: $ocrTextPath"
	}

	& go run ./cmd/extract-lab-summary -input $ocrTextPath -output $outputPath
	if ($LASTEXITCODE -ne 0) {
		throw "Lab summary extraction failed with exit code $LASTEXITCODE"
	}

	Write-Host "`nResumo:`n"
	Get-Content -LiteralPath $outputPath -Encoding UTF8 -Raw
	Write-Host "`nResumo salvo em: $outputPath"
}
catch {
	[Console]::Error.WriteLine($_.Exception.Message)
	exit 1
}
finally {
	if (Test-Path -LiteralPath $temporaryDirectory) {
		Remove-Item -LiteralPath $temporaryDirectory -Recurse -Force
	}
}