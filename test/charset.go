package test

const (
	CharsetLowercase            = "abcdefghijklmnopqrstuvwxyz"
	CharsetUppercase            = "ABCDEFGHIJKLMNOPQRSTUVWYXZ"
	CharsetNumeric              = "1234567890"
	CharsetWhitespace           = " "
	CharsetSymbols              = "!\"#$%&'()*+,-./:;<=>@\\]^_`{|}~"
	CharsetUnicode              = "åäædðÉíïłñóóÓôöøûÜÿźþαΓεέζςирстуцההטילקกดบฝูรรวา่าいイトにニはハヘほホろロ"
	CharsetEmoji                = "😀🤣🤩😈👻😼💔💯💣👏💪🧠🧚🏄👣🐥🦖🥨🍕🍩🌎"
	CharsetAlpha                = CharsetUppercase + CharsetLowercase
	CharsetAlphaNumeric         = CharsetAlpha + CharsetNumeric
	CharsetText                 = CharsetAlphaNumeric + CharsetWhitespace + CharsetSymbols + CharsetUnicode + CharsetEmoji
	CharsetHexidecimalLowercase = CharsetNumeric + "abcdef"
)
