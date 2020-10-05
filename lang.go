package trans

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

var iso639map = map[string]string{
	"ab": "Abkhazian",
	"aa": "Afar",
	"af": "Afrikaans",
	"ak": "Akan",
	"sq": "Albanian",
	"am": "Amharic",
	"ar": "Arabic",
	"an": "Aragonese",
	"hy": "Armenian",
	"as": "Assamese",
	"av": "Avaric",
	"ae": "Avestan",
	"ay": "Aymara",
	"az": "Azerbaijani",
	"bm": "Bambara",
	"ba": "Bashkir",
	"eu": "Basque",
	"be": "Belarusian",
	"bn": "Bengali",
	"bh": "Bihari languages",
	"bi": "Bislama",
	"bs": "Bosnian",
	"br": "Breton",
	"bg": "Bulgarian",
	"my": "Burmese",
	"ca": "Catalan",
	"ch": "Chamorro",
	"ce": "Chechen",
	"ny": "Chichewa",
	"zh": "Chinese",
	"cv": "Chuvash",
	"kw": "Cornish",
	"co": "Corsican",
	"cr": "Cree",
	"hr": "Croatian",
	"cs": "Czech",
	"da": "Danish",
	"dv": "Divehi",
	"nl": "Dutch",
	"dz": "Dzongkha",
	"en": "English",
	"eo": "Esperanto",
	"et": "Estonian",
	"ee": "Ewe",
	"fo": "Faroese",
	"fj": "Fijian",
	"fi": "Finnish",
	"fr": "French",
	"ff": "Fulah",
	"gl": "Galician",
	"ka": "Georgian",
	"de": "German",
	"el": "Greek",
	"gn": "Guarani",
	"gu": "Gujarati",
	"ht": "Haitian",
	"ha": "Hausa",
	"he": "Hebrew",
	"hz": "Herero",
	"hi": "Hindi",
	"ho": "Hiri Motu",
	"hu": "Hungarian",
	"ia": "Interlingua",
	"id": "Indonesian",
	"ie": "Interlingue",
	"ga": "Irish",
	"ig": "Igbo",
	"ik": "Inupiaq",
	"io": "Ido",
	"is": "Icelandic",
	"it": "Italian",
	"iu": "Inuktitut",
	"ja": "Japanese",
	"jv": "Javanese",
	"kl": "Kalaallisut",
	"kn": "Kannada",
	"kr": "Kanuri",
	"ks": "Kashmiri",
	"kk": "Kazakh",
	"km": "Central Khmer",
	"ki": "Kikuyu",
	"rw": "Kinyarwanda",
	"ky": "Kirghiz",
	"kv": "Komi",
	"kg": "Kongo",
	"ko": "Korean",
	"ku": "Kurdish",
	"kj": "Kuanyama",
	"la": "Latin",
	"lb": "Luxembourgish",
	"lg": "Ganda",
	"li": "Limburgan",
	"ln": "Lingala",
	"lo": "Lao",
	"lt": "Lithuanian",
	"lu": "Luba-Katanga",
	"lv": "Latvian",
	"gv": "Manx",
	"mk": "Macedonian",
	"mg": "Malagasy",
	"ms": "Malay",
	"ml": "Malayalam",
	"mt": "Maltese",
	"mi": "Maori",
	"mr": "Marathi",
	"mh": "Marshallese",
	"mn": "Mongolian",
	"na": "Nauru",
	"nv": "Navajo",
	"nd": "North Ndebele",
	"ne": "Nepali",
	"ng": "Ndonga",
	"nb": "Norwegian Bokmål",
	"nn": "Norwegian Nynorsk",
	"no": "Norwegian",
	"ii": "Sichuan Yi",
	"nr": "South Ndebele",
	"oc": "Occitan",
	"oj": "Ojibwa",
	"cu": "Church Slavic",
	"om": "Oromo",
	"or": "Oriya",
	"os": "Ossetian",
	"pa": "Punjabi",
	"pi": "Pali",
	"fa": "Persian",
	"pl": "Polish",
	"ps": "Pashto",
	"pt": "Portuguese",
	"qu": "Quechua",
	"rm": "Romansh",
	"rn": "Rundi",
	"ro": "Romanian",
	"ru": "Russian",
	"sa": "Sanskrit",
	"sc": "Sardinian",
	"sd": "Sindhi",
	"se": "Northern Sami",
	"sm": "Samoan",
	"sg": "Sango",
	"sr": "Serbian",
	"gd": "Gaelic",
	"sn": "Shona",
	"si": "Sinhala",
	"sk": "Slovak",
	"sl": "Slovenian",
	"so": "Somali",
	"st": "Southern Sotho",
	"es": "Spanish",
	"su": "Sundanese",
	"sw": "Swahili",
	"ss": "Swati",
	"sv": "Swedish",
	"ta": "Tamil",
	"te": "Telugu",
	"tg": "Tajik",
	"th": "Thai",
	"ti": "Tigrinya",
	"bo": "Tibetan",
	"tk": "Turkmen",
	"tl": "Tagalog",
	"tn": "Tswana",
	"to": "Tonga",
	"tr": "Turkish",
	"ts": "Tsonga",
	"tt": "Tatar",
	"tw": "Twi",
	"ty": "Tahitian",
	"ug": "Uighur",
	"uk": "Ukrainian",
	"ur": "Urdu",
	"uz": "Uzbek",
	"ve": "Venda",
	"vi": "Vietnamese",
	"vo": "Volapük",
	"wa": "Walloon",
	"cy": "Welsh",
	"wo": "Wolof",
	"fy": "Western Frisian",
	"xh": "Xhosa",
	"yi": "Yiddish",
	"yo": "Yoruba",
	"za": "Zhuang",
	"zu": "Zulu",
}

func LookupLang(s string) (code, name string, err error) {
	c := strings.ToLower(s)
	n, ok := iso639map[c]
	if ok {
		return c, n, nil
	}
	for k, v := range iso639map {
		lv := strings.ToLower(v)
		if strings.Contains(lv, c) {
			return k, v, nil
		}
	}
	return "", "", fmt.Errorf("%s is not found", s)
}

type ISO639 struct {
	Code string
	Name string
}

func LangList() []ISO639 {
	a := make([]ISO639, len(iso639map))
	i := 0
	for k, v := range iso639map {
		a[i] = ISO639{k, v}
		i++
	}
	sort.Slice(a,func(i, j int) bool {
		return a[i].Code < a[j].Code
	})
	return a
}

func CurrentLang() (code, name string) {
	var lang string
	if s, ok := os.LookupEnv("LANG"); ok {
		lang = string(s[:2])
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "Get-Culture | Select-Object -exp Name")
		if bs, err := cmd.Output(); err == nil {
			lang = string(bs[:2])
		}
	}
	if len(lang) == 2 {
		code, name, err := LookupLang(lang)
		if err == nil {
			return code, name
		}
	}
	return "en", "English"
}
