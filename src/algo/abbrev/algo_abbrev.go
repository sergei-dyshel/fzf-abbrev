package abbrev

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	// "time"
	"unicode/utf8"

	"github.com/sergei-dyshel/fzf-abbrev/src/algo"
	"github.com/sergei-dyshel/fzf-abbrev/src/util"
)

const abbrevDebugToFile = false
const abbrevDebugFilename = ".fzf.debug"
const spaceChars = "\t\n\v\f\r "
const pathSeparators = "\\/"

const subWordStartMatchPenalty = 1
const wordStartMatchPenalty = 2
const notLastWordPenalty = 20
const skippedWordStartPenalty = 5
const skippedSubWordStartPenalty = 5

// Options - Abbrev matcher options
type Options struct {
	noScore           bool
	UseByDefault      bool
	preferLastWord    bool
	wordSeparators    string
	subWordSeparators string
	Debug             bool
	FirstMatchOnly    bool
	maxMatchLen       int
}

var Opts = defaultOptions

var defaultOptions = Options{
	UseByDefault:      true,
	noScore:           false,
	preferLastWord:    false,
	wordSeparators:    "",
	subWordSeparators: "_-",
	Debug:             false,
	FirstMatchOnly:    false,
	maxMatchLen:       180,
}

func (opts *Options) Parse(args string) {
	*opts = defaultOptions
	for _, arg := range strings.Split(args, ",") {
		switch arg {
		case "no-default":
			opts.UseByDefault = false
		case "debug":
			opts.Debug = true
		case "file-paths":
			opts.wordSeparators = pathSeparators
			opts.preferLastWord = true
		case "fast":
			opts.FirstMatchOnly = true
		}
	}
}

func (opts *Options) Default() {
	*opts = defaultOptions
}

type charType int

const (
	noChar charType = iota
	charWordSeparator
	charSubWordSeparator
	charLower
	charUpper
	charDigit
)

type boundaryType int

const (
	undeterminedBoundary boundaryType = iota
	noBoundary
	wordStart
	subWordStart
	subWordSeparator
	wordSeparator
)

func calcCharType(input []rune, index int, cache *inputCache) charType {
	if index < 0 || index >= len(input) {
		return noChar
	}
	cached := cache.charTypes[index]
	if cached != noChar {
		return cached
	}
	ch := input[index]
	var chType charType
	switch {
	case ch >= 'a' && ch <= 'z':
		chType = charLower
	case ch >= 'A' && ch <= 'Z':
		chType = charUpper
	case ch >= '0' && ch <= '9':
		chType = charDigit
	case Opts.subWordSeparators != "" &&
		strings.ContainsRune(Opts.subWordSeparators, ch):

		chType = charSubWordSeparator
	case Opts.subWordSeparators != "":
		if strings.ContainsRune(Opts.subWordSeparators, ch) {
			chType = charSubWordSeparator
		} else {
			chType = charWordSeparator
		}
	case Opts.wordSeparators != "":
		if strings.ContainsRune(Opts.wordSeparators, ch) {
			chType = charWordSeparator
		} else {
			chType = charSubWordSeparator
		}
	default:
		panic(fmt.Sprintf("Could not determine char class of '%c", ch))
	}

	cache.charTypes[index] = chType
	return chType
}

func calcBoundaryType(input []rune, index int, cache *inputCache) boundaryType {
	cached := cache.boundaryTypes[index]
	if cached != undeterminedBoundary {
		return cached
	}
	curr := calcCharType(input, index, cache)
	prev := calcCharType(input, index-1, cache)
	next := calcCharType(input, index+1, cache)

	var bType boundaryType

	switch {
	case curr == charWordSeparator:
		bType = wordSeparator
	case (prev == noChar || prev == charWordSeparator) &&
		(curr != charWordSeparator):

		bType = wordStart
	case curr == charSubWordSeparator:
		bType = subWordSeparator
	case prev == charSubWordSeparator:
		bType = subWordStart
	case curr == charDigit:
		bType = subWordSeparator
	// case prev != charDigit && curr == charDigit:
	// 	bType = subWordStart
	case (prev != charLower && prev != charUpper) &&
		(curr == charLower || curr == charUpper):

		bType = subWordStart
	case prev != charUpper && curr == charUpper:
		bType = subWordStart
	case curr == charUpper && next == charLower:
		bType = subWordStart
	default:
		bType = noBoundary
	}

	cache.boundaryTypes[index] = bType
	return bType
}

func debugf(format string, a ...interface{}) {
	f, _ := os.OpenFile(abbrevDebugFilename,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintf(w, format, a...)
}

func markPosInStr(str string, pos []int) string {
	var res string
	lastPos := 0
	for _, ind := range pos {
		if ind > 0 {
			res += str[lastPos:ind]
		}
		res += fmt.Sprintf("{%c}", str[ind])
		lastPos = ind + 1
	}
	res += str[lastPos:]
	return res
}

func toLower(char rune) rune {
	if char >= 'A' && char <= 'Z' {
		return char + 32
	}
	return char
}

type result struct {
	score int
	pos   []int
}

type inputCache struct {
	charTypes     []charType
	boundaryTypes []boundaryType
}

func checkMatch(input []rune, pattern []rune) *result {
	cache := new(inputCache)
	cache.charTypes = make([]charType, len(input))
	cache.boundaryTypes = make([]boundaryType, len(input))

	// start := time.Now()
	res := recursiveMatch(input, 0, pattern, 0, len(input), cache)
	// elapsed := time.Since(start)
	// if elapsed > 100*time.Millisecond {
	// 	fmt.Printf("elapsed %f %d %d %s\n", elapsed.Seconds(), len(input), len(pattern), string(input))
	// 	// fmt.Printf("elapsed: %d pattern: '%s' score: %d input: '%s\n", elapsed, string(pattern),
	// 	// 	res.score, string(input))
	// }
	if res != nil && Opts.Debug {
		debugf("pattern: '%s' score: %d input: '%s\n", string(pattern),
			res.score, markPosInStr(string(input), res.pos))
	}
	return res
}

func better(res1, res2 *result, addPenalty int) bool {
	if res1 == nil {
		return false
	}
	res1.score += addPenalty
	return res2 == nil || res1.score < res2.score
}

func calcCharMatchPenalty(match boundaryType) int {
	switch match {
	case subWordStart:
		return subWordStartMatchPenalty
	case wordStart:
		return wordStartMatchPenalty
	default:
		return 0
	}
}

func checkNotLastWordPenalty(input []rune, startIdx int, cache *inputCache) int {
	if !Opts.preferLastWord || Opts.FirstMatchOnly {
		return 0
	}
	res := 0
	for i := startIdx; i < len(input); i++ {
		matchType := calcBoundaryType(input, i, cache)
		if matchType == wordStart {
			res += notLastWordPenalty
		}
	}
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func recursiveMatch(input []rune, inputIdx int,
	pat []rune, patIdx int, firstMatchedIndex int,
	cache *inputCache) *result {

	if patIdx == len(pat) {
		res := result{0, make([]int, len(pat))}
		res.score += checkNotLastWordPenalty(input, inputIdx, cache)
		return &res
	}
	if inputIdx == len(input) {
		return nil
	}
	patCh := toLower(pat[patIdx])

	var bestRes *result

	if inputIdx+1-firstMatchedIndex > Opts.maxMatchLen {
		return nil
	}
	if toLower(input[inputIdx]) == patCh {
		penalty := 0
		if patIdx == 0 {
			firstMatchedIndex = inputIdx
			penalty += wordStartMatchPenalty
		}
		res := recursiveMatch(input, inputIdx+1, pat, patIdx+1,
			min(firstMatchedIndex, inputIdx), cache)
		if res != nil && Opts.FirstMatchOnly {
			res.pos[patIdx] = inputIdx
			return res
		}
		if better(res, bestRes, penalty) {
			res.pos[patIdx] = inputIdx
			bestRes = res
		}
	}

	skippedStart := noBoundary

	boundary := calcBoundaryType(input, inputIdx, cache)
	if boundary == wordStart {
		skippedStart = wordStart
	} else if boundary == subWordStart && skippedStart != wordStart {
		skippedStart = subWordStart
	}

	accumPenalty := 0
	for i := inputIdx + 1; i < len(input); i++ {
		if i+1-firstMatchedIndex > Opts.maxMatchLen {
			break
		}
		boundary := calcBoundaryType(input, i, cache)
		if boundary == noBoundary {
			continue
		}
		charPenalty := calcCharMatchPenalty(boundary)
		if toLower(input[i]) == toLower(patCh) {
			res := recursiveMatch(input, i+1, pat, patIdx+1,
				min(firstMatchedIndex, i), cache)
			if res != nil && Opts.FirstMatchOnly {
				res.pos[patIdx] = i
				return res
			}
			addPenalty := 0
			// TODO: improve preferLastWord scoring
			if Opts.preferLastWord && patIdx > 0 && boundary == wordStart {
				addPenalty += notLastWordPenalty
			}
			if boundary == subWordStart {
				if skippedStart == wordStart {
					addPenalty += skippedWordStartPenalty
				} else if skippedStart == subWordStart {
					addPenalty += skippedSubWordStartPenalty
				}
			}
			if better(res, bestRes, accumPenalty+charPenalty+addPenalty) {
				res.pos[patIdx] = i
				bestRes = res
			}
		}
		if boundary == wordStart {
			skippedStart = wordStart
		} else if boundary == subWordStart && skippedStart != wordStart {
			skippedStart = subWordStart
		}
		if patIdx > 0 {
			accumPenalty += charPenalty
		}
	}
	return bestRes
}

// TODO
func Match(caseSensitive bool, normalize bool, forward bool,
	text *util.Chars, pattern []rune, withPos bool,
	slab *util.Slab) (algo.Result, *[]int) {

	input := text.ToRunes()
	// fmt.Println(string(runes), string(pattern))
	for _, ch := range input {
		if ch >= utf8.RuneSelf {
			return algo.FuzzyMatchV2(caseSensitive, normalize, forward, text,
				pattern, withPos, slab)
		}
	}
	// TODO: handle pattern of sizes 0 in fuzzy match
	if len(pattern) == 0 {
		return algo.FuzzyMatchV2(caseSensitive, normalize, forward, text,
			pattern, withPos, slab)
	}

	res := checkMatch(input, pattern)

	if res == nil {
		return algo.Result{Start: -1, End: -1, Score: 0}, nil
	}
	pos := res.pos

	// proper init
	algoRes := algo.Result{Start: pos[0], End: pos[len(pos)-1] + 1,
		Score: 1000 - res.score}
	reverse(pos)

	return algoRes, &pos
}

func reverse(a []int) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}
