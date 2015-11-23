package material

type Color uint32

// RGBA returns the unit value of each component.
func (c Color) RGBA() (r, g, b, a float32) {
	ur := uint8(c >> 24)
	ug := uint8(c >> 16)
	ub := uint8(c >> 8)
	ua := uint8(c)
	r, g, b, a = float32(ur)/255, float32(ug)/255, float32(ub)/255, float32(ua)/255
	return
}

const (
	RedPrimary Color = Red500
	Red50      Color = 0xFFEBEEFF
	Red100     Color = 0xFFCDD2FF
	Red200     Color = 0xEF9A9AFF
	Red300     Color = 0xE57373FF
	Red400     Color = 0xEF5350FF
	Red500     Color = 0xF44336FF
	Red600     Color = 0xE53935FF
	Red700     Color = 0xD32F2FFF
	Red800     Color = 0xC62828FF
	Red900     Color = 0xB71C1CFF
	RedA100    Color = 0xFF8A80FF
	RedA200    Color = 0xFF5252FF
	RedA400    Color = 0xFF1744FF
	RedA700    Color = 0xD50000FF

	PinkPrimary Color = Pink500
	Pink50      Color = 0xFCE4ECFF
	Pink100     Color = 0xF8BBD0FF
	Pink200     Color = 0xF48FB1FF
	Pink300     Color = 0xF06292FF
	Pink400     Color = 0xEC407AFF
	Pink500     Color = 0xE91E63FF
	Pink600     Color = 0xD81B60FF
	Pink700     Color = 0xC2185BFF
	Pink800     Color = 0xAD1457FF
	Pink900     Color = 0x880E4FFF
	PinkA100    Color = 0xFF80ABFF
	PinkA200    Color = 0xFF4081FF
	PinkA400    Color = 0xF50057FF
	PinkA700    Color = 0xC51162FF

	PurplePrimary Color = Purple500
	Purple50      Color = 0xF3E5F5FF
	Purple100     Color = 0xE1BEE7FF
	Purple200     Color = 0xCE93D8FF
	Purple300     Color = 0xBA68C8FF
	Purple400     Color = 0xAB47BCFF
	Purple500     Color = 0x9C27B0FF
	Purple600     Color = 0x8E24AAFF
	Purple700     Color = 0x7B1FA2FF
	Purple800     Color = 0x6A1B9AFF
	Purple900     Color = 0x4A148CFF
	PurpleA100    Color = 0xEA80FCFF
	PurpleA200    Color = 0xE040FBFF
	PurpleA400    Color = 0xD500F9FF
	PurpleA700    Color = 0xAA00FFFF

	DeepPurplePrimary Color = DeepPurple500
	DeepPurple50      Color = 0xEDE7F6FF
	DeepPurple100     Color = 0xD1C4E9FF
	DeepPurple200     Color = 0xB39DDBFF
	DeepPurple300     Color = 0x9575CDFF
	DeepPurple400     Color = 0x7E57C2FF
	DeepPurple500     Color = 0x673AB7FF
	DeepPurple600     Color = 0x5E35B1FF
	DeepPurple700     Color = 0x512DA8FF
	DeepPurple800     Color = 0x4527A0FF
	DeepPurple900     Color = 0x311B92FF
	DeepPurpleA100    Color = 0xB388FFFF
	DeepPurpleA200    Color = 0x7C4DFFFF
	DeepPurpleA400    Color = 0x651FFFFF
	DeepPurpleA700    Color = 0x6200EAFF

	IndigoPrimary Color = Indigo500
	Indigo50      Color = 0xE8EAF6FF
	Indigo100     Color = 0xC5CAE9FF
	Indigo200     Color = 0x9FA8DAFF
	Indigo300     Color = 0x7986CBFF
	Indigo400     Color = 0x5C6BC0FF
	Indigo500     Color = 0x3F51B5FF
	Indigo600     Color = 0x3949ABFF
	Indigo700     Color = 0x303F9FFF
	Indigo800     Color = 0x283593FF
	Indigo900     Color = 0x1A237EFF
	IndigoA100    Color = 0x8C9EFFFF
	IndigoA200    Color = 0x536DFEFF
	IndigoA400    Color = 0x3D5AFEFF
	IndigoA700    Color = 0x304FFEFF

	BluePrimary Color = Blue500
	Blue50      Color = 0xE3F2FDFF
	Blue100     Color = 0xBBDEFBFF
	Blue200     Color = 0x90CAF9FF
	Blue300     Color = 0x64B5F6FF
	Blue400     Color = 0x42A5F5FF
	Blue500     Color = 0x2196F3FF
	Blue600     Color = 0x1E88E5FF
	Blue700     Color = 0x1976D2FF
	Blue800     Color = 0x1565C0FF
	Blue900     Color = 0x0D47A1FF
	BlueA100    Color = 0x82B1FFFF
	BlueA200    Color = 0x448AFFFF
	BlueA400    Color = 0x2979FFFF
	BlueA700    Color = 0x2962FFFF

	LightBluePrimary Color = LightBlue500
	LightBlue50      Color = 0xE1F5FEFF
	LightBlue100     Color = 0xB3E5FCFF
	LightBlue200     Color = 0x81D4FAFF
	LightBlue300     Color = 0x4FC3F7FF
	LightBlue400     Color = 0x29B6F6FF
	LightBlue500     Color = 0x03A9F4FF
	LightBlue600     Color = 0x039BE5FF
	LightBlue700     Color = 0x0288D1FF
	LightBlue800     Color = 0x0277BDFF
	LightBlue900     Color = 0x01579BFF
	LightBlueA100    Color = 0x80D8FFFF
	LightBlueA200    Color = 0x40C4FFFF
	LightBlueA400    Color = 0x00B0FFFF
	LightBlueA700    Color = 0x0091EAFF

	CyanPrimary Color = Cyan500
	Cyan50      Color = 0xE0F7FAFF
	Cyan100     Color = 0xB2EBF2FF
	Cyan200     Color = 0x80DEEAFF
	Cyan300     Color = 0x4DD0E1FF
	Cyan400     Color = 0x26C6DAFF
	Cyan500     Color = 0x00BCD4FF
	Cyan600     Color = 0x00ACC1FF
	Cyan700     Color = 0x0097A7FF
	Cyan800     Color = 0x00838FFF
	Cyan900     Color = 0x006064FF
	CyanA100    Color = 0x84FFFFFF
	CyanA200    Color = 0x18FFFFFF
	CyanA400    Color = 0x00E5FFFF
	CyanA700    Color = 0x00B8D4FF

	TealPrimary Color = Teal500
	Teal50      Color = 0xE0F2F1FF
	Teal100     Color = 0xB2DFDBFF
	Teal200     Color = 0x80CBC4FF
	Teal300     Color = 0x4DB6ACFF
	Teal400     Color = 0x26A69AFF
	Teal500     Color = 0x009688FF
	Teal600     Color = 0x00897BFF
	Teal700     Color = 0x00796BFF
	Teal800     Color = 0x00695CFF
	Teal900     Color = 0x004D40FF
	TealA100    Color = 0xA7FFEBFF
	TealA200    Color = 0x64FFDAFF
	TealA400    Color = 0x1DE9B6FF
	TealA700    Color = 0x00BFA5FF

	GreenPrimary Color = Green500
	Green50      Color = 0xE8F5E9FF
	Green100     Color = 0xC8E6C9FF
	Green200     Color = 0xA5D6A7FF
	Green300     Color = 0x81C784FF
	Green400     Color = 0x66BB6AFF
	Green500     Color = 0x4CAF50FF
	Green600     Color = 0x43A047FF
	Green700     Color = 0x388E3CFF
	Green800     Color = 0x2E7D32FF
	Green900     Color = 0x1B5E20FF
	GreenA100    Color = 0xB9F6CAFF
	GreenA200    Color = 0x69F0AEFF
	GreenA400    Color = 0x00E676FF
	GreenA700    Color = 0x00C853FF

	LightGreenPrimary Color = LightGreen500
	LightGreen50      Color = 0xF1F8E9FF
	LightGreen100     Color = 0xDCEDC8FF
	LightGreen200     Color = 0xC5E1A5FF
	LightGreen300     Color = 0xAED581FF
	LightGreen400     Color = 0x9CCC65FF
	LightGreen500     Color = 0x8BC34AFF
	LightGreen600     Color = 0x7CB342FF
	LightGreen700     Color = 0x689F38FF
	LightGreen800     Color = 0x558B2FFF
	LightGreen900     Color = 0x33691EFF
	LightGreenA100    Color = 0xCCFF90FF
	LightGreenA200    Color = 0xB2FF59FF
	LightGreenA400    Color = 0x76FF03FF
	LightGreenA700    Color = 0x64DD17FF

	LimePrimary Color = Lime500
	Lime50      Color = 0xF9FBE7FF
	Lime100     Color = 0xF0F4C3FF
	Lime200     Color = 0xE6EE9CFF
	Lime300     Color = 0xDCE775FF
	Lime400     Color = 0xD4E157FF
	Lime500     Color = 0xCDDC39FF
	Lime600     Color = 0xC0CA33FF
	Lime700     Color = 0xAFB42BFF
	Lime800     Color = 0x9E9D24FF
	Lime900     Color = 0x827717FF
	LimeA100    Color = 0xF4FF81FF
	LimeA200    Color = 0xEEFF41FF
	LimeA400    Color = 0xC6FF00FF
	LimeA700    Color = 0xAEEA00FF

	YellowPrimary Color = Yellow500
	Yellow50      Color = 0xFFFDE7FF
	Yellow100     Color = 0xFFF9C4FF
	Yellow200     Color = 0xFFF59DFF
	Yellow300     Color = 0xFFF176FF
	Yellow400     Color = 0xFFEE58FF
	Yellow500     Color = 0xFFEB3BFF
	Yellow600     Color = 0xFDD835FF
	Yellow700     Color = 0xFBC02DFF
	Yellow800     Color = 0xF9A825FF
	Yellow900     Color = 0xF57F17FF
	YellowA100    Color = 0xFFFF8DFF
	YellowA200    Color = 0xFFFF00FF
	YellowA400    Color = 0xFFEA00FF
	YellowA700    Color = 0xFFD600FF

	AmberPrimary Color = Amber500
	Amber50      Color = 0xFFF8E1FF
	Amber100     Color = 0xFFECB3FF
	Amber200     Color = 0xFFE082FF
	Amber300     Color = 0xFFD54FFF
	Amber400     Color = 0xFFCA28FF
	Amber500     Color = 0xFFC107FF
	Amber600     Color = 0xFFB300FF
	Amber700     Color = 0xFFA000FF
	Amber800     Color = 0xFF8F00FF
	Amber900     Color = 0xFF6F00FF
	AmberA100    Color = 0xFFE57FFF
	AmberA200    Color = 0xFFD740FF
	AmberA400    Color = 0xFFC400FF
	AmberA700    Color = 0xFFAB00FF

	OrangePrimary Color = Orange500
	Orange50      Color = 0xFFF3E0FF
	Orange100     Color = 0xFFE0B2FF
	Orange200     Color = 0xFFCC80FF
	Orange300     Color = 0xFFB74DFF
	Orange400     Color = 0xFFA726FF
	Orange500     Color = 0xFF9800FF
	Orange600     Color = 0xFB8C00FF
	Orange700     Color = 0xF57C00FF
	Orange800     Color = 0xEF6C00FF
	Orange900     Color = 0xE65100FF
	OrangeA100    Color = 0xFFD180FF
	OrangeA200    Color = 0xFFAB40FF
	OrangeA400    Color = 0xFF9100FF
	OrangeA700    Color = 0xFF6D00FF

	DeepOrangePrimary Color = DeepOrange500
	DeepOrange50      Color = 0xFBE9E7FF
	DeepOrange100     Color = 0xFFCCBCFF
	DeepOrange200     Color = 0xFFAB91FF
	DeepOrange300     Color = 0xFF8A65FF
	DeepOrange400     Color = 0xFF7043FF
	DeepOrange500     Color = 0xFF5722FF
	DeepOrange600     Color = 0xF4511EFF
	DeepOrange700     Color = 0xE64A19FF
	DeepOrange800     Color = 0xD84315FF
	DeepOrange900     Color = 0xBF360CFF
	DeepOrangeA100    Color = 0xFF9E80FF
	DeepOrangeA200    Color = 0xFF6E40FF
	DeepOrangeA400    Color = 0xFF3D00FF
	DeepOrangeA700    Color = 0xDD2C00FF

	BrownPrimary Color = Brown500
	Brown50      Color = 0xEFEBE9FF
	Brown100     Color = 0xD7CCC8FF
	Brown200     Color = 0xBCAAA4FF
	Brown300     Color = 0xA1887FFF
	Brown400     Color = 0x8D6E63FF
	Brown500     Color = 0x795548FF
	Brown600     Color = 0x6D4C41FF
	Brown700     Color = 0x5D4037FF
	Brown800     Color = 0x4E342EFF
	Brown900     Color = 0x3E2723FF

	GreyPrimary Color = Grey500
	Grey50      Color = 0xFAFAFAFF
	Grey100     Color = 0xF5F5F5FF
	Grey200     Color = 0xEEEEEEFF
	Grey300     Color = 0xE0E0E0FF
	Grey400     Color = 0xBDBDBDFF
	Grey500     Color = 0x9E9E9EFF
	Grey600     Color = 0x757575FF
	Grey700     Color = 0x616161FF
	Grey800     Color = 0x424242FF
	Grey900     Color = 0x212121FF

	BlueGreyPrimary Color = BlueGrey500
	BlueGrey50      Color = 0xECEFF1FF
	BlueGrey100     Color = 0xCFD8DCFF
	BlueGrey200     Color = 0xB0BEC5FF
	BlueGrey300     Color = 0x90A4AEFF
	BlueGrey400     Color = 0x78909CFF
	BlueGrey500     Color = 0x607D8BFF
	BlueGrey600     Color = 0x546E7AFF
	BlueGrey700     Color = 0x455A64FF
	BlueGrey800     Color = 0x37474FFF
	BlueGrey900     Color = 0x263238FF

	Black Color = 0x000000FF
	White Color = 0xFFFFFFFF
)

type Palette struct {
	Primary, Dark, Light Color
	Accent               Color
}
