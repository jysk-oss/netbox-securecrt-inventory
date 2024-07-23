//go:build darwin

package assets

import _ "embed"

//go:embed osx/green-circle-icon.png
var StatusIconGreen []byte

//go:embed osx/red-circle-icon.png
var StatusIconRed []byte

//go:embed osx/icon.png
var Icon []byte

//go:embed osx/icon-1.png
var AnimateIcon1 []byte

//go:embed osx/icon-2.png
var AnimateIcon2 []byte

//go:embed osx/icon-3.png
var AnimateIcon3 []byte

//go:embed osx/icon-4.png
var AnimateIcon4 []byte

//go:embed osx/icon-5.png
var AnimateIcon5 []byte

//go:embed osx/icon-6.png
var AnimateIcon6 []byte
