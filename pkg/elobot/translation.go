package elobot

const (
	chooseOne            = "Select one option:"
	importList           = "Import board game list from bgg"
	randomCompare        = "Random compare"
	manageItems          = "Manage items"
	top20                = "Top 20"
	settings             = "Settings"
	cancel               = "Cancel"
	twoStepCompare       = "Two step compare"
	setLanguage          = "Set language"
	another              = "Next"
	selectCategory       = "Select Active Category"
	yesAction            = "Yes"
	noAction             = "No"
	yourUserName         = "Your BGG username:"
	yourTopTenList       = "Your top ten list (Category: %s):\n"
	chooseOption         = "This is your %q \nChoose one option or enter a number between 0-100:"
	invalidInput         = "Invalid input"
	importFirst          = "No category, import board games first"
	nothingWasChanged    = "Nothing was changed"
	selectActiveCategory = "Select the category, current active is: "
	configSaved          = "Config saved"
	wishList             = "Wishlist"
	own                  = "Own"
	played               = "Played"
	rated                = "Rated"
	unknown              = "Unknown"
	deleteItem           = "Delete %q"
	compareString        = "%s is %d%% winner"
	equal                = "Equal"
	itemsInYourList      = "%d items was in your %q list, %d was new"
	areYouSure           = "Are you sure? this can't be undone"
)

func translateFa(in string) string {
	switch in {
	case chooseOne:
		return "یکی را انتخاب کنید"
	case importList:
		return "وارد کردن لیست بوردگیمها از bgg"
	case randomCompare:
		return "مقایسه اتفاقی"
	case manageItems:
		return "مدیریت آیتمها"
	case top20:
		return "بیست‌تای برتر"
	case settings:
		return "تنظیمات"
	case cancel:
		return "لغو"
	case twoStepCompare:
		return "مقایسه دو مرحله‌ای"
	case setLanguage:
		return "انتخاب زبان"
	case another:
		return "بعدی"
	case selectCategory:
		return "انتخاب دسته فعال"
	case yesAction:
		return "بله"
	case noAction:
		return "خیر"
	case yourUserName:
		return "نام کاربری بوردگیم گیک شما: "
	case yourTopTenList:
		return "برترینهای شما در دسته %s :\n"
	case chooseOption:
		return "این لیست %q شماست\nیک گزینه انتحاب کنید یا یک عدد بین صفر تا صد وارد کنید"
	case invalidInput:
		return "ورودی نامعتبر"
	case importFirst:
		return "دسته بندی وجود ندارد، ابتدا لیست را از بوردگیم گیک وارد کنید"
	case nothingWasChanged:
		return "چیزی تغییر نکرد"
	case selectActiveCategory:
		return "دسته بندی را انتخاب کنید، دسته بندی فعلی : "
	case unknown:
		return "نامعلوم"
	case configSaved:
		return "تنظیمات ذخیره شد"
	case wishList:
		return "ویش لیست"
	case rated:
		return "درجه داده شده"
	case played:
		return "بازی شده"
	case own:
		return "بازیهای من"
	case deleteItem:
		return "حذف %q"
	case equal:
		return "برابر"
	case compareString:
		return "%s برنده %d%% است"
	case itemsInYourList:
		return "%d آیتم در لیست %q بود، %d آن جدید بود"
	case areYouSure:
		return "آیا مطمئنید؟ این کار قابل برگشت نیست"
	default:
		return in
	}
}

func t(in, lang string) string {
	switch lang {
	case "Fa":
		return translateFa(in)
	case "En":
		fallthrough
	default:
		return in
	}
}
