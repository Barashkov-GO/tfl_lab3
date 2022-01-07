package main

import (
	"regexp"
	"strings"
)

type Nterm struct {
	//	⟨nterm⟩ ::= [A–Z][0–9] ∗

	str string
}

func Nterm_init(str string) (n Nterm) {
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
		t.nt = Nterm_init(str)
	} else {
		t.str = str
	}
	return
}

type Rule struct {
	//	⟨rule⟩ ::= ⟨nterm⟩->[a–z]⟨term⟩ ∗

	nt  Nterm
	str string //	[a-z]
	t   Term
}

func RuleInit(str string) (r Rule) {
	lr := strings.Split(str, "->")
	r.nt = Nterm_init(lr[0])
	var indTerm int
	var tempStr []byte
	for i, c := range lr[1] {
		isAZ, _ := regexp.MatchString("a-z", string(c))
		if isAZ {
			tempStr = append(tempStr, byte(c))
		} else {
			indTerm = i
			break
		}
	}
	r.str = string(tempStr)
	r.t = TermInit(lr[1][indTerm:])
	return
}

type CFG struct {
	//	⟨grammar⟩ ::= ⟨rule⟩ +

	rules []Rule
}

func CFGInit(str string) (cfg CFG) {
	strs := strings.Split(str, "/n")
	for _, s := range strs {
		cfg.rules = append(cfg.rules, RuleInit(s))
	}
	return
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

func main() {
	var str string // from test file

	printNoInfo(recProbablyReg(checkMinWays(treeUnpacking(regAnalysis(CFGInit(str))))))

}
