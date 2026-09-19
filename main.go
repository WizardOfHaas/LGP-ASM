package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

/*
	Syntax is:
	[lable:] [order [arg]] [#comment]

	Arg can be numeric OR symbolic
	If it's symbolic then it has to be replaces on the 2nd pass

	I'd like to implement pre-processor for:
	%macro
	%origin
	%constant

	Literals need to be supported as:
	decimal (no prefix)
	0lLGPHEX (0, lower case L, LGP hex number)
	0LLGPHEX (0, upper case L, LGP hex number)
	0xSTDHEX (0, lower case H, standard hex number)
	0XSTDHEX (0, upper case H, standard hex number)
	0bBINARY (0, lower case B, binary number)
	0BBINARY (0, upper case B, binary number)
*/

type Exp struct {
	//Filled during first pass
	Label          string
	Order          Order
	Argument       []string //Raw, to be converted
	PackedArgument []int    //Converted
	Size           int
	Address        int
}

type Order struct {
	Identifier string
	Value      string
}

var orders = []Order{
	{"brg", "b"},
	{"has", "h"},
	{"cas", "c"},
	{"sta", "s"},
	{"tra", "t"},
	{"ret", "r"},
	{"tst", "t"},
	{"stp", "s"},
	{"prt", "p"},
	{"inp", "i"},
	{"add", "a"},
	{"sub", "s"},
	{"mup", "m"},
	{"mlo", "m"},
	{"div", "d"},
	{"ext", "e"},

	//Pseudo-orders for data definitions
	{"word", ""},
	{"char", ""},
}

var stdHexToLgpMap = map[string]string{"a": "f", "b": "g", "c": "j", "d": "k", "e": "q", "f": "w"}
var lgpHexToStdMap = map[string]string{"0": "0", "1": "1", "2": "2", "3": "3", "4": "4", "5": "5", "6": "6", "7": "7", "8": "8", "9": "9", "f": "a", "g": "b", "j": "c", "k": "d", "q": "e", "w": "f"}

var asciiToFlex = []int{
	-1, -1, -1, -1, -1, -1, -1, -1,
	024, 030, -1, -1, -1, 020, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1,
	003, -1, 016, 042, 032, 026, -1, 040,
	046, 001, 012, 013, 033, 007, 027, 023,
	002, 006, 012, 016, 022, 026, 032, 036,
	042, 046, 017, 017, 004, 013, 010, 023,
	-1, 071, 005, 065, 025, 045, 052, 056,
	061, 021, 062, 066, 006, 035, 031, 043,
	041, 072, 015, 075, 055, 051, 037, 076,
	047, 011, 001, 033, -1, 027, 022, 007,
	-1, 071, 005, 065, 025, 045, 052, 056,
	061, 021, 062, 066, 006, 035, 031, 043,
	041, 072, 015, 075, 055, 051, 037, 076,
	047, 011, 001, -1, 014, -1, 036, 077,
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isBinary(s string) bool{
	m, err := regexp.MatchString("0[bB][01]*", s)
	return err == nil && m
}

func isStdHex(s string) bool{
	m, err := regexp.MatchString("0[xX][0-9a-fA-F]*", s)
	return err == nil && m
}

func isLgpHex(s string) bool{
	m, err := regexp.MatchString("0[lL][0-9fgjkqwFGJKQW]*", s)
	return err == nil && m
}

func isSymbolic(s string) bool {
	quoted := strings.HasPrefix(s, "'") || strings.HasPrefix(s, "\"")
	literal := strings.HasPrefix(s, "0")
	return !quoted && !isNumeric(s) && !literal
}

func getLabel(t []string) *string {
	if strings.HasSuffix(t[0], ":") {
		label := strings.TrimSuffix(t[0], ":")
		return &label
	}

	return nil
}

func getOrderAndArg(t []string) (*Order, *[]string, error) {
	var arg []string

	if len(t) < 1 {
		return nil, nil, errors.New("No tokens")
	}

	for _, o := range orders {
		if t[0] == o.Identifier {
			//Argument will be raw text initially.
			//2nd pass will transform to raw number or symbol ref
			if len(t) > 1 {
				arg = t[1:]
			}

			return &o, &arg, nil
		}
	}

	return nil, nil, errors.New("Invalid Order")
}

func packLiteral(s string) (int, error) {
	/*
		Convert a literal to a numeric val
		0x12EF	Standard hex (0, lower case X, standard hex number)
		0lFGJ5	LGP Hex (0, lower case L, LGP hex number)
		0b0110	Binary (0, lower case B, binary number)
		0x12EF	Standard hex (0, upper case X, standard hex number)
		0lFGJ5	LGP Hex (0, upper case L, LGP hex number)
		0b0110	Binary (0, upper case B, binary number)
		1234	Decimal (no prefix)
	*/

	if isNumeric(s) {
		i, err := strconv.Atoi(s)
		return i, err
	}

	if isBinary(s){
		i, err := strconv.ParseInt(s[2:], 2, 64)
		return int(i), err
	}

	if isStdHex(s){
		i, err := strconv.ParseInt(s[2:], 16, 64)
		return int(i), err
	}

	if isLgpHex(s){
		stdHex := ""

		for _, c := range strings.ToLower(s[2:]){
			lgpHexVal, ok := lgpHexToStdMap[string(c)]

			if !ok{
				return -1, errors.New("Invalid LGP Hex character")
			}

			stdHex += lgpHexVal
		}

		i, err := strconv.ParseInt(stdHex, 16, 64)
		return int(i), err		
	}

	return -1, errors.New("That doesn't look like a literal")
}

func packArg(exp Exp) ([]int, error) {
	var packedArgument []int

	if exp.Order.Identifier == "word" {
		exp.Size = len(exp.Argument)

		//Just cast it
		for _, a := range exp.Argument {
			//aInt, err := strconv.Atoi(a)

			val, err := packLiteral(a)

			if err == nil {
				packedArgument = append(packedArgument, val)
			} else {
				//Do a placeholder for possible symbols
				packedArgument = append(packedArgument, -1)
			}
		}
	} else if exp.Order.Identifier == "char" {
		//For chars I have to do 6-bit packing
		r := regexp.MustCompile("['\"](.*)['\"]")

		var rawChars []rune

		//Test that we have quotes
		for _, a := range exp.Argument {
			match := r.FindStringSubmatch(a)

			if len(match) < 2 {
				return nil, errors.New("Invalid character definition")
			}

			//Drop quotes, break everything into single chars
			rawChars = append(rawChars, []rune(match[1])...)
		}

		//...I'm gonna need a char map here
		for _, r := range rawChars {
			//Right now I'm doing this raw. I will need to do the 6-bit packing later
			packedArgument = append(packedArgument, asciiToFlex[r])
		}
	} else if len(exp.Argument) == 1 {
		//Otherwise, this is a normal order with literal arg
		//	Leave symbolic args as unpacked, they get handled in phase 2
		//i, err := strconv.Atoi(exp.Argument[0])
		val, err := packLiteral(exp.Argument[0])

		if err != nil {
			//Placeholder for symbol replacement
			packedArgument = []int{-1}
		} else {
			packedArgument = []int{val}
		}
	}

	return packedArgument, nil
}

func parseLine(l string) (*Exp, error) {
	exp := Exp{}

	t := strings.Split(l, " ")

	if len(t) == 0 {
		return nil, errors.New("Empty line")
	}

	label := getLabel(t)

	//If we have a label then we can drop it from the line
	if label != nil {
		exp.Label = *label
		_, t = t[0], t[1:]
	}

	if len(t) > 0 {
		order, arg, err := getOrderAndArg(t)

		if err != nil {
			return nil, errors.New("Invalid Order")
		}

		exp.Order = *order
		exp.Argument = *arg
		exp.Size = 1 //Set initial exp size

		packedArgument, err := packArg(exp)

		if err != nil {
			return nil, err
		}

		exp.PackedArgument = packedArgument

		//Only update size if this is more than a usual argument
		if len(exp.PackedArgument) > 1 {
			exp.Size = len(exp.PackedArgument)
		}
	}

	return &exp, nil
}

func replaceSymbols(exp Exp, labels map[string]int) ([]int, error) {
	//char defs can't use symbols, so skip 'em
	if exp.Order != (Order{}) && exp.Order.Identifier == "char" {
		return exp.PackedArgument, nil
	}

	for i, _ := range exp.PackedArgument {
		//Skip anything that is non-symbolic
		maybeSymbol := exp.Argument[i]
		if !isSymbolic(maybeSymbol) {
			continue
		}

		//Is this a Certified Real Symbol(tm)?
		addr, ok := labels[maybeSymbol]

		if !ok {
			return nil, errors.New("I don't know what " + maybeSymbol + " means")
		}

		exp.PackedArgument[i] = addr
	}

	return exp.PackedArgument, nil
}

func hexToLGP(hex string) string {
	lgpHex := hex

	for k, v := range stdHexToLgpMap {
		lgpHex = strings.Replace(lgpHex, k, v, -1)
	}

	return lgpHex
}

func emitRawWords(words []int) string {
	rawHex := ""

	for _, w := range words {
		rawHex += hexToLGP(fmt.Sprintf("%04s", strconv.FormatInt(int64(w), 16)))
	}

	return rawHex
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func main() {
	//Read in the raw text
	lines, err := readLines(os.Args[1])
	if err != nil {
		fmt.Println("readLines: %s", err)
		return
	}

	//First pass: transform raw test into list of expressions
	address := 0
	var expressions []Exp
	var labels = make(map[string]int)

	for _, l := range lines {
		exp, err := parseLine(l)

		if err == nil {
			//Calculate address
			exp.Address = address
			address += exp.Size

			//If it's a label then add it to the lookup table
			if exp.Label != "" {
				labels[exp.Label] = exp.Address
			}

			//Add to list of expressions
			expressions = append(expressions, *exp)
		}
	}

	//Second pass: Substitute label addresses
	for i, e := range expressions {
		if e.Order != (Order{}) && len(e.Argument) > 0 {
			packedArgument, err := replaceSymbols(e, labels)

			if err != nil {
				fmt.Println(err)
				continue
			}

			expressions[i].PackedArgument = packedArgument
		}
	}

	fmt.Println(expressions)

	//And finally: EMIT!
	for _, e := range expressions {
		//Is this a real order?
		if e.Order != (Order{}) && e.Order.Value != "" {
			fmt.Println(e.Order.Value, emitRawWords(e.PackedArgument))
		} else {
			fmt.Println(emitRawWords(e.PackedArgument))
		}
	}
}
