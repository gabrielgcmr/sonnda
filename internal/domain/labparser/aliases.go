package labparser

func ResolveAnalyteCode(name string) (string, bool) {
	normalized := NormalizeForMatch(name)
	if normalized == "" {
		return "", false
	}

	code, ok := hemogramAnalyteAliases[normalized]
	return code, ok
}

var hemogramAnalyteAliases = map[string]string{
	"hemacias":    "rbc",
	"eritrocitos": "rbc",
	"rbc":         "rbc",

	"hematocrito": "hematocrit",
	"ht":          "hematocrit",
	"hct":         "hematocrit",

	"hemoglobina": "hemoglobin",
	"hb":          "hemoglobin",
	"hgb":         "hemoglobin",

	"vcm": "mcv",
	"mcv": "mcv",

	"hcm": "mch",
	"mch": "mch",

	"chcm": "mchc",
	"mchc": "mchc",

	"rdw": "rdw",

	"leucocitos": "leukocytes",
	"wbc":        "leukocytes",

	"neutrofilos": "neutrophils",
	"neutrophils": "neutrophils",

	"promielocitos": "promyelocytes",
	"promyelocytes": "promyelocytes",

	"mielocitos": "myelocytes",
	"myelocytes": "myelocytes",

	"metamielocitos": "metamyelocytes",
	"metamyelocytes": "metamyelocytes",

	"bastoes":    "bands",
	"bastonetes": "bands",
	"bands":      "bands",

	"segmentados": "segmented_neutrophils",

	"eosinofilos": "eosinophils",
	"eosinophils": "eosinophils",

	"basofilos": "basophils",
	"basophils": "basophils",

	"linfocitos tipicos": "lymphocytes",
	"linfocitos":         "lymphocytes",
	"lymphocytes":        "lymphocytes",

	"linfocitos atipicos": "atypical_lymphocytes",

	"monocitos": "monocytes",
	"monocytes": "monocytes",

	"blastos": "blasts",
	"blasts":  "blasts",

	"plaquetas": "platelets",
	"plt":       "platelets",

	"vpm": "mpv",
	"mpv": "mpv",
}
