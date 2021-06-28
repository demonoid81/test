package strcase

var uppercaseAcronym = map[string]string{
	"ID": "id",
	"UUID": "uuid",
}

// ConfigureAcronym allows you to add additional words which will be considered acronyms
func ConfigureAcronym(key, val string) {
	uppercaseAcronym[key] = val
}