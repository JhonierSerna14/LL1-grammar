package controllers

import (
	"Project/models"
	"errors"
	"fmt"
	"reflect"
)

type GrammarController struct {
	grammar    models.Grammar
	production models.Production
	maxDepth   int16
}

func NewGrammarController() *GrammarController {
	return &GrammarController{}
}

func (gc *GrammarController) GetGrammar() models.Grammar {
	return gc.grammar
}

func (gc *GrammarController) SetGrammar(grammar models.Grammar) {
	// When a grammar is created, all the processes are initialized automatically
	gc.maxDepth = 1000
	gc.grammar = grammar
	gc.removeFactorization()
	gc.deleteRecursion()
	e := errors.New("")
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

func (gc *GrammarController) removeFactorization() {
	productions := gc.findProductionsWithFactorization()
	for _, p := range productions {
		if len(p) > 1 {
			maxLen := len(p[0].Right)     // We assign any amount as a guess
			for i := 1; i < len(p); i++ { // We go through all the productions that have factorization
				l := compareProductions(p[0], p[i]) // We look for the maximum number of characters that match from left to right
				if l < maxLen {
					maxLen = l // We assign the minimum number of matches
				}
			}
			var name = gc.addQuoteToVariable(p[0].Left) // We add ' in the end of the original variable, and it's the new production that has the productions not repeated
			// We create an auxiliary production in which we will store the variables that are repeated and at the end
			// we will add a reference to a new production that will have what is not repeated
			var aux models.Production
			aux.Left = p[0].Left
			for _, q := range p[0].Right[0:maxLen] {
				aux.Right = append(aux.Right, q)
			}
			aux.Right = append(aux.Right, name)
			gc.grammar.Productions = append(gc.grammar.Productions, aux) // We add the new modified production to the list

			// The new production will have everything that is not repeated in the with factorization productions
			var aux2 models.Production
			aux2.Left = name
			gc.grammar.NonTerminals = append(gc.grammar.NonTerminals, aux2.Left)
			for _, in := range p {
				for _, in2 := range in.Right[maxLen:] { // We send the slice in such a way that the common symbols are not added again
					aux2.Right = append(aux2.Right, in2)
				}
				if aux2.Right == nil {
					// The case in which removing the repeated symbols leaves the slice empty, so we must allow a λ in the new production
					aux2.Right = append(aux2.Right, "λ")
				}
				gc.grammar.Productions = append(gc.grammar.Productions, aux2)
				aux2.Right = nil

				gc.removeProductionByValue(in) // Delete the original productions
			}
		}
	}
}

func (gc *GrammarController) findProductionsWithFactorization() [][]models.Production {
	var with = gc.symbolsWithFactorization()
	var aux = make(map[string][]models.Production)
	for _, nonTerminal := range with {
		for _, production := range gc.GetProductions() {
			if nonTerminal == production.Left {
				firstChar := production.Right[0]
				if _, ok := aux[firstChar]; !ok {
					aux[firstChar] = []models.Production{production}
				} else {
					aux[firstChar] = append(aux[firstChar], production)
				}
			}
		}
	}

	var result = make([][]models.Production, 0)
	for _, productions := range aux {
		if len(productions) > 1 {
			result = append(result, productions)
		}
	}
	return result
}

func (gc *GrammarController) symbolsWithFactorization() []string {
	var withFactorization = make([]string, 0)
	for _, nonTerminal := range gc.GetNonTerminals() {
		var aux = make([]string, 0)
		for _, production := range gc.GetProductions() {
			if nonTerminal == production.Left {
				if !contains(aux, production.Right[0]) {
					aux = append(aux, production.Right[0])
				} else {
					if !contains(withFactorization, production.Left) {
						withFactorization = append(withFactorization, production.Left)
					}
				}
			}
		}
	}
	return withFactorization
}

func (gc *GrammarController) deleteRecursion() {
	var withRecursion = hasRecursion(gc.grammar) // We will do the procedure only for productions with recursion
	for _, with := range withRecursion {
		var name = ""
		for i, production := range gc.grammar.Productions {
			if production.Left == with { // We search the left symbol in all productions
				if production.Left == production.Right[0] { // We find in the productions the production with recursion
					var aux = production.Right // We make a copy to be able to empty the original production
					production.Right = nil
					name = gc.addQuoteToVariable(production.Left) // We add ' to the symbol to differentiate it
					production.Left = name
					for _, x := range aux[1:] { // We change the order of production, avoiding the first symbol
						production.Right = append(production.Right, x)
					}
					production.Right = append(production.Right, production.Left) // We add the symbol that initially produced the recursion
					gc.grammar.Productions[i] = production                       // We add the modified production in the original position of the slice

				} else {
					// If the production does not generate recursion, the symbol with ' is added at the end
					if gc.grammar.Productions[i].Right[0] != "λ" {
						if name == "" {
							name = gc.addQuoteToVariable(production.Left) // We add ' to the symbol to differentiate it
						}
						gc.grammar.Productions[i].Right = append(gc.grammar.Productions[i].Right, name)
					}
				}
			}
		}
		gc.grammar.NonTerminals = append(gc.grammar.NonTerminals, name) // We add the new production to the non-Terminals
		var aux = gc.production
		aux.Left = name
		aux.Right = append(aux.Right, "λ")
		gc.grammar.Productions = append(gc.grammar.Productions, aux) // We add a new production with λ as the rule indicates
	}
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
calculateFirst: This function receives a non-terminal symbol as a parameter and searches it among all the productions of the grammar
to then find the first production (The first position of production.right) and evaluate if it is terminal or not,
it uses recursion to find the first ones of a non-terminal symbol.
*/
func (gc *GrammarController) calculateFirst(nonTerminal string, depth int16) ([]string, error) {
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
calculateFollow: This function takes as a parameter a non-terminal symbol, which it searches on the right side of all productions, applies two rules:
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
calculatePrediction: In this function, we will find the prediction set of a production following 3 rules:
- If the first of the production is terminal, it is added to the productions
- If the first of the production is not terminal, we find its first
- In case of λ we find the following
We create a slice for each production to then intersect and check if a non-terminal produces the same symbols in different productions
*/

func (gc *GrammarController) calculatePrediction(production models.Production) []string {
	prediction := make([]string, 0)

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
				aux := gc.grammar.Follows[production.Left]
				for _, sym := range aux {
					if !contains(prediction, sym) {
						prediction = append(prediction, sym)
					}
				}
			} else {
				if !contains(prediction, symbol) {
					prediction = append(prediction, symbol)
				}
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

/*
In this function we will search for all the symbols that generate recursion by left,
we will search if the left part of the production is equal to the first symbol of its right part, thus generating a direct self-call.
We return a slice with all the symbols that have recursion in any of their productions
*/
func hasRecursion(grammar models.Grammar) []string {
	symbols := make([]string, 0)
	for _, nonTerminal := range grammar.NonTerminals {
		for _, production := range grammar.Productions {
			if production.Left == nonTerminal {
				if production.Left == production.Right[0] {
					if !contains(symbols, production.Left) {
						symbols = append(symbols, production.Left)
					}
				}
			}
		}
	}
	return symbols
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

func compareProductions(p1, p2 models.Production) int {
	len1 := len(p1.Right)
	len2 := len(p2.Right)

	minLen := len1
	if len2 < len1 {
		minLen = len2
	}

	for i := 0; i < minLen; i++ {
		if p1.Right[i] != p2.Right[i] {
			return i
		}
	}

	return minLen
}

// The function receives a variable name to which apostrophes will be added at the end,
// if the variable with an apostrophe already exists, another one will be added until the variable does not exist
func (gc *GrammarController) addQuoteToVariable(name string) string {
	newName := name + "'"
	for gc.variableExists(newName) {
		newName += "'"
	}
	return newName
}

func (gc *GrammarController) variableExists(name string) bool {
	for _, production := range gc.grammar.Productions {
		if production.Left == name {
			return true
		}
	}
	return false
}

func (gc *GrammarController) removeProductionByValue(prod models.Production) {
	for i, p := range gc.grammar.Productions {
		if reflect.DeepEqual(p, prod) {
			gc.removeProduction(i)
			break
		}
	}
}

func (gc *GrammarController) removeProduction(index int) {
	gc.grammar.Productions = append(gc.grammar.Productions[:index], gc.grammar.Productions[index+1:]...)
}
