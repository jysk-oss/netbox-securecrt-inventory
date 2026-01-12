//go:build linux

package assets

import _ "embed"

//go:embed linux/green-circle-icon.png
var StatusIconGreen []byte

//go:embed linux/red-circle-icon.png
var StatusIconRed []byte

//go:embed linux/icon.png
var Icon []byte

//go:embed linux/icon-1.png
var AnimateIcon1 []byte

//go:embed linux/icon-2.png
var AnimateIcon2 []byte

//go:embed linux/icon-3.png
var AnimateIcon3 []byte

//go:embed linux/icon-4.png
var AnimateIcon4 []byte

//go:embed linux/icon-5.png
var AnimateIcon5 []byte

//go:embed linux/icon-6.png
var AnimateIcon6 []byte
