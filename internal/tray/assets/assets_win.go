//go:build windows

package assets

import _ "embed"

//go:embed win/green-circle-icon.ico
var StatusIconGreen []byte

//go:embed win/red-circle-icon.ico
var StatusIconRed []byte

//go:embed win/icon.ico
var Icon []byte

//go:embed win/icon-1.ico
var AnimateIcon1 []byte

//go:embed win/icon-2.ico
var AnimateIcon2 []byte

//go:embed win/icon-3.ico
var AnimateIcon3 []byte

//go:embed win/icon-4.ico
var AnimateIcon4 []byte

//go:embed win/icon-5.ico
var AnimateIcon5 []byte

//go:embed win/icon-6.ico
var AnimateIcon6 []byte
