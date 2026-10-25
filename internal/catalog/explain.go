package catalog

type Explanation struct{ What, UseCase, Analogy string }

var explanations = map[string]Explanation{}

func registerExplanations(items map[string]Explanation) {
	for name, item := range items {
		explanations[name] = item
	}
}

func Explain(name string) (Explanation, bool) {
	item, ok := explanations[name]
	return item, ok
}
