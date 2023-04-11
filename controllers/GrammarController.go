package controllers

import (
	"Project/models"
	"errors"
	"fmt"
)

type GrammarController struct {
	grammar  models.Grammar
	maxDepth int16
}

func NewGrammarController() *GrammarController {
	return &GrammarController{}
}

func (gc *GrammarController) GetGrammar() models.Grammar {
	return gc.grammar
}

func (gc *GrammarController) SetGrammar(grammar models.Grammar) {
	e := errors.New("")
	// When a grammar is created, all the processes are initialized automatically
	gc.maxDepth = 1000
	gc.grammar = grammar
	gc.grammar.Firsts, e = gc.calculateFirsts()
	if e != nil {
		gc.error()
		return
	}
	gc.grammar.Follows, e = gc.calculateFollows()
	if e != nil {
		gc.error()
		return
	}
	gc.grammar.PredictionSet = gc.calculatePredictionSet()
	gc.grammar.IsLL1 = gc.isLL1(gc.grammar.PredictionSet)

}

func (gc *GrammarController) GetInitial() string {
	return gc.grammar.Initial
}

func (gc *GrammarController) GetTerminals() []string {
	return gc.grammar.Terminals
}

func (gc *GrammarController) GetNonTerminals() []string {
	return gc.grammar.NonTerminals
}

func (gc *GrammarController) GetProductions() []models.Production {
	return gc.grammar.Productions
}

func (gc *GrammarController) GetFirsts() map[string][]string {
	return gc.grammar.Firsts
}

func (gc *GrammarController) GetFollows() map[string][]string {
	return gc.grammar.Follows
}

func (gc *GrammarController) IsLL1() string {
	return gc.grammar.IsLL1
}

func (gc *GrammarController) calculateFirsts() (map[string][]string, error) {
	e := errors.New("")
	firsts := make(map[string][]string)

	// Calculate the first for each non-terminal symbol
	for _, nonTerminal := range gc.grammar.NonTerminals {
		firsts[nonTerminal], e = gc.calculateFirst(nonTerminal, 0)
	}

	return firsts, e
}

/*
This function receives a non-terminal symbol as a parameter and searches it among all the productions of the grammar
to then find the first production (The first position of production.right) and evaluate if it is terminal or not,
it uses recursion to find the first ones of a non-terminal symbol.
*/
func (gc *GrammarController) calculateFirst(nonTerminal string, depth int16) ([]string, error) {
	if depth > gc.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded")
	}
	if depth > gc.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded")
	}
	firsts := make([]string, 0)

	for _, production := range gc.grammar.Productions {
		if production.Left == nonTerminal {
			if gc.isTerminal(production.Right[0]) {
				// If begins with a terminal symbol, it's add to the firsts
				if !contains(firsts, production.Right[0]) {
					firsts = append(firsts, production.Right[0])
				}
			} else {
				// If begins with a non-terminal symbol, we must find his firsts
				symbols, err := gc.calculateFirst(production.Right[0], depth+1)
				if err != nil {
					return nil, err
				}
				for _, symbol := range symbols {
					if !contains(firsts, symbol) {
						firsts = append(firsts, symbol)
					}
				}
			}
		}
	}

	return firsts, nil
}

func (gc *GrammarController) calculateFollows() (map[string][]string, error) {
	e := errors.New("")
	follows := make(map[string][]string)
	// Calculate the follows for each non-terminal symbol
	for _, nonTerminal := range gc.grammar.NonTerminals {
		follows[nonTerminal], e = gc.calculateFollow(nonTerminal, 0)
	}
	return follows, e
}

/*
This function takes as a parameter a non-terminal symbol, which it searches on the right side of all productions, applies two rules:
- If the following is not terminal, add his firsts
- If the following is λ add his followings
the function uses recursion to find the follows of the non-terminal symbol
*/

func (gc *GrammarController) calculateFollow(nonTerminal string, depth int16) ([]string, error) {
	if depth > gc.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded")
	}
	follows := make([]string, 0)

	if nonTerminal == gc.GetInitial() {
		follows = append(follows, "$") // This find the initial production and add $ as the rule say
	}

	for _, production := range gc.grammar.Productions { //Search in all productions for the non-terminal symbol
		for i, right := range production.Right { // Search on the right side of the productions
			if right == nonTerminal { // Find the symbol that was sends by calculateFollows()
				if i+1 < len(production.Right) { // Verify that follows is not empty
					if gc.isNonTerminal(production.Right[i+1]) {
						symbols := gc.grammar.Firsts[production.Right[i+1]] // If the follows is non-terminal, we must add his firsts
						for _, symbol := range symbols {
							//We go through the first ones that return us, and we look for if we find λ in them
							if symbol != "λ" {
								if !contains(follows, symbol) { // Add symbols that don't already exist in the output
									follows = append(follows, symbol)
								}
							} else { //If we find λ in the first, we must add the follows
								aux, err := gc.calculateFollow(production.Right[i+1], depth+1)
								if err != nil {
									return nil, err
								}
								for _, symbol := range aux {
									if !contains(follows, symbol) {
										follows = append(follows, symbol)
									}
								}
							}
						}

					} else {
						// If the follows is a terminal, we add it as long as it is not already in the output
						if !contains(follows, production.Right[i+1]) {
							follows = append(follows, production.Right[i+1])
						}
					}
				} else { //If the follow is empty add the follows of production.Left
					if production.Left != nonTerminal { // Verify that the follows are not his own follows
						symbols, err := gc.calculateFollow(production.Left, depth+1)
						if err != nil {
							return nil, err
						}
						for _, symbol := range symbols {
							if !contains(follows, symbol) {
								follows = append(follows, symbol)
							}
						}
					}
				}
			}
		}
	}
	return follows, nil
}

func (gc *GrammarController) calculatePredictionSet() map[string][][]string {
	prediction := make(map[string][][]string) // In this case we have a map that contains a slice of slices

	for _, production := range gc.grammar.Productions {
		// For each production we add a slice with his prediction sets
		prediction[production.Left] = append(prediction[production.Left], gc.calculatePrediction(production))
	}

	return prediction
}

/*
In this function, we will find the prediction set of a production following 3 rules:
- If the first of the production is terminal, it is added to the productions
- If the first of the production is not terminal, we find its first
- In case of λ we find the following
We create a slice for each production to then intersect and check if a non-terminal produces the same symbols in different productions
*/

func (gc *GrammarController) calculatePrediction(production models.Production) []string {
	prediction := make([]string, 0)

	// If the non-terminal symbol has an output that begins with a terminal, that terminal is a first
	if gc.isTerminal(production.Right[0]) {
		if production.Right[0] == "λ" { // We add the follows if the firsts are λ
			aux := gc.grammar.Follows[production.Left] // Search in the maps, in this way we avoid calculating again
			for _, symbol := range aux {
				if !contains(prediction, symbol) { // To avoid repeating symbols
					prediction = append(prediction, symbol)
				}
			}
		} else {
			if !contains(prediction, production.Right[0]) { // We add the symbols that are terminal but not λ
				prediction = append(prediction, production.Right[0])
			}
		}
	} else { // If the non-terminal symbol has a production that begins with another non-terminal, the first ones of that non-terminal are added.
		symbols, _ := gc.calculateFirst(production.Right[0], 0)
		for _, symbol := range symbols {
			if symbol == "λ" {
				aux := gc.grammar.Follows[production.Right[0]]
				for _, symbol := range aux {
					if !contains(prediction, symbol) {
						prediction = append(prediction, symbol)
					}
				}
			}
			if !contains(prediction, symbol) {
				prediction = append(prediction, symbol)
			}
		}
	}
	return prediction
}

/*
In this function we will try to join the slices of each production produced by a non-terminal symbol,
if a repeated one is found, it is determined that the grammar is not LL1
*/

func (gc *GrammarController) isLL1(sets map[string][][]string) string {
	result := "It's LL1"
	for _, nonTerminal := range gc.grammar.NonTerminals {
		aux := make([]string, 0)
		for _, a := range sets[nonTerminal] {
			for _, b := range a {
				if !contains(aux, b) {
					aux = append(aux, b)
				} else {
					result = fmt.Sprintf("Is false because %v already exists in %v", b, aux)
				}
			}
		}
	}
	return result
}

func (gc *GrammarController) isTerminal(symbol string) bool {
	return !gc.isNonTerminal(symbol)
}

func (gc *GrammarController) isNonTerminal(symbol string) bool {
	for _, nonTerminal := range gc.grammar.NonTerminals {
		if nonTerminal == symbol {
			return true
		}
	}
	return false
}

// Verify if the value already exists in the slice
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func (gc *GrammarController) error() {
	gc.grammar.Firsts = nil
	gc.grammar.Follows = nil
	gc.grammar.PredictionSet = nil
	gc.grammar.IsLL1 = "Not LL1, There is a cycle in the symbols"
}
