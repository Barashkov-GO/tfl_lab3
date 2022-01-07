package main

import (
	"fmt"
	"os"
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

	nt  Nterm
	str string //	[a-z]
	t   []Term
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
			r.t = append(r.t, TermInit(lr[1][:indSep]))
			lr[1] = lr[1][indSep:]
		}
	}
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
	isTerm bool
}

func getTree(cfg CFG) (t *Tree) {
	rules := cfg.rules
	t.value = rules[0].nt.str
	for _, r := range rules {
		f, tr := treeSearch(r.nt, t)
		if f {
			t = tr
		}
		for _, s := range r.t {
			var newTree Tree
			newTree.value = s.nt.str
			t.subs = append(t.subs, &newTree)
		}
	}
	return
}

func printTree(t *Tree) {
	fmt.Println(t.value)
	for _, v := range t.subs {
		printTree(v)
	}
}

func treeSearch(nt Nterm, tr *Tree) (bool, *Tree) {
	if tr.value != nt.str {
		for _, v := range tr.subs {
			treeSearch(nt, v)
		}
	} else {
		return true, tr
	}
	return false, nil
}

func regAnalysis() {
	/*
		Анализ регулярных подмножеств грамматики.
		Нахождение максимальных множеств M i нетерминалов
		V j таких, что все правые части правил вида V j → . . .
		содержат только нетерминалы из M i , причём все эти
		части праволинейны.
	*/

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
	t := getTree(cfg)
	printTree(t)
	return
}

const TESTS_COUNT = 1

func main() {
	//var str string // from test file

	for i := 1; i <= TESTS_COUNT; i++ {
		cfg := read("tests/test" + strconv.Itoa(i) + ".txt")
		fmt.Println(cfg.toString())
		//getTree(cfg)
		//printNoInfo(recProbablyReg(checkMinWays(treeUnpacking(regAnalysis(cfg)))))
	}
	//

}
