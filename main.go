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
	subs   map[string]*Tree
	isTerm bool
}

func getTree(cfg CFG, str string, baseStr string) (t Tree) {
	t.value = str
	t.subs = make(map[string]*Tree)
	var inds []int
	for i, r := range cfg.rules { // находим все индексы, правила по которым в левой части содержат данный нетерминал
		if r.nt.str == str {
			inds = append(inds, i)
		}
	}
	for _, i := range inds { // идем по найденным индексам
		rule := cfg.rules[i]
		nT := getTree(cfg, rule.str, baseStr)
		t.subs[rule.str] = &nT
		for _, terms := range rule.t {
			var searchStr string
			if baseStr == terms.nt.str {
				var nT Tree
				nT.value = baseStr
				t.subs[baseStr] = &nT

				// добавить этот терм, все остальные как листья и сделать ретурн
				for _, termsNew := range rule.t {
					var nT Tree
					if termsNew.nt.str != "" {
						nT.value = termsNew.nt.str
					} else {
						nT.value = termsNew.str
					}
					t.subs[nT.value] = &nT
				}
				fmt.Println("Map for " + rule.nt.str)
				for i, v := range t.subs {
					fmt.Println(i + " - " + v.value)
				}
				return
			}
			if terms.nt.str != "" {
				searchStr = terms.nt.str
			} else {
				searchStr = terms.str
			}
			nT := getTree(cfg, searchStr, baseStr)
			t.subs[searchStr] = &nT

			fmt.Println("Map for " + rule.nt.str)
			for i, v := range t.subs {
				fmt.Println(i + " - " + v.value)
			}
		}
	}
	return t
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

func printTree(t Tree) (str map[string]string) {
	str = make(map[string]string)
	for _, v := range t.subs {
		str[t.value+"->"+v.value+"\n"] = t.value + "->" + v.value + "\n"
		for k, val := range printTree(*v) {
			str[k] = val
		}
	}
	return
}

func getStringFromMap(str map[string]string) (out string) {
	for _, v := range str {
		out += v
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

func printMapNterm(m map[string]Nterm) {
	fmt.Println("M:")
	for _, v := range m {
		fmt.Print(v.str + " ")
	}
	fmt.Println("---------------")
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
	fmt.Println(res)
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

const TestsCount = 4

func main() {
	for i := 4; i <= TestsCount; i++ {
		cfg := read("tests/test" + strconv.Itoa(i) + ".txt")
		m := make(map[string]bool)
		getChildren("S", cfg, &m)
		for v, _ := range m {
			t1 := getTree(cfg, v, v)
			//fmt.Println("Nonterminal tree for: " + t1.value)
			//for i, v := range t1.subs {
			//	fmt.Println(i + " - " + v.value)
			//}
			graphViz(i, t1)
		}
		//fmt.Println("Children nonterminals for " + "S" + ":\n")
		//for v, _ := range m {
		//	fmt.Println("\t" + v)
		//}

		printMapNterm(regAnalysis(cfg))
	}
}
