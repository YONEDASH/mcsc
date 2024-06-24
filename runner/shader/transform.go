package shader

func InitTransformers() map[string]Transformer {
	return map[string]Transformer{
		"":          transformNone{},
		"halfUpper": transformHalfUpper{},
		"halfLower": transformHalfLower{},
		"halfUpperLower": transformHalfUpperLower{
			upper: transformHalfUpper{},
			lower: transformHalfLower{},
		},
	}
}

type transformNone struct{}

func (transformNone) Generate(entries []string) []string {
	return entries
}

type transformHalfUpper struct{}

func (transformHalfUpper) Generate(entries []string) []string {
	var result []string
	for _, entry := range entries {
		result = append(result, entry+":half=upper")
	}
	return result
}

type transformHalfLower struct{}

func (transformHalfLower) Generate(entries []string) []string {
	var result []string
	for _, entry := range entries {
		result = append(result, entry+":half=lower")
	}
	return result
}

type transformHalfUpperLower struct {
	upper transformHalfUpper
	lower transformHalfLower
}

func (t transformHalfUpperLower) Generate(entries []string) []string {
	return append(t.lower.Generate(entries), t.upper.Generate(entries)...)
}
