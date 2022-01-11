package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	_ "sort"
	"strconv"
	"strings"
)

type Nterm struct {
	//	⟨nterm⟩ ::= [A–Z][0–9] ∗

	str string
}

func NtermInit(str string) (n Nterm) {
	n.str = str
	return
}

type Term struct {
	//	⟨term⟩ ::= ⟨nterm⟩ | [a–z]

	nt  Nterm
	str string //	[a-z]
}

func TermInit(str string) (t Term) {
	isNt, _ := regexp.MatchString("[a-z]", str)
	if !isNt {
		t.nt = NtermInit(str)
	} else {
		t.str = str
	}
	return
}

type Rule struct {
	//	⟨rule⟩ ::= ⟨nterm⟩->[a–z]⟨term⟩ ∗

	nt    Nterm
	str   string //	[a-z]
	t     []Term
	isSLG bool // A -> aB | A -> a
}

func RuleSLG(r Rule) bool {
	for i := 0; i < len(r.t)-1; i++ {
		if r.t[i].nt.str != "" {
			return false
		}
	}
	return true
}

func RuleInit(str string) (r Rule) {
	lr := strings.Split(str, "->")
	r.nt = NtermInit(lr[0])  // parse nterm ->
	r.str = string(lr[1][0]) // parse [a-z]

	lr[1] = lr[1][1:]

	for len(lr[1]) > 0 { // parse term
		first := string(lr[1][0])
		isAZ, _ := regexp.MatchString("[A-Z]", first)
		if !isAZ { // if [a-z]
			r.t = append(r.t, TermInit(first))
			lr[1] = lr[1][1:]
		} else { // if not [a-z] == if [A-Z]
			indSep := 1
			if len(lr[1]) == 1 {
				r.t = append(r.t, TermInit(lr[1]))
				break
			}
			flag := true
			for flag { // while 0-9, moving forward to take the whole Term
				flag, _ = regexp.MatchString("[0-9]", string(lr[1][indSep]))
				indSep++
			}
			r.t = append(r.t, TermInit(lr[1][:indSep-1]))
			lr[1] = lr[1][indSep-1:]
		}
	}
	r.isSLG = RuleSLG(r)
	return
}

type CFG struct {
	//	⟨grammar⟩ ::= ⟨rule⟩ +

	rules []Rule
}

func CFGInit(str string) (cfg CFG) {
	strs := strings.Split(str, "\n")
	for _, s := range strs {
		cfg.rules = append(cfg.rules, RuleInit(s))
	}
	return
}

func (cfg CFG) toString() (str string) {
	str = "CFG:\n"
	for _, r := range cfg.rules {
		str += "\t" + r.nt.str + " -> " + r.str
		for _, v := range r.t {
			str += v.nt.str + v.str
		}
		str += "\n"
	}
	str += "-----------------------\n"
	return
}

type Tree struct {
	value  string
	subs   []*Tree
	number int
}

var cnt = 0
var wasEnding = false
var nonTermPath []string

func appendTree(subs []*Tree, value string) []*Tree {
	var nT Tree
	nT.value = value
	nT.number = cnt + 1
	cnt++
	return append(subs, &nT)
}

func isInPath(nonTerm string) bool {
	for _, v := range nonTermPath {
		if v == nonTerm {
			return true
		}
	}
	return false
}

func getTree(cfg CFG, nonTerm string, baseNonTerm string, F1 *[]Term, F2 *[]Term) (t Tree) {
	t.value = nonTerm
	var indices []int
	for i, r := range cfg.rules { // находим все индексы, правила по которым в левой части содержат данный нетерминал
		if r.nt.str == nonTerm {
			indices = append(indices, i)
		}
	}
	for _, ind := range indices { // идем по всем правилам данного нетерминала
		rule := cfg.rules[ind]
		t.subs = appendTree(t.subs, rule.str) // добавляем первый терминал как листик
		var tN Term
		tN.str = rule.str
		*F1 = append(*F1, tN)
		for _, term := range rule.t { // идем по термам
			if term.str != "" { // если встретили терминал
				t.subs = appendTree(t.subs, term.str)
				if !wasEnding {
					var tN Term
					tN.str = term.str
					*F1 = append(*F1, tN)
				} else {
					var tN Term
					tN.str = term.str
					*F2 = append(*F2, tN)
				}
			} else { // если встретили нетерминал
				if isInPath(term.nt.str) { // если по этому нетерминалу уже строили поддерево
					t.subs = appendTree(t.subs, term.nt.str)
				} else {
					nonTermPath = append(nonTermPath, term.nt.str)
					if term.nt.str == baseNonTerm { // если нетерминал начальный
						if wasEnding {
							var tN Term
							tN.nt.str = term.nt.str
							*F2 = append(*F2, tN)
						}
						wasEnding = true
						t.subs = appendTree(t.subs, term.nt.str)
					} else { // если нетерминал не начальный
						if !wasEnding { // если начальный нетерминал еще не нашли
							nT := getTree(cfg, term.nt.str, baseNonTerm, F1, F2)
							nT.number = cnt + 1
							cnt++
							t.subs = append(t.subs, &nT)
						} else { // если начальный нетерминал уже нашли
							var tN Term
							tN.nt.str = term.nt.str
							*F2 = append(*F2, tN)
							t.subs = appendTree(t.subs, term.nt.str)
						}
					}
				}
			}
		}

	}
	return t
}

func checkF2(F2 []Term, m map[string]Nterm) bool {
	if len(F2) == 0 {
		return false
	}
	for _, val := range F2 {
		if val.str == "" {
			_, f2 := m[val.nt.str]
			if !f2 {
				return false
			}
		}
	}
	return true
}

func checkF1F2Plus(cfg CFG, F1 []Term, F2 []Term, F2Start []Term) bool {
	if len(F1) == 0 {
		return true
	}
	if len(F2) == 0 {
		if len(F2Start) != 0 {
			return checkF1F2Plus(cfg, F1, F2Start, F2Start)
		} else {
			return false
		}
	}
	if F2[0].str != "" {
		if F1[0].str != F2[0].str {
			return false
		} else {
			return checkF1F2Plus(cfg, F1[1:], F2[1:], F2Start)
		}
	} else {
		f := false
		for _, r := range cfg.rules {
			if r.nt.str == F2[0].nt.str {
				if r.t[0].str == F1[0].str {
					F1 = F1[1:]
					F2 = append(r.t, F2...)
					if checkF1F2Plus(cfg, F1, F2, F2Start) {
						f = true
					}
				}
			}
		}
		if !f {
			return false
		}
	}
	return true
}

var path map[string]string

func getAllTerminalStrings(cfg CFG, indBegin int, nonTerm string, out *map[string]string) bool {
	var cfgNew CFG
	for i := indBegin; i < len(cfg.rules); i++ {
		cfgNew.rules = append(cfgNew.rules, cfg.rules[i])
	}
	for i := 0; i < indBegin; i++ {
		cfgNew.rules = append(cfgNew.rules, cfg.rules[i])
	}
	var outStr string
	f := getTerminalString(cfgNew, nonTerm, &outStr)

	_, b := (*out)[outStr]
	if !b {
		(*out)[outStr] = outStr
	}
	return f
}

// 3
func getTerminalString(cfg CFG, nonTerm string, out *string) bool {
	var F1 []Term
	var F2 []Term
	t := getTree(cfg, nonTerm, nonTerm, &F1, &F2)
	for _, ch := range t.subs {
		f, _ := regexp.MatchString("[a-z]", ch.value)
		if f {
			*out += ch.value
		} else {
			_, b := path[ch.value]
			if b {
				return false
			} else {
				path[ch.value] = ch.value
				return getTerminalString(cfg, ch.value, out)
			}
		}
	}
	return true
}

// 4
func checkRegular(cfg CFG, nonTerm string, reg *[]string, probablyReg *[]string) {
	/*
		Рекурсивно замыкаем множества регулярных и возможно
		регулярных нетерминалов.
		Если при переписывании
		нетерминала B все правые части содержат только
		регулярные нетерминалы, он регулярен. Если регулярные
		и возможно регулярные — возможно регулярен.
	*/
	fReg := true
	fProbablyReg := true
	for _, r := range cfg.rules {
		if r.nt.str == nonTerm {
			for _, t := range r.t {
				fReg2 := false
				fProbablyReg2 := false
				for _, regs := range *reg {
					if regs == t.nt.str {
						fReg2 = true
					}
				}
				if !fReg2 {
					fReg = false
				}
				for _, regs := range *probablyReg {
					if regs == t.nt.str {
						fProbablyReg2 = true
					}
				}
				if !fProbablyReg2 {
					fProbablyReg = false
				}
			}
		}
	}
	if fProbablyReg {
		*probablyReg = append(*probablyReg, nonTerm)
	}
	if fReg {
		*reg = append(*reg, nonTerm)
	}
}

func getChildren(str string, cfg CFG, m *map[string]bool) { // получить все достижимые нетерминалы из дерева данного нетерминала
	for _, r := range cfg.rules {
		if r.nt.str == str {
			for _, nt := range r.t {
				if len(nt.str) == 0 {
					if !(*m)[nt.nt.str] {
						(*m)[nt.nt.str] = true
						getChildren(nt.nt.str, cfg, m)
					}
				}
			}
		}
	}
	return
}

func printTree(t Tree) (str []string) {
	for _, v := range t.subs {
		str = append(str, "\t"+strconv.Itoa(t.number)+"[label=\""+t.value+"\"]\n")
		str = append(str, "\t"+strconv.Itoa(v.number)+"[label=\""+v.value+"\"]\n")
		for _, val := range printTree(*v) {
			str = append(str, val)
		}
	}
	for _, v := range t.subs {
		str = append(str, "\t"+strconv.Itoa(t.number)+"->"+strconv.Itoa(v.number)+"\n")
		for _, val := range printTree(*v) {
			str = append(str, val)
		}
	}
	return
}

func getStringFromMap(str []string) (out string) {
	unique := make(map[string]int)
	for _, v := range str {
		a, f := unique[v]
		if !f {
			unique[v] = 1
		} else {
			unique[v] = a + 1
		}
	}
	for _, v := range str {
		if unique[v] != 0 {
			out += v
			unique[v] = 0
		}
	}
	return
}

func regAnalysis(cfg CFG) (out map[string]Nterm) {
	/*
		Анализ регулярных подмножеств грамматики.
		Нахождение максимальных множеств M i нетерминалов
		V j таких, что все правые части правил вида V j → . . .
		содержат только нетерминалы из M i , причём все эти
		части праволинейны.
	*/
	out = make(map[string]Nterm)
	for _, r := range cfg.rules {
		if checkNtermForTerminalRule(cfg, r.nt) {
			out[r.nt.str] = r.nt
			break
		}
	}
	for _, r := range cfg.rules {
		if checkNtermForLastNtermOfRule(cfg, r.nt, out) {
			out[r.nt.str] = r.nt
		} else {
			delete(out, r.nt.str)
		}
	}
	return
}

func checkNtermForLastNtermOfRule(cfg CFG, nt Nterm, out map[string]Nterm) bool {
	var indices []int
	for ind, rule := range cfg.rules { // нашли все правила с этим нетерминалом
		if rule.nt.str == nt.str {
			indices = append(indices, ind)
		}
	}
	f := true
	for _, ind := range indices {
		rule := cfg.rules[ind]
		for i := 0; i < len(rule.t)-1; i++ {
			if rule.t[i].nt.str != "" {
				return false
			}
		}
		if len(rule.t) != 0 {
			lastNterm := rule.t[len(rule.t)-1]
			_, b := out[lastNterm.nt.str]
			if lastNterm.nt.str != "" && !b {
				f = false
			}
		}
	}
	return f
}

func mapSearch(m map[string]Nterm, nterm Nterm) bool {
	_, f := m[nterm.str]
	return f
}

func checkNtermForTerminalRule(cfg CFG, nt Nterm) bool {
	var indices []int
	for ind, rule := range cfg.rules { // нашли все правила с этим нетерминалом
		if rule.nt.str == nt.str {
			indices = append(indices, ind)
		}
	}
	for _, ind := range indices { // добавили нетерминалы, которые переходят в терминальную строку
		f := true
		rule := cfg.rules[ind]
		for _, term := range rule.t {
			if term.nt.str != "" {
				f = false
			}
		}
		if f {
			return true
		}
	}
	return false
}

func preparing(str string) (out string) {
	strs := strings.Split(str, "\n")
	sort.Strings(strs)
	for _, s := range strs {
		out += s + "\n"
	}
	out = out[:len(out)-1]
	out = strings.ReplaceAll(out, " ", "")
	return
}

func read(path string) (cfg CFG) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	data := make([]byte, 64)
	n, _ := file.Read(data)
	cfg = CFGInit(preparing(string(data[:n])))
	return
}

func write(path string, res string) {
	file, _ := os.Create(path)
	file.Write([]byte("digraph G {\n"))
	file.Write([]byte(res))
	file.Write([]byte("\n}"))
	defer file.Close()
}

func graphViz(i int, t Tree) {
	write("results/test"+strconv.Itoa(i)+"_"+t.value+".gv", getStringFromMap(printTree(t)))
	cmd := exec.Command("dot",
		"-Tpng",
		"results/test"+strconv.Itoa(i)+"_"+t.value+".gv",
		"-o",
		"results/images/test"+strconv.Itoa(i)+"_"+t.value+".png")
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func printAnswer(r []string, nr []string, pr []string) {
	m := make(map[string]bool)
	for _, v := range r {
		m[v] = true
		fmt.Println("Язык " + v + " регулярен")
	}
	for _, v := range pr {
		_, b := m[v]
		if !b {
			fmt.Println("Язык " + v + " возможно регулярен")
		}
	}
	for _, v := range nr {
		_, b := m[v]
		if !b {
			fmt.Println(
				"Дерево накачки нетерминала " +
					v +
					" подозрительно на нерегулярную накачку")
		}
	}
}

const TestsStart = 1
const TestsCount = 8

func main() {
	for i := TestsStart; i <= TestsCount; i++ {
		var (
			regular            []string
			probablyNonRegular []string
			probablyRegular    []string
		)

		fmt.Println("TEST " + strconv.Itoa(i))
		cfg := read("tests/test" + strconv.Itoa(i) + ".txt")
		childrenS := make(map[string]bool)
		getChildren("S", cfg, &childrenS)
		reg := regAnalysis(cfg)
		for v, _ := range childrenS {
			_, b := reg[v]
			if b {
				regular = append(regular, v)
				continue
			}
			cnt = 0
			var F1, F2 []Term
			wasEnding = false
			nonTermPath = nonTermPath[0:0]
			t1 := getTree(cfg, v, v, &F1, &F2)
			//fmt.Println(v)
			//for _, v := range F1 {
			//	fmt.Println("\tF1", v.str)
			//}
			//for _, v := range F2 {
			//	fmt.Println("\tF2", v.str, v.nt.str)
			//}
			if checkF2(F2, regAnalysis(cfg)) {
				if !checkF1F2Plus(cfg, F1, F2, F2) {
					// если Ф1 не входит в Ф2+
					probablyNonRegular = append(probablyNonRegular, v)
				} else {
					// если Ф1 входит в Ф2+
					str := make(map[string]string)
					path = make(map[string]string)
					nonTermPath = nonTermPath[0:0]
					for j := 0; j < len(cfg.rules); j++ {
						// перебираем все циклические сдвиги правил,
						// чтобы собрать все терминальные строки
						getAllTerminalStrings(cfg, 0, v, &str)
					}
					isTerminalStringsInF2 := true
					for a, _ := range str {
						// перебираем все терминальные строки из получившейся мапы
						var FF1 []Term
						for _, c := range a {
							// делаем из строки массив термов
							var T Term
							T.str = string(c)
							FF1 = append(FF1, T)
						}
						if !checkF1F2Plus(cfg, FF1, F2, F2) {
							// если какой либо из новых Ф1 не входит в Ф2+
							isTerminalStringsInF2 = false
						}
					}
					if isTerminalStringsInF2 {
						// если все новые Ф1 входят в Ф2+, то возможна регулярности
						_, b := reg[v]
						if b {
							// если нетерминал есть в множестве Mi, то регулярен
							regular = append(regular, v)
						} else {
							// если нетерминала нет в множестве Mi, то возможно регулярен
							probablyRegular = append(probablyRegular, v)
						}
					}
				}
			}
			graphViz(i, t1)
		}
		for v, _ := range childrenS {
			checkRegular(cfg, v, &regular, &probablyRegular)
		}
		//fmt.Println(regular, probablyRegular, probablyNonRegular)
		l1 := len(regular)
		l2 := len(probablyRegular)
		l3 := len(probablyNonRegular)
		checkRegular(cfg, "S", &regular, &probablyRegular)
		if l1 == len(regular) && l2 == len(probablyRegular) && l3 == 0 {
			/*
				Если рекурсивное замыкание не дало никакой
				информации об исходном нетерминале S, но не было и
				подозрительных нерегулярных накачек S, сообщаем, что
				регулярность языка не удалось определить.
			*/
			fmt.Println("Регулярность языка не удалось определить")
		}
		printAnswer(regular, probablyNonRegular, probablyRegular)
	}
}
