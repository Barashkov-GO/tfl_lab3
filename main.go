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
	if len(r.t) == 0 {
		return true
	}
	if len(r.t) == 1 && r.t[0].str == "" {
		return true
	}
	return false
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
var pathToRoot []string

func appendTree(subs []*Tree, value string) []*Tree {
	var nT Tree
	nT.value = value
	nT.number = cnt + 1
	cnt++
	return append(subs, &nT)
}

func isInPath(nonTerm string) bool {
	for _, v := range pathToRoot {
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
					tN.nt.str = term.str
					*F2 = append(*F2, tN)
				}
			} else { // если встретили нетерминал
				if term.nt.str == baseNonTerm { // если нетерминал начальный
					if wasEnding {
						var tN Term
						tN.nt.str = term.nt.str
						*F2 = append(*F2, tN)
					}
					wasEnding = true
					t.subs = appendTree(t.subs, term.nt.str)
					pathToRoot = append(pathToRoot, term.nt.str)
				} else { // если нетерминал не начальный
					if isInPath(term.nt.str) {
						return
					}
					if !wasEnding { // если начальный нетерминал еще не нашли
						nT := getTree(cfg, term.nt.str, baseNonTerm, F1, F2)
						nT.number = cnt + 1
						cnt++
						t.subs = append(t.subs, &nT)
						if wasEnding {
							pathToRoot = append(pathToRoot, term.nt.str)
						}
					} else { // если начальный нетерминал уже нашли
						var tN Term
						tN.nt.str = term.nt.str
						*F2 = append(*F2, tN)
						t.subs = appendTree(t.subs, term.nt.str)
						//pathToRoot = append(pathToRoot, term.nt.str)
					}
				}
			}
		}

	}
	return t
}

func checkF2(F2 []Term, m map[string]Nterm) bool {
	for _, val := range F2 {
		v := val.nt.str
		f1, _ := regexp.MatchString("[a-z]", v)
		if !f1 {
			_, f2 := m[v]
			if !f2 {
				return false
			}
		}
	}
	return true
}

func checkF1F2Plus(cfg CFG, F1 []Term, F2 []Term) bool {
	if len(F2) == 0 {
		return false
	}
	if F2[0].str != "" {
		if F1[0].str != F2[0].str {
			return false
		} else {
			return checkF1F2Plus(cfg, F1[1:], F2[1:])
		}
	} else {
		for _, r := range cfg.rules {
			if r.nt.str == F2[0].nt.str {
				if r.t[0].str == F1[0].str {
					F1 = F1[1:]
					F2 = append(r.t, F2...)
					return checkF1F2Plus(cfg, F1, F2)
				}
			}
		}
	}
	return true
}

func printTreeArray(F []string) {
	fmt.Println("\tF\n")
	for _, v := range F {
		fmt.Println("\t\t" + v)
	}
}

func treeSearch(t *Tree, str string) *Tree {
	if t.value == str {
		return t
	}
	for _, v := range t.subs {
		treeSearch(v, str)
	}
	return t
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
		if checkNtermNterm(cfg, r.nt) {
			out[r.nt.str] = r.nt
		}
	}
	for _, r := range cfg.rules {
		if checkNterm(cfg, r.nt, out) {
			out[r.nt.str] = r.nt
		}
	}
	return
}

func mapSearch(m map[string]Nterm, nterm Nterm) bool {
	_, f := m[nterm.str]
	return f
}

func checkNtermNterm(cfg CFG, nt Nterm) bool {
	for _, r := range cfg.rules {
		if r.nt.str == nt.str {
			if len(r.t) != 0 {
				return false
			}
		}
	}
	return true
}

func checkNterm(cfg CFG, nt Nterm, m map[string]Nterm) bool { // подходит ли нетерминал со всеми своими правилами под условия
	for _, r := range cfg.rules {
		if r.nt.str == nt.str {
			if !r.isSLG {
				return false
			}
			if len(r.t) == 1 && r.t[0].nt.str != nt.str && !mapSearch(m, r.t[0].nt) {
				return false
			}
		}
	}
	return true
}

func treeUnpacking() {
	/*
		Развёртка дерева левосторонних разборов исходной
		грамматики. Для каждого достижимого из стартового
		нетерминала A строим дерево развёртки до первых
		накачек вида Φ 1 AΦ 2 , где Φ 1 — терминальная строка.
		Если оказалось, что Φ 2 состоит только из терминалов или
		регулярных нетерминалов (входящих в какое-нибудь из
		M i ), тогда проверяем, входит ли Φ 1 в язык Φ +
		2 . Если не
		входит, тогда выводим дерево накачки нетерминала A как
		подозрительное на нерегулярную накачку.
	*/

}

func checkMinWays() {
	/*
		Если Φ 1 ∈ L(Φ +
		2 ), тогда проверяем все кратчайшие
		конечные пути развёртки A до терминальной строки на
		вхождение в L(Φ +
		2 ). Если построенные на них строки
		также входят в L(Φ +
		2 ), сообщаем о возможной
		регулярности языка A. Если A ∈ M i , сразу сообщаем о
		его регулярности.
	*/

}

func recProbablyReg() {
	/*
		Рекурсивно замыкаем множества регулярных и возможно
		регулярных нетерминалов. Если при переписывании
		нетерминала B все правые части содержат только
		регулярные нетерминалы, он регулярен. Если регулярные
		и возможно регулярные — возможно регулярен.
	*/

}

func printNoInfo() {
	/*
		Если рекурсивное замыкание не дало никакой
		информации об исходном нетерминале S, но не было и
		подозрительных нерегулярных накачек S, сообщаем, что
		регулярность языка не удалось определить.
	*/

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

const TestsCount = 5
const TestsStart = 1

func main() {
	for i := TestsStart; i <= TestsCount; i++ {
		fmt.Println("TEST " + strconv.Itoa(i))
		cfg := read("tests/test" + strconv.Itoa(i) + ".txt")
		m := make(map[string]bool)
		getChildren("S", cfg, &m)
		for v, _ := range m {
			cnt = 0
			var F1, F2 []Term
			wasEnding = false
			pathToRoot = pathToRoot[0:0]
			t1 := getTree(cfg, v, v, &F1, &F2)
			//printTreeArray(F1)
			//printTreeArray(F2)
			//fmt.Println(F2, "\t-\t", checkF2(F2, regAnalysis(cfg)))
			if checkF2(F2, regAnalysis(cfg)) {
				if !checkF1F2Plus(cfg, F1, F2) {
					fmt.Println(
						"Дерево накачки нетерминала " +
							v +
							" подозрительно на нерегулярную накачку")
				}
			}
			graphViz(i, t1)
		}
	}
}
