package api

type TranslateApi interface {
	// Translate translates the given text from the source language to the target language.
	// It returns the translated text as a string, or an error if the translation fails.
	Translate(text []string, sourceLang string, targetLang string) ([]string, error)
	Close() error
}
