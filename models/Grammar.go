package models

type Grammar struct {
	Initial       string
	Terminals     []string
	NonTerminals  []string
	Productions   []Production
	Firsts        map[string][]string
	Follows       map[string][]string
	PredictionSet map[string][][]string
	IsLL1         string
}
