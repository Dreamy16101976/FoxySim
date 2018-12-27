/*  FoxySim - online DC & AC electric circuits simulator
    Version 1.0
    Copyright (C) 2017-2018 Alexey "FoxyLab" Voronin
    Email:    support@foxylab.com
    Website:  https://acdc.foxylab.com
    This software is licensed under the Creative Commons Attribution-NonCommercial-ShareAlike 4.0 (CC BY-NC-SA 4.0) License.
*/
 
package main

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"math/cmplx"
	"net/http"
	"strings"
	"time"
	"./libs"
	"strconv"
	"github.com/gonum/matrix/mat64"
)

var locales map[string]map[string]string

func msg(locale, key string) string {
	if v, ok := locales[locale]; ok {
		if v2, ok := v[key]; ok {
			return v2
		}
	}
	return ""
}

var dbg bool //флаг отладки

const port string = "7777" //порт

const salt string = "12345678"

//компонент схемы
type Element struct {
	name   string
	from   int
	to     int
	fValue float64
	cValue complex128
}

type ZElement struct {
	name   string
	from   int
	to     int
	cValue complex128
}

//индуктивная связь
type KElement struct {
	name string
	l1   string
	l2   string
	k    float64
	l1b  int
	l2b  int
}

//амперметр
type PAElement struct {
	name  string
	plus  int
	minus int
}

//вольтметр
type PVElement struct {
	name  string
	plus  int
	minus int
}

//ваттметр
type PWElement struct {
	name   string
	iplus  int
	iminus int
	uplus  int
	uminus int
}

//варметр
type PQElement struct {
	name   string
	iplus  int
	iminus int
	uplus  int
	uminus int
}

//фазометр
type PFElement struct {
	name   string
	iplus  int
	iminus int
	uplus  int
	uminus int
}

type SElement struct {
	name  string
	n1    int
	n2    int
	n3    int
	n4    int
	fGain float64
	cGain complex128
}

//узел
type Node struct {
	Name  int
	Count int
}

//симуляция
func calc(netlist string, lang string) string {
	var element Element     //компонент
	var zElement ZElement   //комплексное сопротивление
	var kElement KElement   //индуктивная связь
	var paElement PAElement //амперметр
	var pvElement PVElement //вольтметр
	var pwElement PWElement //ваттметр
	var pqElement PQElement //варметр
	var pfElement PFElement //фазометр
	var sElement SElement
	var node Node    //узел
	var mode string  //режим симуляции
	var w float64    //угловая частота (для режима AC)
	var brkFlag bool //флаг прерывания анализа списка соединений
	var nr int = 0   //кол-во резисторов
	var nl int = 0   //кол-во катушек
	var nc int = 0   //кол-во конденсаторов
	var nz int = 0   //кол-во комплексных сопротивлений
	var nv int = 0   //кол-во источников ЭДС
	var ni int = 0   //кол-во источников тока
	var nk int = 0   //кол-во индуктивных связей
	var npa int = 0  //кол-во амперметров
	var npv int = 0  //кол-во вольтметров
	var npw int = 0  //кол-во ваттметров
	var npq int = 0  //кол-во варметров
	var npf int = 0  //кол-во фазометров
	var ne int = 0   //кол-во
	var nf int = 0   //кол-во
	var ng int = 0   //кол-во
	var nh int = 0   //кол-во
	//списки компонентов
	var r_elements = make([]Element, 0)
	var l_elements = make([]Element, 0)
	var c_elements = make([]Element, 0)
	var z_elements = make([]ZElement, 0)
	var v_elements = make([]Element, 0)
	var i_elements = make([]Element, 0)
	var k_elements = make([]KElement, 0)
	var pa_elements = make([]PAElement, 0)
	var pv_elements = make([]PVElement, 0)
	var pw_elements = make([]PWElement, 0)
	var pq_elements = make([]PQElement, 0)
	var pf_elements = make([]PFElement, 0)
	var e_elements = make([]SElement, 0)
	var f_elements = make([]SElement, 0)
	var g_elements = make([]SElement, 0)
	var h_elements = make([]SElement, 0)
	var n int = 0 //кол-во узлов
	var m int = 0 //кол-во независимых источников ЭДС
	//временные переменные
	var int64Tmp int64
	var intTmp int
	var boolTmp bool
	var floatTmp float64
	var stringTmp string
	var absValue float64
	var argValue float64
	//флаги результатов
	var err error
	var success bool

	var tmp1DMatrix []complex128
	var zb [][]complex128
	var y [][]complex128
	var az [][]float64

	var angle string = "DEG"  //градусы по умолчанию
	var decfmt string = "FIX" //формат с фиксированной точкой по умолчанию
	var digits = 6

	//инициализация списка узлов
	var nodes = make([]Node, 0)
	node.Name = 0
	node.Count = 0
	nodes = append(nodes, node)
	//считывание списка соединений
	lines := strings.Split(netlist, "\n")
	//проверка на пустой список соединений
	if len(lines) == 0 {
		return "<font color=\"red\">" + msg(lang, "err00") + "</font>"
	}

	//анализ директив
	params := make(map[string]string)
	brkFlag = false
	for k := 0; k < len(lines); k++ {
		fields := strings.Fields(lines[k])
		if len(fields) != 0 { //если непустая строка
			ch := strings.ToUpper(fields[0])[0]
			if ch == '.' { //если директива
				if strings.ToUpper(fields[0]) == ".DC" {
					if len(fields) != 1 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					mode = "DC" //расчет цепи постоянного тока
				}
				if strings.ToUpper(fields[0]) == ".AC" {
					if len(fields) > 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err15") + "</font>"
					}
					if len(fields) == 2 {
						w, success = libs.Str2Freq(fields[1]) //определение угловой частоты синусоидального тока, рад/с
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err15") + "</font>"
						}
					} else {
						w = math.NaN() //частота не задана
					}
					mode = "AC" //расчет цепи синусоидального тока
				}
				if strings.ToUpper(fields[0]) == ".PARAM" {
					if len(fields) != 3 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					//
					params[strings.ToUpper(fields[1])] = fields[2] //запись для параметра пары имя-значение
				}
				if strings.ToUpper(fields[0]) == ".DEG" {
					if len(fields) != 1 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					angle = "DEG" //градусы
				}
				if strings.ToUpper(fields[0]) == ".RAD" {
					if len(fields) != 1 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					angle = "RAD" //радианы
				}
				if strings.ToUpper(fields[0]) == ".FIX" {
					if len(fields) > 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if len(fields) == 2 {
						if len(fields[1]) != 1 {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						switch (fields[1])[0] {
						case '1':
							digits = 1
						case '2':
							digits = 2
						case '3':
							digits = 3
						case '4':
							digits = 4
						case '5':
							digits = 5
						case '6':
							digits = 6
						default:
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
					} else {
						digits = 6
					}
					decfmt = "FIX" //формат с фиксированной точкой
				}
				if strings.ToUpper(fields[0]) == ".SCI" {
					if len(fields) > 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if len(fields) == 2 {
						if len(fields[1]) != 1 {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						switch (fields[1])[0] {
						case '1':
							digits = 1
						case '2':
							digits = 2
						case '3':
							digits = 3
						case '4':
							digits = 4
						case '5':
							digits = 5
						case '6':
							digits = 6
						default:
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
					} else {
						digits = 4
					}
					decfmt = "SCI" //научный формат
				}
				if strings.ToUpper(fields[0]) == ".END" {
					if len(fields) != 1 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					brkFlag = true //подъем флага окончания анализа
				}
				//недопустимая директива
				if (strings.ToUpper(fields[0]) != ".DC") && (strings.ToUpper(fields[0]) != ".AC") && (strings.ToUpper(fields[0]) != ".END") && (strings.ToUpper(fields[0]) != ".PARAM") && (strings.ToUpper(fields[0]) != ".DEG") && (strings.ToUpper(fields[0]) != ".RAD") && (strings.ToUpper(fields[0]) != ".FIX") && (strings.ToUpper(fields[0]) != ".SCI") {
					return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err16") + "</font>"
				}
			}
		}
		if brkFlag {
			break //прекращение анализа при поднятом флаге
		}
	}
	if dbg {
		fmt.Println(mode)
		fmt.Println(params)
	}
	//режим симуляции не задан
	if mode == "" {
		return "<font color=\"red\">" + msg(lang, "err02") + "</font>"
	}
	//анализ компонентов
	brkFlag = false //сброс флага прерывания анализа
	for k := 0; k < len(lines); k++ {
		fields := strings.Fields(lines[k]) //раскладываем строку на отдельные поля
		if len(fields) != 0 {              //если непустая строка
			//подстановка параметров
			for j := 0; j < len(fields); j++ {
				stringTmp = string((fields[j])[0]) + string((fields[j])[len(fields[j])-1]) //комбинация первого и последнего символа
				if stringTmp == "{}" {                                                     //если найден параметр
					if len(fields[j]) < 3 { //имя параметра не задано
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err12") + "</font>"
					}
					stringTmp = strings.ToUpper((fields[j])[1 : len(fields[j])-1]) //определение имени параметра
					stringTmp, success := params[stringTmp]
					if success {
						fields[j] = stringTmp //замена имени параметра на значение в списке соединений
					} else {

					}
				}
			}
			ch := strings.ToUpper(fields[0])[0] //первый символ первого поля
			switch ch {
			case '.':
				{
					if strings.ToUpper(fields[0]) == ".END" { //окончание списка соединений
						if len(fields) != 1 {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						brkFlag = true //подъем флага прерывания анализа
					}
				}
			case 'Z': //комплексное сопротивление
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if ((len(fields) != 5) && (len(fields) != 4)) && (mode == "AC") {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if mode == "DC" {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err17") + "</font>"
					}
					nz++ //инкремент количества комплексных сопротивлений
					zElement.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					zElement.from = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							zElement.from = index
						}
					}
					if zElement.from == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						zElement.from = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					zElement.to = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							zElement.to = index
						}
					}
					if zElement.to == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						zElement.to = n
					}
					absValue, success = libs.Str2Val(fields[3])
					if success != true {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					argValue = 0.0
					if len(fields) == 5 {
						if (fields[4])[len(fields[4])-1] == 'i' {
							if len(fields[4]) == 1 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							} else {
								//алгебраическая форма
								if absValue < 0 {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err19") + "</font>"
								}
								xValue := absValue
								yValue, success := libs.Str2Val((fields[4])[0 : len(fields[4])-1])
								if success != true {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
								}
								absValue = math.Sqrt(xValue*xValue + yValue*yValue)
								if absValue == 0 {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err08") + "</font>"
								}
								if xValue != 0 {
									if xValue > 0 {
										argValue = math.Atan(yValue / xValue)
									} else {
										argValue = math.Atan(yValue/xValue) + math.Pi
									}
								} else {
									if yValue > 0 {
										argValue = math.Pi / 2.0
									} else {
										argValue = -math.Pi / 2.0
									}
								}
								zElement.cValue = cmplx.Rect(absValue, argValue)
							}
						} else {
							//экспоненциальная форма
							if absValue <= 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err08") + "</font>"
							}
							argValue, success = libs.Str2Angle(fields[4])
							if success != true {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							zElement.cValue = cmplx.Rect(absValue, argValue)
						}
					} else {
						if absValue <= 0 {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err08") + "</font>"
						}
						zElement.cValue = cmplx.Rect(absValue, argValue)
					}
					//проверка на дублирование имени компонента
					for i := range z_elements {
						if z_elements[i].name == zElement.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					z_elements = append(z_elements, zElement)
				}
			case 'R': //резистор
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if len(fields) != 4 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					nr++ //инкремент количества резисторов
					element.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.from = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.from = index
						}
					}
					if element.from == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.from = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.to = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.to = index
						}
					}
					if element.to == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.to = n
					}
					element.fValue, success = libs.Str2Val(fields[3])
					if success != true {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if element.fValue <= 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err08") + "</font>"
					}
					element.cValue = 0 + 0i
					//проверка на дублирование имени компонента
					for i := range r_elements {
						if r_elements[i].name == element.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					r_elements = append(r_elements, element)
				}
			case 'T': //длинная линия
				{
					if len(fields[0]) < 3 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if (strings.ToUpper(fields[0])[1] == 'Z') && (mode == "DC") {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err17") + "</font>"
					}
					//узлы четырехполюсника
					if (strings.ToUpper(fields[0])[1] == 'R') || (strings.ToUpper(fields[0])[1] == 'Z') {
						//имя
						tName := strings.ToUpper(fields[0])
						//узел 1
						int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
						if err != nil {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						intTmp = int(int64Tmp)
						if intTmp < 0 {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
						}
						tNode1 := -1
						for index := range nodes {
							if nodes[index].Name == intTmp {
								nodes[index].Count++
								tNode1 = index
							}
						}
						if tNode1 == -1 {
							//добавление нового узла в таблицу
							node.Name = intTmp //имя
							node.Count = 1     //счетчик
							nodes = append(nodes, node)
							n++ //увеличение на 1 числа узлов
							tNode1 = n
						}
						//узел 2
						int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
						if err != nil {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						intTmp = int(int64Tmp)
						if intTmp < 0 {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
						}
						tNode2 := -1
						for index := range nodes {
							if nodes[index].Name == intTmp {
								nodes[index].Count++
								tNode2 = index
							}
						}
						if tNode2 == -1 {
							//добавление нового узла в таблицу
							node.Name = intTmp //имя
							node.Count = 1     //счетчик
							nodes = append(nodes, node)
							n++ //увеличение на 1 числа узлов
							tNode2 = n
						}
						//узел 3
						int64Tmp, err = strconv.ParseInt(fields[3], 10, 0)
						if err != nil {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						intTmp = int(int64Tmp)
						if intTmp < 0 {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
						}
						tNode3 := -1
						for index := range nodes {
							if nodes[index].Name == intTmp {
								nodes[index].Count++
								tNode3 = index
							}
						}
						if tNode3 == -1 {
							//добавление нового узла в таблицу
							node.Name = intTmp //имя
							node.Count = 1     //счетчик
							nodes = append(nodes, node)
							n++ //увеличение на 1 числа узлов
							tNode3 = n
						}
						switch strings.ToUpper(fields[0])[1] {
						case 'R':
							{
								if len(fields) != 7 {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
								}
								tR, success := libs.Str2Val(fields[4])
								if success != true {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
								}
								if tR <= 0 {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err20") + "</font>"
								}
								tA, success := libs.Str2Val(fields[5])
								if success != true {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
								}
								if tA <= 0 {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err21") + "</font>"
								}
								tLen, success := libs.Str2Val(fields[6])
								if success != true {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
								}
								if tLen <= 0 {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err22") + "</font>"
								}
								//расчет сопротивлений tR1 tR2 tR3 FIXME
								tR1 := tA * tLen * tR * math.Sinh(tA*tLen) / (tA * tLen)
								tR2 := tR * tA * tLen * math.Sinh(tA*tLen) / (tA * tLen * (math.Cosh(tA*tLen) - 1))
								tR3 := tR2 //симметричная линия
								//задание элементов
								//element.name element.from element.to element.fValue
								//R1
								element.name = "R1." + tName
								element.cValue = 0 + 0i
								element.fValue = tR1
								element.from = tNode1
								element.to = tNode2
								//проверка на дублирование имени компонента
								for i := range r_elements {
									if r_elements[i].name == element.name {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
									}
								}
								r_elements = append(r_elements, element)
								nr++
								//R2
								element.name = "R2." + tName
								element.cValue = 0 + 0i
								element.fValue = tR2
								element.from = tNode1
								element.to = tNode3
								//проверка на дублирование имени компонента
								for i := range r_elements {
									if r_elements[i].name == element.name {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
									}
								}
								r_elements = append(r_elements, element)
								nr++
								//R3
								element.name = "R3." + tName
								element.cValue = 0 + 0i
								element.fValue = tR3
								element.from = tNode2
								element.to = tNode3
								//проверка на дублирование имени компонента
								for i := range r_elements {
									if r_elements[i].name == element.name {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
									}
								}
								r_elements = append(r_elements, element)
								nr++
							}
						case 'Z':
							{
								if len(fields) != 9 {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
								}
								//волновое сопротивление
								absValue, success = libs.Str2Val(fields[4])
								if success != true {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
								}
								argValue = 0.0
								if (fields[5])[len(fields[5])-1] == 'i' {
									if len(fields[5]) == 1 {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
									} else {
										//алгебраическая форма
										if absValue < 0 {
											return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err19") + "</font>"
										}
										xValue := absValue
										yValue, success := libs.Str2Val((fields[5])[0 : len(fields[5])-1])
										if success != true {
											return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
										}
										absValue = math.Sqrt(xValue*xValue + yValue*yValue)
										if absValue == 0 {
											return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err08") + "</font>"
										}
										if xValue != 0 {
											if xValue > 0 {
												argValue = math.Atan(yValue / xValue)
											} else {
												argValue = math.Atan(yValue/xValue) + math.Pi
											}
										} else {
											if yValue > 0 {
												argValue = math.Pi / 2.0
											} else {
												argValue = -math.Pi / 2.0
											}
										}
									}
								} else {
									//экспоненциальная форма
									if absValue <= 0 {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err08") + "</font>"
									}
									argValue, success = libs.Str2Angle(fields[5])
									if success != true {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
									}
								}
								tZ := cmplx.Rect(absValue, argValue)
								//постоянная передачи
								absValue, success = libs.Str2Val(fields[6])
								if success != true {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
								}
								argValue = 0.0
								if (fields[7])[len(fields[7])-1] == 'i' {
									if len(fields[7]) == 1 {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
									} else {
										//алгебраическая форма
										if absValue < 0 {
											return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err19") + "</font>"
										}
										xValue := absValue
										yValue, success := libs.Str2Val((fields[7])[0 : len(fields[7])-1])
										if success != true {
											return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
										}
										absValue = math.Sqrt(xValue*xValue + yValue*yValue)
										if absValue == 0 {
											return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err08") + "</font>"
										}
										if xValue != 0 {
											if xValue > 0 {
												argValue = math.Atan(yValue / xValue)
											} else {
												argValue = math.Atan(yValue/xValue) + math.Pi
											}
										} else {
											if yValue > 0 {
												argValue = math.Pi / 2.0
											} else {
												argValue = -math.Pi / 2.0
											}
										}
									}
								} else {
									//экспоненциальная форма
									if absValue <= 0 {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err08") + "</font>"
									}
									argValue, success = libs.Str2Angle(fields[7])
									if success != true {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
									}
								}
								tG := cmplx.Rect(absValue, argValue)
								//длина
								tLen, success := libs.Str2Val(fields[8])
								if success != true {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
								}
								if tLen <= 0 {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err22") + "</font>"
								}
								//расчет сопротивлений FIXME
								tZ1 := tG * cmplx.Rect(tLen, 0) * tZ * cmplx.Sinh(tG*cmplx.Rect(tLen, 0)) / (tG * cmplx.Rect(tLen, 0))
								tZ2 := tZ * tG * cmplx.Rect(tLen, 0) * cmplx.Sinh(tG*cmplx.Rect(tLen, 0)) / (tG * cmplx.Rect(tLen, 0) * (cmplx.Cosh(tG*cmplx.Rect(tLen, 0)) - 1))
								tZ3 := tZ2 //симметричная линия
								//задание элементов
								//Z1
								zElement.name = "Z1." + tName
								zElement.cValue = tZ1
								zElement.from = tNode1
								zElement.to = tNode2
								//проверка на дублирование имени компонента
								for i := range z_elements {
									if z_elements[i].name == zElement.name {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
									}
								}
								z_elements = append(z_elements, zElement)
								nz++
								//Z2
								zElement.name = "Z2." + tName
								zElement.cValue = tZ2
								zElement.from = tNode1
								zElement.to = tNode3
								//проверка на дублирование имени компонента
								for i := range z_elements {
									if z_elements[i].name == zElement.name {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
									}
								}
								z_elements = append(z_elements, zElement)
								nz++
								//Z3
								zElement.name = "Z3." + tName
								zElement.cValue = tZ3
								zElement.from = tNode2
								zElement.to = tNode3
								//проверка на дублирование имени компонента
								for i := range z_elements {
									if z_elements[i].name == zElement.name {
										return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
									}
								}
								z_elements = append(z_elements, zElement)
								nz++
							}
						}
					} else {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
				}
			case 'L': //катушка
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if len(fields) != 4 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					nl++ //инкремент количества катушек индуктивности
					element.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.from = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.from = index
						}
					}
					if element.from == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.from = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.to = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.to = index
						}
					}
					if element.to == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.to = n
					}
					element.fValue, success = libs.Str2Val(fields[3])
					if success != true {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if element.fValue <= 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err10") + "</font>"
					}
					element.cValue = 0 + 0i
					///проверка на дублирование имени компонента
					for i := range l_elements {
						if l_elements[i].name == element.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					l_elements = append(l_elements, element)
				}
			case 'C': //конденсатор
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if len(fields) != 4 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					nc++ //инкремент количества конденсаторов
					element.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.from = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.from = index
						}
					}
					if element.from == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.from = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.to = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.to = index
						}
					}
					if element.to == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.to = n
					}
					element.fValue, success = libs.Str2Val(fields[3])
					if success != true {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if element.fValue <= 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err11") + "</font>"
					}
					element.cValue = 0 + 0i
					//проверка на дублирование имени компонента
					for i := range c_elements {
						if c_elements[i].name == element.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					c_elements = append(c_elements, element)
				}
			case 'V': //независимый источник ЭДС
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if ((len(fields) != 4) && (mode == "DC")) || (((len(fields) != 5) && (len(fields) != 4)) && (mode == "AC")) {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					nv++ //инкремент количества независимых источников ЭДС
					element.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.from = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.from = index
						}
					}
					if element.from == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.from = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.to = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.to = index
						}
					}
					if element.to == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.to = n
					}
					if mode == "DC" {
						element.fValue, success = libs.Str2Val(fields[3])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						element.cValue = 0 + 0i
					}
					if mode == "AC" {
						absValue, success = libs.Str2Val(fields[3])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						argValue = 0.0
						if len(fields) == 5 {
							argValue, success = libs.Str2Angle(fields[4])
							if success != true {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
						}
						element.cValue = cmplx.Rect(absValue, argValue)
						element.fValue = 0
					}
					//проверка на дублирование имени компонента
					for i := range v_elements {
						if v_elements[i].name == element.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					v_elements = append(v_elements, element)
				}
			case 'I': //независимый источник тока
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if ((len(fields) != 4) && (mode == "DC")) || (((len(fields) != 5) && (len(fields) != 4)) && (mode == "AC")) {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					ni++ //инкремент количества независимых источников тока
					element.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.from = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.from = index
						}
					}
					if element.from == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.from = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					element.to = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							element.to = index
						}
					}
					if element.to == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						element.to = n
					}
					if mode == "DC" {
						element.fValue, success = libs.Str2Val(fields[3])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						element.cValue = 0 + 0i
					}
					if mode == "AC" {
						absValue, success = libs.Str2Val(fields[3])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						argValue = 0.0
						if len(fields) == 5 {
							argValue, success = libs.Str2Angle(fields[4])
							if success != true {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
						}
						element.cValue = cmplx.Rect(absValue, argValue)
						element.fValue = 0
					}
					//проверка на дублирование имени компонента
					for i := range i_elements {
						if i_elements[i].name == element.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					i_elements = append(i_elements, element)
				}
			case 'K': //индуктивная связь
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if len(fields) != 4 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					nk++ //инкремент количества независимых источников тока
					kElement.name = strings.ToUpper(fields[0])
					if strings.ToUpper(fields[1])[0] != 'L' {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					kElement.l1 = strings.ToUpper(fields[1])
					if strings.ToUpper(fields[2])[0] != 'L' {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					kElement.l2 = strings.ToUpper(fields[2])
					kElement.k, success = libs.Str2Val(fields[3])
					if success != true {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err14") + "</font>"
					}
					if (kElement.k < 0.0) || (kElement.k > 1.0) {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err14") + "</font>"
					}
					//проверка на дублирование имени компонента
					for i := range k_elements {
						if k_elements[i].name == kElement.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					k_elements = append(k_elements, kElement)
				}
			case 'P':
				{
					if len(fields[0]) < 3 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					switch strings.ToUpper(fields[0])[1] {
					case 'A':
						{
							if len(fields) != 3 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							npa++
							paElement.name = strings.ToUpper(fields[0])
							int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							paElement.plus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									paElement.plus = index
								}
							}
							if paElement.plus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								paElement.plus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							paElement.minus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									paElement.minus = index
								}
							}
							if paElement.minus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								paElement.minus = n
							}
							//проверка на дублирование имени компонента
							for i := range pa_elements {
								if pa_elements[i].name == paElement.name {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
								}
							}
							pa_elements = append(pa_elements, paElement)
						}
					case 'V':
						{
							if len(fields) != 3 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							npv++
							pvElement.name = strings.ToUpper(fields[0])
							int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pvElement.plus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pvElement.plus = index
								}
							}
							if pvElement.plus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pvElement.plus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pvElement.minus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pvElement.minus = index
								}
							}
							if pvElement.minus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pvElement.minus = n
							}
							//проверка на дублирование имени компонента
							for i := range pv_elements {
								if pv_elements[i].name == pvElement.name {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
								}
							}
							pv_elements = append(pv_elements, pvElement)
						}
					case 'W':
						{
							if len(fields) != 5 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							npw++
							pwElement.name = strings.ToUpper(fields[0])
							int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pwElement.iplus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pwElement.iplus = index
								}
							}
							if pwElement.iplus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pwElement.iplus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pwElement.iminus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pwElement.iminus = index
								}
							}
							if pwElement.iminus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pwElement.iminus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[3], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pwElement.uplus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pwElement.uplus = index
								}
							}
							if pwElement.uplus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pwElement.uplus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[4], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pwElement.uminus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pwElement.uminus = index
								}
							}
							if pwElement.uminus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pwElement.uminus = n
							}
							//проверка на дублирование имени компонента
							for i := range pw_elements {
								if pw_elements[i].name == pwElement.name {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
								}
							}
							pw_elements = append(pw_elements, pwElement)
						}
					case 'Q':
						{
							if len(fields) != 5 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							npq++
							pqElement.name = strings.ToUpper(fields[0])
							int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pqElement.iplus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pqElement.iplus = index
								}
							}
							if pqElement.iplus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pqElement.iplus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pqElement.iminus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pqElement.iminus = index
								}
							}
							if pqElement.iminus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pqElement.iminus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[3], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pqElement.uplus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pqElement.uplus = index
								}
							}
							if pqElement.uplus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pqElement.uplus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[4], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pqElement.uminus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pqElement.uminus = index
								}
							}
							if pqElement.uminus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pqElement.uminus = n
							}
							//проверка на дублирование имени компонента
							for i := range pq_elements {
								if pq_elements[i].name == pqElement.name {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
								}
							}
							pq_elements = append(pq_elements, pqElement)
						}
					case 'F':
						{
							if len(fields) != 5 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							npf++
							pfElement.name = strings.ToUpper(fields[0])
							int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pfElement.iplus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pfElement.iplus = index
								}
							}
							if pfElement.iplus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pfElement.iplus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pfElement.iminus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pfElement.iminus = index
								}
							}
							if pfElement.iminus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pfElement.iminus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[3], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pfElement.uplus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pfElement.uplus = index
								}
							}
							if pfElement.uplus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pfElement.uplus = n
							}
							int64Tmp, err = strconv.ParseInt(fields[4], 10, 0)
							if err != nil {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
							intTmp = int(int64Tmp)
							if intTmp < 0 {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
							}
							pfElement.uminus = -1
							for index := range nodes {
								if nodes[index].Name == intTmp {
									nodes[index].Count++
									pfElement.uminus = index
								}
							}
							if pfElement.uminus == -1 {
								//добавление нового узла в таблицу
								node.Name = intTmp //имя
								node.Count = 1     //счетчик
								nodes = append(nodes, node)
								n++ //увеличение на 1 числа узлов
								pfElement.uminus = n
							}
							//проверка на дублирование имени компонента
							for i := range pf_elements {
								if pf_elements[i].name == pfElement.name {
									return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
								}
							}
							pf_elements = append(pf_elements, pfElement)
						}
					}
				}
			case 'E': //источник напряжения, управляемый напряжением (ИНУН)
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if ((len(fields) != 6) && (mode == "DC")) || ((len(fields) != 7) && (len(fields) != 6) && (mode == "AC")) {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					ne++
					sElement.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n1 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n1 = index
						}
					}
					if sElement.n1 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n1 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n2 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n2 = index
						}
					}
					if sElement.n2 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n2 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[3], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n3 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n3 = index
						}
					}
					if sElement.n3 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n3 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[4], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n4 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n4 = index
						}
					}
					if sElement.n4 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n4 = n
					}
					if mode == "DC" {
						sElement.fGain, success = libs.Str2Val(fields[5])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						sElement.cGain = 0 + 0i
					}
					if mode == "AC" {
						absValue, success = libs.Str2Val(fields[5])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						argValue = 0.0
						if len(fields) == 7 {
							argValue, success = libs.Str2Angle(fields[6])
							if success != true {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
						}
						sElement.cGain = cmplx.Rect(absValue, argValue)
						sElement.fGain = 0
					}
					//проверка на дублирование имени компонента
					for i := range e_elements {
						if e_elements[i].name == sElement.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					e_elements = append(e_elements, sElement)
				}
			case 'F': //источник тока, управляемый током (ИТУТ)
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if ((len(fields) != 6) && (mode == "DC")) || ((len(fields) != 7) && (len(fields) != 6) && (mode == "AC")) {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					nf++
					sElement.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n1 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n1 = index
						}
					}
					if sElement.n1 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n1 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n2 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n2 = index
						}
					}
					if sElement.n2 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n2 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[3], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n3 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n3 = index
						}
					}
					if sElement.n3 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n3 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[4], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n4 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n4 = index
						}
					}
					if sElement.n4 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n4 = n
					}
					if mode == "DC" {
						sElement.fGain, success = libs.Str2Val(fields[5])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						sElement.cGain = 0 + 0i
					}
					if mode == "AC" {
						absValue, success = libs.Str2Val(fields[5])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						argValue = 0.0
						if len(fields) == 7 {
							argValue, success = libs.Str2Angle(fields[6])
							if success != true {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
						}
						sElement.cGain = cmplx.Rect(absValue, argValue)
						sElement.fGain = 0
					}
					//проверка на дублирование имени компонента
					for i := range f_elements {
						if f_elements[i].name == sElement.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					f_elements = append(f_elements, sElement)
				}
			case 'G': //источник тока, управляемый напряжением (ИТУН)
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if ((len(fields) != 6) && (mode == "DC")) || ((len(fields) != 7) && (len(fields) != 6) && (mode == "AC")) {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					ng++
					sElement.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n1 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n1 = index
						}
					}
					if sElement.n1 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n1 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n2 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n2 = index
						}
					}
					if sElement.n2 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n2 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[3], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n3 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n3 = index
						}
					}
					if sElement.n3 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n3 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[4], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n4 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n4 = index
						}
					}
					if sElement.n4 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n4 = n
					}
					if mode == "DC" {
						sElement.fGain, success = libs.Str2Val(fields[5])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						sElement.cGain = 0 + 0i
					}
					if mode == "AC" {
						absValue, success = libs.Str2Val(fields[5])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						argValue = 0.0
						if len(fields) == 7 {
							argValue, success = libs.Str2Angle(fields[6])
							if success != true {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
						}
						sElement.cGain = cmplx.Rect(absValue, argValue)
						sElement.fGain = 0
					}
					//проверка на дублирование имени компонента
					for i := range g_elements {
						if g_elements[i].name == sElement.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					g_elements = append(g_elements, sElement)
				}
			case 'H': //источник напряжения, управляемый током (ИНУТ)
				{
					if len(fields[0]) < 2 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					if ((len(fields) != 6) && (mode == "DC")) || ((len(fields) != 7) && (len(fields) != 6) && (mode == "AC")) {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					nh++
					sElement.name = strings.ToUpper(fields[0])
					int64Tmp, err = strconv.ParseInt(fields[1], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n1 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n1 = index
						}
					}
					if sElement.n1 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n1 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[2], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n2 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n2 = index
						}
					}
					if sElement.n2 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n2 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[3], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n3 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n3 = index
						}
					}
					if sElement.n3 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n3 = n
					}
					int64Tmp, err = strconv.ParseInt(fields[4], 10, 0)
					if err != nil {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
					}
					intTmp = int(int64Tmp)
					if intTmp < 0 {
						return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err04") + "</font>"
					}
					sElement.n4 = -1
					for index := range nodes {
						if nodes[index].Name == intTmp {
							nodes[index].Count++
							sElement.n4 = index
						}
					}
					if sElement.n4 == -1 {
						//добавление нового узла в таблицу
						node.Name = intTmp //имя
						node.Count = 1     //счетчик
						nodes = append(nodes, node)
						n++ //увеличение на 1 числа узлов
						sElement.n4 = n
					}
					if mode == "DC" {
						sElement.fGain, success = libs.Str2Val(fields[5])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						sElement.cGain = 0 + 0i
					}
					if mode == "AC" {
						absValue, success = libs.Str2Val(fields[5])
						if success != true {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
						}
						argValue = 0.0
						if len(fields) == 7 {
							argValue, success = libs.Str2Angle(fields[6])
							if success != true {
								return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
							}
						}
						sElement.cGain = cmplx.Rect(absValue, argValue)
						sElement.fGain = 0
					}
					//проверка на дублирование имени компонента
					for i := range h_elements {
						if h_elements[i].name == sElement.name {
							return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + ":<br/>" + msg(lang, "err13") + "</font>"
						}
					}
					h_elements = append(h_elements, sElement)
				}
			case '*': //комментарий
				//comment
			default:
				return "<font color=\"red\">" + msg(lang, "err01") + " " + strconv.Itoa(k+1) + "</font>"
			}
		}
		if brkFlag { //окончание анализа компонентов
			break
		}
	}
	//проверка наличия компонентов
	if (nr == 0) && (nv == 0) && (ni == 0) && (nc == 0) && (nl == 0) && ((mode == "AC") && (nz == 0)) {
		return "<font color=\"red\">" + msg(lang, "err03") + "</font>"
	}
	//проверка наличия земли
	if nodes[0].Count == 0 {
		return "<font color=\"red\">" + msg(lang, "err05") + "</font>"
	}
	//проверка на плавающие узлы
	for index := range nodes {
		if nodes[index].Count < 2 {
			return "<font color=\"red\">" + strconv.Itoa(nodes[index].Name) + msg(lang, "err06") + "</font>"
		}
	}

	//проверка элементов индуктивных связей
	for i := range k_elements {
		boolTmp = false
		for j := range l_elements {
			if k_elements[i].l1 == l_elements[j].name {
				boolTmp = true
				k_elements[i].l1b = j
			}
		}
		if !boolTmp {
			return "<font color=\"red\">" + msg(lang, "err18") + "</font>"
		}
		boolTmp = false
		for j := range l_elements {
			if k_elements[i].l2 == l_elements[j].name {
				boolTmp = true
				k_elements[i].l2b = j
			}
		}
		if !boolTmp {
			return "<font color=\"red\">" + msg(lang, "err18") + "</font>"
		}
	}

	//проверка задания частоты
	if mode == "AC" {
		if (math.IsNaN(w)) && ((nl > 0) || (nc > 0)) {
			return "<font color=\"red\">" + msg(lang, "error") + ":<br/>" + msg(lang, "err15") + "</font>"
		}
	}

	//режим DC
	if mode == "DC" {
		if dbg {
			fmt.Println(strconv.Itoa(n) + " node(s)")
			fmt.Println("Nodes:")
			fmt.Println(nodes)
			fmt.Println("DC")
			fmt.Println("Components:")
			fmt.Println("R:")
			fmt.Println(r_elements)
			fmt.Println("L:")
			fmt.Println(l_elements)
			fmt.Println("C:")
			fmt.Println(c_elements)
			fmt.Println("V:")
			fmt.Println(v_elements)
			fmt.Println("I:")
			fmt.Println(i_elements)
			fmt.Println("PA:")
			fmt.Println(pa_elements)
			fmt.Println("PV:")
			fmt.Println(pv_elements)
			fmt.Println("PW:")
			fmt.Println(pw_elements)
			fmt.Println("PQ:")
			fmt.Println(pq_elements)
			fmt.Println("PF:")
			fmt.Println(pf_elements)
			fmt.Println("E:")
			fmt.Println(e_elements)
			fmt.Println("F:")
			fmt.Println(f_elements)
			fmt.Println("G:")
			fmt.Println(g_elements)
			fmt.Println("H:")
			fmt.Println(h_elements)
		}
		//составление матрицы проводимостей ветвей и редуцированной матрицы инциденций
		g := mat64.NewDense(nr, nr, nil)
		ar := mat64.NewDense(n, nr, nil)
		for k := 0; k < nr; k++ {
			g.Set(k, k, 1.0/r_elements[k].fValue)
			if r_elements[k].from > 0 {
				ar.Set(r_elements[k].from-1, k, 1.0)
			}
			if r_elements[k].to > 0 {
				ar.Set(r_elements[k].to-1, k, -1.0)
			}
		}
		//транспонирование матрицы инциденций
		art := ar.T()
		//вычисление матрицы узловых проводимостей
		gn := mat64.NewDense(n, n, nil)
		arg := mat64.NewDense(n, nr, nil)
		arg.Mul(ar, g)
		gn.Mul(arg, art)
		m = nv + nl + npa + npw + npq + npf + ne + nf + ng + nh*2 //кол-во эквивалентных источников ЭДС
		//составление матриц эквивалентных источников ЭДС
		ae := mat64.NewDense(n, m, nil)
		e := mat64.NewDense(m, 1, nil)
		//учет независимых источников ЭДС
		for k := 0; k < nv; k++ {
			e.Set(k, 0, v_elements[k].fValue)
			if v_elements[k].from > 0 {
				ae.Set(v_elements[k].from-1, k, 1.0)
			}
			if v_elements[k].to > 0 {
				ae.Set(v_elements[k].to-1, k, -1.0)
			}
		}
		//учет индуктивностей как нулевых источников ЭДС
		for k := 0; k < nl; k++ {
			e.Set(k+nv, 0, 0.0)
			if l_elements[k].from > 0 {
				ae.Set(l_elements[k].from-1, k+nv, 1.0)
			}
			if l_elements[k].to > 0 {
				ae.Set(l_elements[k].to-1, k+nv, -1.0)
			}
		}
		//учет амперметров как нулевых источников ЭДС
		for k := 0; k < npa; k++ {
			e.Set(k+nv+nl, 0, 0.0)
			if pa_elements[k].plus > 0 {
				ae.Set(pa_elements[k].plus-1, k+nv+nl, 1.0)
			}
			if pa_elements[k].minus > 0 {
				ae.Set(pa_elements[k].minus-1, k+nv+nl, -1.0)
			}
		}
		//учет ваттметров как нулевых источников ЭДС
		for k := 0; k < npw; k++ {
			e.Set(k+nv+nl+npa, 0, 0.0)
			if pw_elements[k].iplus > 0 {
				ae.Set(pw_elements[k].iplus-1, k+nv+nl+npa, 1.0)
			}
			if pw_elements[k].iminus > 0 {
				ae.Set(pw_elements[k].iminus-1, k+nv+nl+npa, -1.0)
			}
		}
		//учет фазометров как нулевых источников ЭДС
		for k := 0; k < npf; k++ {
			e.Set(k+nv+nl+npa+npw, 0, 0.0)
			if pf_elements[k].iplus > 0 {
				ae.Set(pf_elements[k].iplus-1, k+nv+nl+npa+npw, 1.0)
			}
			if pf_elements[k].iminus > 0 {
				ae.Set(pf_elements[k].iminus-1, k+nv+nl+npa+npw, -1.0)
			}
		}
		//учет варметров как нулевых источников ЭДС
		for k := 0; k < npq; k++ {
			e.Set(k+nv+nl+npa+npw+npf, 0, 0.0)
			if pq_elements[k].iplus > 0 {
				ae.Set(pq_elements[k].iplus-1, k+nv+nl+npa+npw+npf, 1.0)
			}
			if pq_elements[k].iminus > 0 {
				ae.Set(pq_elements[k].iminus-1, k+nv+nl+npa+npw+npf, -1.0)
			}
		}
		//
		//aet := mat64.ae.T()
		aet := mat64.NewDense(m, n, nil)
		//транспонирование матрицы
		for k := 0; k < m; k++ {
			for j := 0; j < n; j++ {
				aet.Set(k, j, ae.At(j, k))
			}
		}
		//
		d := mat64.NewDense(m, m, nil)
		//учет зависимых источников
		//VCVS
		for k := 0; k < ne; k++ {
			if e_elements[k].n1 > 0 {
				ae.Set(e_elements[k].n1-1, k+nv+nl+npa+npw+npf+npq, -1.0)
				aet.Set(k+nv+nl+npa+npw+npf+npq, e_elements[k].n1-1, -1.0)
			}
			if e_elements[k].n2 > 0 {
				ae.Set(e_elements[k].n2-1, k+nv+nl+npa+npw+npf+npq, 1.0)
				aet.Set(k+nv+nl+npa+npw+npf+npq, e_elements[k].n2-1, 1.0)
			}
			if e_elements[k].n3 > 0 {
				aet.Set(k+nv+nl+npa+npw+npf+npq, e_elements[k].n3-1, e_elements[k].fGain)
			}
			if e_elements[k].n4 > 0 {
				aet.Set(k+nv+nl+npa+npw+npf+npq, e_elements[k].n4-1, -e_elements[k].fGain)
			}
		}
		//CCCS
		for k := 0; k < nf; k++ {
			if f_elements[k].n1 > 0 {
				ae.Set(f_elements[k].n1-1, k+nv+nl+npa+npw+npf+npq+ne, 1.0)
			}
			if f_elements[k].n2 > 0 {
				ae.Set(f_elements[k].n2-1, k+nv+nl+npa+npw+npf+npq+ne, -1.0)
			}
			if f_elements[k].n3 > 0 {
				ae.Set(f_elements[k].n3-1, k+nv+nl+npa+npw+npf+npq+ne, 1.0/f_elements[k].fGain)
				aet.Set(k+nv+nl+npa+npw+npf+npq+ne, f_elements[k].n3-1, 1.0)
			}
			if f_elements[k].n4 > 0 {
				ae.Set(f_elements[k].n4-1, k+nv+nl+npa+npw+npf+npq+ne, -1.0/f_elements[k].fGain)
				aet.Set(k+nv+nl+npa+npw+npf+npq+ne, f_elements[k].n4-1, -1.0)
			}
		}
		//VCCS
		for k := 0; k < ng; k++ {
			if g_elements[k].n1 > 0 {
				ae.Set(g_elements[k].n1-1, k+nv+nl+npa+npw+npf+npq+ne+nf, 1.0)
			}
			if g_elements[k].n2 > 0 {
				ae.Set(g_elements[k].n2-1, k+nv+nl+npa+npw+npf+npq+ne+nf, -1.0)
			}
			if g_elements[k].n3 > 0 {
				aet.Set(k+nv+nl+npa+npw+npf+npq+ne+nf, g_elements[k].n3-1, 1.0)
			}
			if g_elements[k].n4 > 0 {
				aet.Set(k+nv+nl+npa+npw+npf+npq+ne+nf, g_elements[k].n4-1, -1.0)
			}
			d.Set(k+nv+nl+npa+npw+npf+npq+ne+nf, k+nv+nl+npa+npw+npf+npq+ne+nf, -1.0/g_elements[k].fGain)
		}
		//CCVS
		for k := 0; k < nh; k++ {
			if h_elements[k].n1 > 0 {
				ae.Set(h_elements[k].n1-1, k*2+1+nv+nl+npa+npw+npf+npq+ne+nf+ng, -1.0)
				aet.Set(k*2+nv+nl+npa+npw+npf+npq+ne+nf+ng, h_elements[k].n1-1, 1.0)
			}
			if h_elements[k].n2 > 0 {
				ae.Set(h_elements[k].n2-1, k*2+1+nv+nl+npa+npw+npf+npq+ne+nf+ng, 1.0)
				aet.Set(k*2+nv+nl+npa+npw+npf+npq+ne+nf+ng, h_elements[k].n2-1, -1.0)
			}
			if h_elements[k].n3 > 0 {
				ae.Set(h_elements[k].n3-1, k*2+nv+nl+npa+npw+npf+npq+ne+nf+ng, 1.0)
				aet.Set(k*2+1+nv+nl+npa+npw+npf+npq+ne+nf+ng, h_elements[k].n3-1, 1.0)
			}
			if h_elements[k].n4 > 0 {
				ae.Set(h_elements[k].n4-1, k*2+nv+nl+npa+npw+npf+npq+ne+nf+ng, -1.0)
				aet.Set(k*2+1+nv+nl+npa+npw+npf+npq+ne+nf+ng, h_elements[k].n4-1, -1.0)
			}
			d.Set(k*2+nv+nl+npa+npw+npf+npq+ne+nf+ng, k*2+nv+nl+npa+npw+npf+npq+ne+nf+ng, -h_elements[k].fGain)
		}
		//составление матриц эквивалентных источников тока
		ai := mat64.NewDense(n, ni+nc+npv+npw+npf+npq, nil)
		i := mat64.NewDense(ni+nc+npv+npw+npf+npq, 1, nil)
		//учет независимых источников тока
		for k := 0; k < ni; k++ {
			i.Set(k, 0, -i_elements[k].fValue)
			if i_elements[k].from > 0 {
				ai.Set(i_elements[k].from-1, k, 1.0)
			}
			if i_elements[k].to > 0 {
				ai.Set(i_elements[k].to-1, k, -1.0)
			}
		}
		//учет конденсаторов как нулевых источников тока
		for k := 0; k < nc; k++ {
			i.Set(k+ni, 0, 0.0)
			if c_elements[k].from > 0 {
				ai.Set(c_elements[k].from-1, k+ni, 1.0)
			}
			if c_elements[k].to > 0 {
				ai.Set(c_elements[k].to-1, k+ni, -1.0)
			}
		}
		//учет вольтметров как нулевых источников тока
		for k := 0; k < npv; k++ {
			i.Set(k+ni+nc, 0, 0.0)
			if pv_elements[k].plus > 0 {
				ai.Set(pv_elements[k].plus-1, k+ni+nc, 1.0)
			}
			if pv_elements[k].minus > 0 {
				ai.Set(pv_elements[k].minus-1, k+ni+nc, -1.0)
			}
		}
		//учет ваттметров как нулевых источников тока
		for k := 0; k < npw; k++ {
			i.Set(k+ni+nc+npv, 0, 0.0)
			if pw_elements[k].uplus > 0 {
				ai.Set(pw_elements[k].uplus-1, k+ni+nc+npv, 1.0)
			}
			if pw_elements[k].uminus > 0 {
				ai.Set(pw_elements[k].uminus-1, k+ni+nc+npv, -1.0)
			}
		}
		//учет фазометров как нулевых источников тока
		for k := 0; k < npf; k++ {
			i.Set(k+ni+nc+npv+npw, 0, 0.0)
			if pf_elements[k].uplus > 0 {
				ai.Set(pf_elements[k].uplus-1, k+ni+nc+npv+npw, 1.0)
			}
			if pf_elements[k].uminus > 0 {
				ai.Set(pf_elements[k].uminus-1, k+ni+nc+npv+npw, -1.0)
			}
		}
		//учет варметров как нулевых источников тока
		for k := 0; k < npq; k++ {
			i.Set(k+ni+nc+npv+npw+npf, 0, 0.0)
			if pq_elements[k].uplus > 0 {
				ai.Set(pq_elements[k].uplus-1, k+ni+nc+npv+npw+npf, 1.0)
			}
			if pq_elements[k].uminus > 0 {
				ai.Set(pq_elements[k].uminus-1, k+ni+nc+npv+npw+npf, -1.0)
			}
		}
		//вычисление матрицы узловых токов
		in := mat64.NewDense(n, 1, nil)
		in.Mul(ai, i)

		//
		a := mat64.NewDense(n+m, n+m, nil)
		for k := 0; k < n; k++ {
			for l := 0; l < n; l++ {
				a.Set(k, l, gn.At(k, l))
			}
		}
		for k := 0; k < n; k++ {
			for l := 0; l < m; l++ {
				a.Set(k, l+n, ae.At(k, l))
			}
		}
		for k := 0; k < m; k++ {
			for l := 0; l < n; l++ {
				a.Set(k+n, l, aet.At(k, l))
			}
		}
		for k := 0; k < m; k++ {
			for l := 0; l < m; l++ {
				a.Set(k+n, l+n, d.At(k, l))
			}
		}
		//
		z := mat64.NewDense(n+m, 1, nil)
		for k := 0; k < n; k++ {
			z.Set(k, 0, in.At(k, 0))
		}
		for k := 0; k < m; k++ {
			z.Set(k+n, 0, e.At(k, 0))
		}

		if dbg {
			fmt.Printf("GN :\n%v\n", mat64.Formatted(gn, mat64.Prefix(""), mat64.Excerpt(0)))
			fmt.Printf("A :\n%v\n", mat64.Formatted(a, mat64.Prefix(""), mat64.Excerpt(0)))
			fmt.Printf("B :\n%v\n", mat64.Formatted(z, mat64.Prefix(""), mat64.Excerpt(0)))
		}
		//
		if mat64.Det(a) == 0 {
			return "<font color=\"red\">" + msg(lang, "err07") + "</font>"

		}
		//вычисление обратной матрицы
		arev := mat64.NewDense(n+m, n+m, nil)
		arev.Inverse(a)
		//решение узлового уравнения
		x := mat64.NewDense(n+m, 1, nil)
		x.Mul(arev, z)
		if dbg {
			fmt.Printf("X :\n%v\n", mat64.Formatted(x, mat64.Prefix(""), mat64.Excerpt(0)))
		}
		//выделение матрицы узловых потенциалов
		fn := mat64.NewDense(n, 1, nil)
		for k := 0; k < n; k++ {
			fn.Set(k, 0, x.At(k, 0))
		}
		//выделение матрицы токов эквивалентных источников ЭДС
		ie := mat64.NewDense(m, 1, nil)
		for k := 0; k < m; k++ {
			ie.Set(k, 0, x.At(k+n, 0))
		}
		//вычисление падений напряжения на резисторах
		ur := mat64.NewDense(nr, 1, nil)
		ur.Mul(ar.T(), fn)
		//вычисления токов в резисторах
		ir := mat64.NewDense(nr, 1, nil)
		ir.Mul(g, ur)
		if dbg {
			fmt.Printf("UR :\n%v\n", mat64.Formatted(ur, mat64.Prefix(""), mat64.Excerpt(0)))
			fmt.Printf("IR :\n%v\n", mat64.Formatted(ir, mat64.Prefix(""), mat64.Excerpt(0)))
		}
		//вычисление напряжений на источниках тока
		uj := mat64.NewDense(ni+nc+npv+npw+npf+npq, 1, nil)
		uj.Mul(ai.T(), fn)
		//вычисление ЭДС зависимых источников
		es := mat64.NewDense(ne+nh, 1, nil)
		for k := 0; k < ne; k++ {
			es.Set(k, 0, 0)
			if e_elements[k].n1 > 0 {
				es.Set(k, 0, fn.At(e_elements[k].n1-1, 0))
			}
			if e_elements[k].n2 > 0 {
				es.Set(k, 0, es.At(k, 0)-fn.At(e_elements[k].n2-1, 0))
			}
		}
		for k := 0; k < nh; k++ {
			es.Set(k+ne, 0, 0)
			if h_elements[k].n1 > 0 {
				es.Set(k+ne, 0, fn.At(h_elements[k].n1-1, 0))
			}
			if h_elements[k].n2 > 0 {
				es.Set(k+ne, 0, es.At(k+ne, 0)-fn.At(h_elements[k].n2-1, 0))
			}
		}
		//вычисление напряжений зависимых источников
		us := mat64.NewDense(nf+ng, 1, nil)
		for k := 0; k < nf; k++ {
			us.Set(k, 0, 0)
			if f_elements[k].n1 > 0 {
				us.Set(k, 0, -fn.At(f_elements[k].n1-1, 0))
			}
			if f_elements[k].n2 > 0 {
				us.Set(k, 0, us.At(k, 0)+fn.At(f_elements[k].n2-1, 0))
			}
		}
		for k := 0; k < ng; k++ {
			us.Set(k+nf, 0, 0)
			if g_elements[k].n1 > 0 {
				us.Set(k+nf, 0, -fn.At(g_elements[k].n1-1, 0))
			}
			if g_elements[k].n2 > 0 {
				us.Set(k+nf, 0, us.At(k+nf, 0)+fn.At(g_elements[k].n2-1, 0))
			}
		}
		//вывод результатов
		res := msg(lang, "mode") + " " + "<font color=\"blue\">DC</font><br/>"
		res = res + msg(lang, "results") + "<br/><table border=\"2\" bordercolor=\"black\"  rules=\"rows\">"
		//амперметры
		for k := 0; k < npa; k++ {
			res = res + "<tr><td><font color=\"blue\">" + pa_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Float2Str(ie.At(k+nv+nl, 0), decfmt, digits) + " <b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;</td><td></td><td  align=\"right\">"
			res = res + "&nbsp;&nbsp;</td></tr>"
		}
		//вольтметры
		for k := 0; k < npv; k++ {
			res = res + "<tr><td><font color=\"blue\">" + pv_elements[k].name + "</font>:</td><td></td><td  align=\"right\">" + libs.Float2Str(uj.At(k+ni+nc, 0), decfmt, digits) + " <b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + "&nbsp;&nbsp;</td></tr>"
		}
		//ваттметры
		for k := 0; k < npw; k++ {
			res = res + "<tr><td><font color=\"blue\">" + pw_elements[k].name + "</font>:</td><td  align=\"right\">&nbsp;&nbsp;</td><td></td><td  align=\"right\">"
			res = res + libs.Float2Str(-ie.At(k+nv+nl+npa, 0)*(-uj.At(k+ni+nc+npv, 0)), decfmt, digits) + " <b>" + msg(lang, "watt") + "</b></td></tr>"
		}
		//фазометры и варметры игнорируются
		//резисторы
		for k := 0; k < nr; k++ {
			res = res + "<tr><td><font color=\"blue\">" + r_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Float2Str(ir.At(k, 0), decfmt, digits) + " <b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + libs.Float2Str(ur.At(k, 0), decfmt, digits) + " <b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + libs.Float2Str(ur.At(k, 0)*ir.At(k, 0), decfmt, digits) + " <b>" + msg(lang, "watt") + "</b></td></tr>"
		}
		//катушки
		for k := 0; k < nl; k++ {
			res = res + "<tr><td><font color=\"blue\">" + l_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Float2Str(ie.At(k+nv, 0), decfmt, digits) + " <b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;</td><td></td><td  align=\"right\">"
			res = res + "&nbsp;&nbsp;</td></tr>"
		}
		//конденсаторы
		for k := 0; k < nc; k++ {
			res = res + "<tr><td><font color=\"blue\">" + c_elements[k].name + "</font>:</td><td></td><td  align=\"right\">" + libs.Float2Str(uj.At(k+ni, 0), decfmt, digits) + " <b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + "&nbsp;&nbsp;</td></tr>"
		}
		//источники ЭДС
		for k := 0; k < nv; k++ {
			res = res + "<tr><td><font color=\"blue\">" + v_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Float2Str(-ie.At(k, 0), decfmt, digits) + " <b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;</td><td></td><td  align=\"right\">"
			res = res + libs.Float2Str(-ie.At(k, 0)*e.At(k, 0), decfmt, digits) + " <b>" + msg(lang, "watt") + "</b></td></tr>"
		}
		//источники тока
		for k := 0; k < ni; k++ {
			res = res + "<tr><td><font color=\"blue\">" + i_elements[k].name + "</font>:</td><td></td><td  align=\"right\">" + libs.Float2Str(-uj.At(k, 0), decfmt, digits) + " <b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + libs.Float2Str(uj.At(k, 0)*i.At(k, 0), decfmt, digits) + " <b>" + msg(lang, "watt") + "</b></td></tr>"
		}
		//VCVS E
		for k := 0; k < ne; k++ {
			res = res + "<tr><td><font color=\"blue\">" + e_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Float2Str(ie.At(k+nv+nl, 0), decfmt, digits) + " <b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + libs.Float2Str(es.At(k, 0), decfmt, digits) + " <b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;</td>"
			res = res + "<td  align=\"right\">" + libs.Float2Str(es.At(k, 0)*ie.At(k+nv+nl, 0), decfmt, digits) + " <b>" + msg(lang, "watt") + "</b></td></tr>"
		}
		//CCCS F
		for k := 0; k < nf; k++ {
			res = res + "<tr><td><font color=\"blue\">" + f_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Float2Str(-ie.At(k+nv+nl+ne, 0), decfmt, digits) + " <b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + libs.Float2Str(us.At(k, 0), decfmt, digits) + " <b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;</td>"
			res = res + "<td  align=\"right\">" + libs.Float2Str(-us.At(k, 0)*ie.At(k+nv+nl+ne, 0), decfmt, digits) + " <b>" + msg(lang, "watt") + "</b></td></tr>"
		}
		//VCCS G
		for k := 0; k < ng; k++ {
			res = res + "<tr><td><font color=\"blue\">" + g_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Float2Str(-ie.At(k+nv+nl+ne+nf, 0), decfmt, digits) + " <b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + libs.Float2Str(us.At(k+nf, 0), decfmt, digits) + " <b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;</td>"
			res = res + "<td  align=\"right\">" + libs.Float2Str(-us.At(k+nf, 0)*ie.At(k+nv+nl+ne+nf, 0), decfmt, digits) + " <b>" + msg(lang, "watt") + "</b></td></tr>"
		}
		//CCVS H
		for k := 0; k < nh; k++ {
			res = res + "<tr><td><font color=\"blue\">" + h_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Float2Str(ie.At(k*2+1+nv+nl+ne+nf+ng, 0), decfmt, digits) + " <b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<td  align=\"right\">"
			res = res + libs.Float2Str(es.At(k+ne, 0), decfmt, digits) + " <b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;</td>"
			res = res + "<td  align=\"right\">" + libs.Float2Str(es.At(k+ne, 0)*ie.At(k*2+1+nv+nl+ne+nf+ng, 0), decfmt, digits) + " <b>" + msg(lang, "watt") + "</b></td></tr></table>"
		}
		return res
	}
	//расчет цепи синусоидального тока
	if mode == "AC" {
		if dbg {
			fmt.Println("AC")
			fmt.Println("Frequency = " + libs.Float2Str(w/2.0/math.Pi, decfmt, digits) + " Hz")
		}
		//круговая частота
		//заполнение матрицы проводимостей ветвей и матрицы инциденций
		y = libs.Crt2DComplex(nr+nl+nc+nz, nr+nl+nc+nz)
		az = libs.Crt2DFloat(n, nr+nl+nc+nz)
		if nk == 0 {
			//учет резисторов
			for k := 0; k < nr; k++ {
				y[k][k] = (1 + 0i) / cmplx.Rect(r_elements[k].fValue, 0)
				if r_elements[k].from > 0 {
					az[r_elements[k].from-1][k] = 1.0
				}
				if r_elements[k].to > 0 {
					az[r_elements[k].to-1][k] = -1.0
				}
			}
			//учет катушек индуктивности
			for k := 0; k < nl; k++ {
				y[k+nr][k+nr] = cmplx.Rect(1.0/(l_elements[k].fValue*w), -math.Pi/2.0)
				if l_elements[k].from > 0 {
					az[l_elements[k].from-1][k+nr] = 1.0
				}
				if l_elements[k].to > 0 {
					az[l_elements[k].to-1][k+nr] = -1.0
				}
			}
			//учет конденсаторов
			for k := 0; k < nc; k++ {
				y[k+nr+nl][k+nr+nl] = cmplx.Rect(c_elements[k].fValue*w, math.Pi/2.0)
				if c_elements[k].from > 0 {
					az[c_elements[k].from-1][k+nr+nl] = 1.0
				}
				if c_elements[k].to > 0 {
					az[c_elements[k].to-1][k+nr+nl] = -1.0
				}
			}
			//учет комплексных сопротивлений
			for k := 0; k < nz; k++ {
				y[k+nr+nl+nc][k+nr+nl+nc] = (1 + 0i) / z_elements[k].cValue
				if z_elements[k].from > 0 {
					az[z_elements[k].from-1][k+nr+nl+nc] = 1.0
				}
				if z_elements[k].to > 0 {
					az[z_elements[k].to-1][k+nr+nl+nc] = -1.0
				}
			}
		} else {
			//при наличии индуктивных связей
			//составление матрицы сопротивлений ветвей
			zb = libs.Crt2DComplex(nr+nl+nc+nz, nr+nl+nc)
			//учет резисторов
			for k := 0; k < nr; k++ {
				zb[k][k] = cmplx.Rect(r_elements[k].fValue, 0)
				if r_elements[k].from > 0 {
					az[r_elements[k].from-1][k] = 1.0
				}
				if r_elements[k].to > 0 {
					az[r_elements[k].to-1][k] = -1.0
				}
			}
			//учет катушек индуктивности
			for k := 0; k < nl; k++ {
				zb[k+nr][k+nr] = cmplx.Rect(l_elements[k].fValue*w, math.Pi/2.0)
				if l_elements[k].from > 0 {
					az[l_elements[k].from-1][k+nr] = 1.0
				}
				if l_elements[k].to > 0 {
					az[l_elements[k].to-1][k+nr] = -1.0
				}
			}
			//учет конденсаторов
			for k := 0; k < nc; k++ {
				zb[k+nr+nl][k+nr+nl] = cmplx.Rect(1.0/(c_elements[k].fValue*w), -math.Pi/2.0)
				if c_elements[k].from > 0 {
					az[c_elements[k].from-1][k+nr+nl] = 1.0
				}
				if c_elements[k].to > 0 {
					az[c_elements[k].to-1][k+nr+nl] = -1.0
				}
			}
			//учет комплексных сопротивлений
			for k := 0; k < nz; k++ {
				zb[k+nr+nl+nc][k+nr+nl+nc] = z_elements[k].cValue
				if z_elements[k].from > 0 {
					az[z_elements[k].from-1][k+nr+nl+nc] = 1.0
				}
				if z_elements[k].to > 0 {
					az[z_elements[k].to-1][k+nr+nl+nc] = -1.0
				}
			}
			//учет индуктивных связей
			for k := 0; k < nk; k++ {
				floatTmp = w * k_elements[k].k * math.Sqrt(l_elements[k_elements[k].l1b].fValue*l_elements[k_elements[k].l2b].fValue)
				fmt.Println("Zm")
				fmt.Println(floatTmp)
				zb[k_elements[k].l1b+nr][k_elements[k].l2b+nr] = cmplx.Rect(floatTmp, math.Pi/2.0)
				zb[k_elements[k].l2b+nr][k_elements[k].l1b+nr] = cmplx.Rect(floatTmp, math.Pi/2.0)
			}
			if dbg {
				fmt.Println("Z")
				fmt.Println(zb)
			}
			//обращение матрицы сопротивлений
			tmp1DMatrix = libs.Crt1DComplex(nr + nl + nc + nz)
			for k := 0; k < (nr + nl + nc + nz); k++ {
				//заполнение матрицы свободных членов
				for j := 0; j < (nr + nl + nc + nz); j++ {
					if j == k {
						tmp1DMatrix[j] = 1.0 + 0.0i
					} else {
						tmp1DMatrix[j] = 0.0 + 0.0i
					}
				}
				tmp1DMatrix, _ = libs.GaussComplex(zb, tmp1DMatrix)
				//получение k-го столбца обратной матрицы
				for j := 0; j < (nr + nl + nc + nz); j++ {
					y[j][k] = tmp1DMatrix[j]
				}
			}
		}
		if dbg {
			fmt.Println("Y")
			fmt.Println(y)
		}
		//транспонирование матрицы инциденций
		azt := libs.T2DFloat(az)

		azg, _ := libs.Mul2DFloat2DComplex(az, y)
		yn, _ := libs.Mul2DComplex2DFloat(azg, azt)

		//заполнение матриц источиков
		m = nv + npa + npw + npf + npq + ne + nf + ng + nh*2 //количество источников ЭДС
		ae := libs.Crt2DComplex(n, m)
		e := libs.Crt1DComplex(m)
		//учет независимых источников ЭДС
		for k := 0; k < nv; k++ {
			e[k] = v_elements[k].cValue
			if v_elements[k].from > 0 {
				ae[v_elements[k].from-1][k] = 1 + 0i
			}
			if v_elements[k].to > 0 {
				ae[v_elements[k].to-1][k] = -1 + 0i
			}
		}
		//учет амперметров
		for k := 0; k < npa; k++ {
			e[k+nv] = 0.0 + 0.0i
			if pa_elements[k].plus > 0 {
				ae[pa_elements[k].plus-1][k+nv] = 1 + 0i
			}
			if pa_elements[k].minus > 0 {
				ae[pa_elements[k].minus-1][k+nv] = -1 + 0i
			}
		}
		//учет ваттметров
		for k := 0; k < npw; k++ {
			e[k+nv+npa] = 0.0 + 0.0i
			if pw_elements[k].iplus > 0 {
				ae[pw_elements[k].iplus-1][k+nv+npa] = 1 + 0i
			}
			if pw_elements[k].iminus > 0 {
				ae[pw_elements[k].iminus-1][k+nv+npa] = -1 + 0i
			}
		}
		//учет фазометров
		for k := 0; k < npf; k++ {
			e[k+nv+npa+npw] = 0.0 + 0.0i
			if pf_elements[k].iplus > 0 {
				ae[pf_elements[k].iplus-1][k+nv+npa+npw] = 1 + 0i
			}
			if pf_elements[k].iminus > 0 {
				ae[pf_elements[k].iminus-1][k+nv+npa+npw] = -1 + 0i
			}
		}
		//учет варметров
		for k := 0; k < npq; k++ {
			e[k+nv+npa+npw+npf] = 0.0 + 0.0i
			if pq_elements[k].iplus > 0 {
				ae[pq_elements[k].iplus-1][k+nv+npa+npw+npf] = 1 + 0i
			}
			if pq_elements[k].iminus > 0 {
				ae[pq_elements[k].iminus-1][k+nv+npa+npw+npf] = -1 + 0i
			}
		}
		//транспонирование
		aet := libs.T2DComplex(ae)

		d := libs.Crt2DComplex(m, m)
		//учет зависимых источников
		//VCVS
		for k := 0; k < ne; k++ {
			if e_elements[k].n1 > 0 {
				ae[e_elements[k].n1-1][k+nv+npa+npw+npf+npq] = -1 + 0i
				aet[k+nv+npa+npw+npf+npq][e_elements[k].n1-1] = -1 + 0i
			}
			if e_elements[k].n2 > 0 {
				ae[e_elements[k].n2-1][k+nv+npa+npw+npf+npq] = 1 + 0i
				aet[k+nv+npa+npw+npf+npq][e_elements[k].n2-1] = 1 + 0i
			}
			if e_elements[k].n3 > 0 {
				aet[k+nv+npa+npw+npf+npq][e_elements[k].n3-1] = e_elements[k].cGain
			}
			if e_elements[k].n4 > 0 {
				aet[k+nv+npa+npw+npf+npq][e_elements[k].n4-1] = -e_elements[k].cGain
			}
		}
		//CCCS
		for k := 0; k < nf; k++ {
			if f_elements[k].n1 > 0 {
				ae[f_elements[k].n1-1][k+nv+npa+npw+npf+npq+ne] = 1 + 0i
			}
			if f_elements[k].n2 > 0 {
				ae[f_elements[k].n2-1][k+nv+npa+npw+npf+npq+ne] = -1 + 0i
			}
			if f_elements[k].n3 > 0 {
				ae[f_elements[k].n3-1][k+nv+npa+npw+npf+npq+ne] = 1.0 / f_elements[k].cGain
				aet[k+nv+npa+npw+npf+npq+ne][f_elements[k].n3-1] = 1 + 0i
			}
			if f_elements[k].n4 > 0 {
				ae[f_elements[k].n4-1][k+nv+npa+npw+npf+npq+ne] = -1.0 / f_elements[k].cGain
				aet[k+nv+npa+npw+npf+npq+ne][f_elements[k].n4-1] = -1 + 0i
			}
		}
		//VCCS
		for k := 0; k < ng; k++ {
			if g_elements[k].n1 > 0 {
				ae[g_elements[k].n1-1][k+nv+npa+npw+npf+npq+ne+nf] = 1 + 0i
			}
			if g_elements[k].n2 > 0 {
				ae[g_elements[k].n2-1][k+nv+npa+npw+npf+npq+ne+nf] = -1 + 0i
			}
			if g_elements[k].n3 > 0 {
				aet[k+nv+npa+npw+npf+npq+ne+nf][g_elements[k].n3-1] = 1 + 0i
			}
			if g_elements[k].n4 > 0 {
				aet[k+nv+npa+npw+npf+npq+ne+nf][g_elements[k].n4-1] = -1 + 0i
			}
			d[k+nv+npa+npw+npf+npq+ne+nf][k+nv+npa+npw+npf+npq+ne+nf] = -1.0 / g_elements[k].cGain
		}
		//CCVS
		for k := 0; k < nh; k++ {
			if h_elements[k].n1 > 0 {
				ae[h_elements[k].n1-1][k*2+1+nv+npa+npw+npf+npq+ne+nf+ng] = -1 + 0i
				aet[k*2+nv+npa+npw+npf+npq+ne+nf+ng][h_elements[k].n1-1] = 1 + 0i
			}
			if h_elements[k].n2 > 0 {
				ae[h_elements[k].n2-1][k*2+1+nv+npa+npw+npf+npq+ne+nf+ng] = 1 + 0i
				aet[k*2+nv+npa+npw+npf+npq+ne+nf+ng][h_elements[k].n2-1] = -1 + 0i
			}
			if h_elements[k].n3 > 0 {
				ae[h_elements[k].n3-1][k*2+nv+npa+npw+npf+npq+ne+nf+ng] = 1 + 0i
				aet[k*2+1+nv+npa+npw+npf+npq+ne+nf+ng][h_elements[k].n3-1] = 1 + 0i
			}
			if h_elements[k].n4 > 0 {
				ae[h_elements[k].n4-1][k*2+nv+npa+npw+npf+npq+ne+nf+ng] = -1 + 0i
				aet[k*2+1+nv+npa+npw+npf+npq+ne+nf+ng][h_elements[k].n4-1] = -1 + 0i
			}
			d[k*2+nv+npa+npw+npf+npq+ne+nf+ng][k*2+nv+npa+npw+npf+npq+ne+nf+ng] = -h_elements[k].cGain
		}

		if dbg {
			fmt.Println("E")
			fmt.Println(e)
		}
		//определение матрицы узловых токов
		in := libs.Crt1DComplex(n)
		ai := libs.Crt2DFloat(n, ni+npv+npw+npf+npq)
		i := libs.Crt1DComplex(ni + npv + npw + npf + npq)
		if ni+npv+npw+npf+npq > 0 {
			//учет источников тока
			for k := 0; k < ni; k++ {
				i[k] = -i_elements[k].cValue
				if i_elements[k].from > 0 {
					ai[i_elements[k].from-1][k] = 1.0
				}
				if i_elements[k].to > 0 {
					ai[i_elements[k].to-1][k] = -1.0
				}
			}
			//учет вольтметров
			for k := 0; k < npv; k++ {
				i[k+ni] = 0.0 + 0.0i
				if pv_elements[k].plus > 0 {
					ai[pv_elements[k].plus-1][k+ni] = 1.0
				}
				if pv_elements[k].minus > 0 {
					ai[pv_elements[k].minus-1][k+ni] = -1.0
				}
			}
			//учет ваттметров
			for k := 0; k < npw; k++ {
				i[k+ni+npv] = 0.0 + 0.0i
				if pw_elements[k].uplus > 0 {
					ai[pw_elements[k].uplus-1][k+ni+npv] = 1.0
				}
				if pw_elements[k].uminus > 0 {
					ai[pw_elements[k].uminus-1][k+ni+npv] = -1.0
				}
			}
			//учет фазометров
			for k := 0; k < npf; k++ {
				i[k+ni+npv+npw] = 0.0 + 0.0i
				if pf_elements[k].uplus > 0 {
					ai[pf_elements[k].uplus-1][k+ni+npv+npw] = 1.0
				}
				if pf_elements[k].uminus > 0 {
					ai[pf_elements[k].uminus-1][k+ni+npv+npw] = -1.0
				}
			}
			in, _ = libs.Mul2DFloat1DComplex(ai, i)
		}
		//учет варметров
		for k := 0; k < npq; k++ {
			i[k+ni+npv+npw+npf] = 0.0 + 0.0i
			if pq_elements[k].uplus > 0 {
				ai[pq_elements[k].uplus-1][k+ni+npv+npw+npf] = 1.0
			}
			if pq_elements[k].uminus > 0 {
				ai[pq_elements[k].uminus-1][k+ni+npv+npw+npf] = -1.0
			}
		}
		//
		a := libs.Crt2DComplex(n+m, n+m)
		for k := 0; k < n; k++ {
			for l := 0; l < n; l++ {
				a[k][l] = yn[k][l]
			}
		}
		for k := 0; k < n; k++ {
			for l := 0; l < m; l++ {
				a[k][l+n] = ae[k][l]
			}
		}
		for k := 0; k < m; k++ {
			for l := 0; l < n; l++ {
				a[k+n][l] = aet[k][l]
			}
		}
		for k := 0; k < m; k++ {
			for l := 0; l < m; l++ {
				a[k+n][l+n] = d[k][l]
			}
		}
		//
		z := libs.Crt1DComplex(n + m)
		for k := 0; k < n; k++ {
			z[k] = in[k]
		}
		for k := 0; k < m; k++ {
			z[k+n] = e[k]
		}

		if dbg {
			fmt.Println("YN")
			fmt.Println(yn)
			fmt.Println("A")
			fmt.Println(a)
			fmt.Println("B")
			fmt.Println(z)
		}
		//решение системы уравнений методом Гаусса
		x, _ := libs.GaussComplex(a, z)
		if dbg {
			fmt.Println("X")
			fmt.Println(x)
		}
		//выделение матрицы узловых потенциалов
		fn := libs.Crt1DComplex(n)
		for k := 0; k < n; k++ {
			fn[k] = x[k]
		}
		//выделение матрицы токов источников ЭДС
		ie := libs.Crt1DComplex(m)
		for k := 0; k < m; k++ {
			ie[k] = x[k+n]
		}
		//напряжения на резисторах
		uz, _ := libs.Mul2DFloat1DComplex(azt, fn)
		//токи в резисторах
		iz, _ := libs.Mul2DComplex1DComplex(y, uz)
		if dbg {
			fmt.Println("UZ")
			fmt.Println(uz)
			fmt.Println("IZ")
			fmt.Println(iz)
		}
		//напряжения на источниках тока
		uj := libs.Crt1DComplex(ni + npv + npw + npf + npq)
		if ni+npv+npw+npf+npq > 0 {
			ait := libs.T2DFloat(ai)
			uj, _ = libs.Mul2DFloat1DComplex(ait, fn)
			if dbg {
				fmt.Println("UJ")
				fmt.Println(uj)
			}
		}
		//вычисление ЭДС зависимых источников
		es := libs.Crt1DComplex(ne + nh)
		for k := 0; k < ne; k++ {
			es[k] = 0 + 0i
			if e_elements[k].n1 > 0 {
				es[k] = fn[e_elements[k].n1-1]
			}
			if e_elements[k].n2 > 0 {
				es[k] = es[k] - fn[e_elements[k].n2-1]
			}
		}
		for k := 0; k < nh; k++ {
			es[k+ne] = 0 + 0i
			if h_elements[k].n1 > 0 {
				es[k+ne] = fn[h_elements[k].n1-1]
			}
			if h_elements[k].n2 > 0 {
				es[k+ne] = es[k+ne] - fn[h_elements[k].n2-1]
			}
		}
		//вычисление напряжений зависимых источников
		us := libs.Crt1DComplex(nf + ng)
		for k := 0; k < nf; k++ {
			us[k] = 0 + 0i
			if f_elements[k].n1 > 0 {
				us[k] = -fn[f_elements[k].n1-1]
			}
			if f_elements[k].n2 > 0 {
				us[k] = us[k] + fn[f_elements[k].n2-1]
			}
		}
		for k := 0; k < ng; k++ {
			us[k+nf] = 0 + 0i
			if g_elements[k].n1 > 0 {
				us[k+nf] = -fn[g_elements[k].n1-1]
			}
			if g_elements[k].n2 > 0 {
				us[k+nf] = us[k+nf] + fn[g_elements[k].n2-1]
			}
		}
		//вывод результатов
		res := msg(lang, "mode") + " " + "<font color=\"blue\">AC</font><br/>"
		res = res + msg(lang, "results") + "<br/><table border=\"2\" bordercolor=\"black\"  rules=\"rows\">"
		//амперметры
		for k := 0; k < npa; k++ {
			res = res + "<tr><td><font color=\"blue\">" + pa_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Float2Str(cmplx.Abs(ie[k+nv]), decfmt, digits) + "</td><td  valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">"
			res = res + "&nbsp;&nbsp;</td><td>&nbsp;&nbsp;</td></tr>"
		}
		//вольтметры
		for k := 0; k < npv; k++ {
			res = res + "<tr><td><font color=\"blue\">" + pv_elements[k].name + "</font>:</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">" + libs.Float2Str(cmplx.Abs(uj[k+ni]), decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + "&nbsp;&nbsp;</td><td>&nbsp;&nbsp;</td></tr>"
		}
		//ваттметры
		for k := 0; k < npw; k++ {
			res = res + "<tr><td><font color=\"blue\">" + pw_elements[k].name + "</font>:</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;&nbsp;</td><td>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + libs.Float2Str(cmplx.Abs(uj[k+ni+npv]*cmplx.Conj(ie[k+nv+npa]))*math.Cos(cmplx.Phase(uj[k+ni+npv]*cmplx.Conj(ie[k+nv+npa]))), decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "watt") + "</b></td></tr>"
		}
		//варметры
		for k := 0; k < npq; k++ {
			res = res + "<tr><td><font color=\"blue\">" + pq_elements[k].name + "</font>:</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;&nbsp;</td><td>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + libs.Float2Str(cmplx.Abs(uj[k+ni+npv+npw+npf]*cmplx.Conj(ie[k+nv+npa+npw+npf]))*math.Sin(cmplx.Phase(uj[k+ni+npv+npw+npf]*cmplx.Conj(ie[k+nv+npa+npw+npf]))), decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "var") + "</b></td></tr>"
		}
		//фазометры
		for k := 0; k < npf; k++ {
			res = res + "<tr><td><font color=\"blue\">" + pf_elements[k].name + "</font>:</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;&nbsp;</td><td>&nbsp;&nbsp;</td><td  align=\"right\">"
			res = res + libs.Float2Str((cmplx.Phase(uj[k+ni+npv+npw])-cmplx.Phase(ie[k+nv+npa+npw]))*180/math.Pi, decfmt, digits) + "</td><td  valign=\"top\"><b>" + msg(lang, "deg") + "</b></td></tr>"
		}
		//резисторы
		for k := 0; k < nr; k++ {
			res = res + "<tr><td><font color=\"blue\">" + r_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Complex2Str(iz[k], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(uz[k], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\" valign=\"top\">"
			res = res + libs.Float2Str(cmplx.Abs(uz[k]*cmplx.Conj(iz[k]))*math.Cos(cmplx.Phase(uz[k]*cmplx.Conj(iz[k]))), decfmt, digits) + "<br/>&nbsp;</td><td valign=\"top\"><b>" + msg(lang, "watt") + "</b><br/>&nbsp;</td></tr>"
		}
		//катушки индуктивности
		for k := 0; k < nl; k++ {
			res = res + "<tr><td><font color=\"blue\">" + l_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Complex2Str(iz[k+nr], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(uz[k+nr], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\" valign=\"top\">"
			res = res + libs.Float2Str(cmplx.Abs(uz[k+nr]*cmplx.Conj(iz[k+nr]))*math.Sin(cmplx.Phase(uz[k+nr]*cmplx.Conj(iz[k+nr]))), decfmt, digits) + "<br/>&nbsp;</td><td valign=\"top\"><b>" + msg(lang, "var") + "</b><br/>&nbsp;</td></tr>"
		}
		//конденсаторы
		for k := 0; k < nc; k++ {
			res = res + "<tr><td><font color=\"blue\">" + c_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Complex2Str(iz[k+nr+nl], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(uz[k+nr+nl], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\" valign=\"top\">"
			res = res + libs.Float2Str(cmplx.Abs(uz[k+nr+nl]*cmplx.Conj(iz[k+nr+nl]))*math.Sin(cmplx.Phase(uz[k+nr+nl]*cmplx.Conj(iz[k+nr+nl]))), decfmt, digits) + "<br/>&nbsp;</td><td valign=\"top\"><b>" + msg(lang, "var") + "</b><br/>&nbsp;</td></tr>"
		}
		//комплексные сопротивления
		for k := 0; k < nz; k++ {
			res = res + "<tr><td><font color=\"blue\">" + z_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Complex2Str(iz[k+nr+nl+nc], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(uz[k+nr+nl+nc], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(cmplx.Conj(iz[k+nr+nl+nc])*uz[k+nr+nl+nc], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "voltamper") + "</b><br/>&nbsp;</td></tr>"
		}
		//независимые источники ЭДС
		for k := 0; k < nv; k++ {
			res = res + "<tr><td><font color=\"blue\">" + v_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Complex2Str(-ie[k], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(cmplx.Conj(-ie[k])*e[k], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "voltamper") + "</b><br/>&nbsp;</td></tr>"
		}
		//независимые источники тока
		if ni > 0 {
			for k := 0; k < ni; k++ {
				res = res + "<tr><td><font color=\"blue\">" + i_elements[k].name + "</font>:</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">&nbsp;</td><td  align=\"right\">" + libs.Complex2Str(-uj[k], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
				res = res + libs.Complex2Str(uj[k]*cmplx.Conj(i[k]), angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "voltamper") + "</b><br/>&nbsp;</td></tr>"
			}
		}
		//VCVS E
		for k := 0; k < ne; k++ {
			res = res + "<tr><td><font color=\"blue\">" + e_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Complex2Str(ie[k+nv], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(es[k], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(cmplx.Conj(ie[k+nv])*es[k], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "voltamper") + "</b><br/>&nbsp;</td></tr>"
		}
		//CCCS F
		for k := 0; k < nf; k++ {
			res = res + "<tr><td><font color=\"blue\">" + f_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Complex2Str(-ie[k+nv+ne], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(us[k], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(cmplx.Conj(-ie[k+nv+ne])*us[k], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "voltamper") + "</b><br/>&nbsp;</td></tr>"
		}
		//VCCS G
		for k := 0; k < ng; k++ {
			res = res + "<tr><td><font color=\"blue\">" + g_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Complex2Str(-ie[k+nv+ne+nf], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(us[k+nf], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(cmplx.Conj(-ie[k+nv+ne+nf])*us[k+nf], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "voltamper") + "</b><br/>&nbsp;</td></tr>"
		}
		//CCVS H
		for k := 0; k < nh; k++ {
			res = res + "<tr><td><font color=\"blue\">" + h_elements[k].name + "</font>:</td><td  align=\"right\">" + libs.Complex2Str(ie[k*2+1+nv+ne+nf+ng], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "amper") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(es[k+ne], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "volt") + "</b>&nbsp;&nbsp;<br/>&nbsp;</td><td  align=\"right\">"
			res = res + libs.Complex2Str(cmplx.Conj(ie[k*2+1+nv+ne+nf+ng])*es[k+ne], angle, decfmt, digits) + "</td><td valign=\"top\"><b>" + msg(lang, "voltamper") + "</b><br/>&nbsp;</td></tr></table>"
		}
		return res
	}
	//
	return "<font color=\"red\">" + msg(lang, "err02") + "</font>"
}

//экранирование символов в списке соединений
func safe(text string) string {
	text = strings.Replace(text, ">", "&gt;", -1)
	text = strings.Replace(text, "<", "&lt;", -1)
	text = strings.Replace(text, "/", "&#47;", -1)
	text = strings.Replace(text, "\\", "&#92;", -1)
	text = strings.Replace(text, "\"", "&quot;", -1)
	text = strings.Replace(text, "'", "&#39;", -1)
	text = strings.Replace(text, "&", "&amp;", -1)
	return text
}

//переключение языка на русский
func ru(w http.ResponseWriter, r *http.Request) {
	var lang = "RU"
	expiration := time.Now().Add(365 * 24 * time.Hour)
	langPtrCookie := http.Cookie{Name: "lang", Value: lang, Expires: expiration}
	http.SetCookie(w, &langPtrCookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

//переключение языка на английский
func en(w http.ResponseWriter, r *http.Request) {
	var lang = "EN"
	expiration := time.Now().Add(365 * 24 * time.Hour)
	langPtrCookie := http.Cookie{Name: "lang", Value: lang, Expires: expiration}
	http.SetCookie(w, &langPtrCookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

//вывод справки
func hlp(w http.ResponseWriter, r *http.Request) {
	var lang = "EN"
	langCookie, err := r.Cookie("lang")
	if err == nil {
		lang = langCookie.Value
	}
	if lang == "EN" {
		t, err := template.ParseFiles("hlp_en.gtpl")
		if err != nil {
			log.Fatal("Template reading error")
		}
		t.Execute(w, nil)
	} else {
		t, err := template.ParseFiles("hlp_ru.gtpl")
		if err != nil {
			log.Fatal("Template reading error")
		}
		t.Execute(w, nil)
	}
}

//очистка списка соединений
func rst(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "foxysim", Value: "", Expires: expiration}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func home(w http.ResponseWriter, r *http.Request) {
	var ip string
	ip = ""
	ipSlice := []string{}
	ipSlice = append(ipSlice, r.Header.Get("X-Forwarded-For"))
	ipSlice = append(ipSlice, r.Header.Get("x-forwarded-for"))
	ipSlice = append(ipSlice, r.Header.Get("X-FORWARDED-FOR"))
	if len(ipSlice) > 0 {
		ip = ipSlice[0]
	}
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	security := libs.GetMD5Hash(ip + salt)
	var lang = "EN"
	langCookie, err := r.Cookie("lang")
	if err == nil {
		lang = langCookie.Value
	}
	expiration := time.Now().Add(365 * 24 * time.Hour)
	langPtrCookie := http.Cookie{Name: "lang", Value: lang, Expires: expiration}
	http.SetCookie(w, &langPtrCookie)
	cookie, err := r.Cookie("foxysim")
	netTmp := ""
	if err == nil {
		netTmp = cookie.Value
	}
	netTmp = safe(netTmp)
	vars := map[string]interface{}{
		"Net": template.HTML(strings.Replace(netTmp, "$", "\r\n", -1)),
		"Sec": template.HTML(security),
		"Ip":  template.HTML(ip),
	}
	if lang == "EN" {
		t, err := template.ParseFiles("sim_en.gtpl")
		if err != nil {
			log.Fatal("Template reading error")
		}
		t.Execute(w, vars)
	} else {
		t, err := template.ParseFiles("sim_ru.gtpl")
		if err != nil {
			log.Fatal("Template reading error")
		}
		t.Execute(w, vars)
	}
}

//симуляция
func sim(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		var ip string
		ip = ""
		ipSlice := []string{}
		ipSlice = append(ipSlice, r.Header.Get("X-Forwarded-For"))
		ipSlice = append(ipSlice, r.Header.Get("x-forwarded-for"))
		ipSlice = append(ipSlice, r.Header.Get("X-FORWARDED-FOR"))
		if len(ipSlice) > 0 {
			ip = ipSlice[0]
		}
		if ip == "" {
			ip = strings.Split(r.RemoteAddr, ":")[0]
		}
		security := libs.GetMD5Hash(ip + salt)
		r.ParseForm()
		if r.Form["security"][0] != security {
			http.Redirect(w, r, "/", http.StatusFound)
		}
		//установка языка
		var lang = "EN"
		langCookie, err := r.Cookie("lang")
		if err == nil {
			lang = langCookie.Value
		}
		netlist := r.Form["netlist"][0]
		netlist = safe(netlist)
		res := calc(netlist, lang)
		vars := map[string]interface{}{
			"Res": template.HTML(res),
		}
		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "foxysim", Value: strings.Replace(netlist, "\r\n", "$", -1), Expires: expiration}
		http.SetCookie(w, &cookie)
		if lang == "EN" {
			t, err := template.ParseFiles("ans_en.gtpl")
			if err != nil {
				log.Fatal("Template reading error")
			}
			t.Execute(w, vars)
		} else {
			t, err := template.ParseFiles("ans_ru.gtpl")
			if err != nil {
				log.Fatal("Template reading error")
			}
			t.Execute(w, vars)
		}
	}
}

func main() {
	dbg = true
	localesNum := 2
	localesCapacity := 30

	locales = make(map[string]map[string]string, localesNum)
	EN := make(map[string]string, localesCapacity)
	EN["error"] = "ERROR"
	EN["err00"] = "Empty netlist!"
	EN["err01"] = "ERROR at line"
	EN["err02"] = "Simulation mode is undefined!"
	EN["err03"] = "Components not found!"
	EN["err04"] = "only non-negative node numbers allowed"
	EN["err05"] = "Ground node is undefined!"
	EN["err06"] = "is floating node"
	EN["err07"] = "Singular matrix!"
	EN["err08"] = "only positive resistance values allowed"
	EN["err09"] = "AC mode is not yet implemented"
	EN["err10"] = "only positive inductance values allowed"
	EN["err11"] = "only positive capacitance values allowed"
	EN["err12"] = "parameters name is undefined"
	EN["err13"] = "duplicate component name"
	EN["err14"] = "invalid coefficient of coupling"
	EN["err14"] = "invalid coefficient of coupling"
	EN["err15"] = "frequency is undefined"
	EN["err16"] = "invalid directive"
	EN["err17"] = "Z components are allowed in AC mode only"
	EN["err18"] = "Error in setting the inductive coupling"
	EN["err19"] = "negative resistance"
	EN["err20"] = "invalid characteristic impedance"
	EN["err21"] = "invalid propagation constant"
	EN["err22"] = "invalid length"

	EN["mode"] = "Mode:"
	EN["results"] = "Results:"
	EN["amper"] = "A"
	EN["volt"] = "V"
	EN["watt"] = "W"
	EN["var"] = "var"
	EN["voltamper"] = "V·A"
	EN["deg"] = "&deg;"

	locales["EN"] = EN

	RU := make(map[string]string, localesCapacity)
	RU["error"] = "ОШИБКА"
	RU["err00"] = "Список соединений пуст!"
	RU["err01"] = "ОШИБКА в строке"
	RU["err02"] = "Режим симуляции не задан!"
	RU["err03"] = "Компоненты не заданы!"
	RU["err04"] = "номер узла некорректен"
	RU["err05"] = "Земля не подключена"
	RU["err06"] = " - плавающий узел"
	RU["err07"] = "Сингулярная матрица!"
	RU["err08"] = "допустимы только положительные значения сопротивлений"
	RU["err09"] = "AC режим еще не реализован"
	RU["err10"] = "допустимы только положительные значения индуктивностей"
	RU["err11"] = "допустимы только положительные значения емкостей"
	RU["err12"] = "имя параметра не задано"
	RU["err13"] = "дублирующееся имя компонента"
	RU["err14"] = "коэффициент связи некорректен"
	RU["err15"] = "частота не задана"
	RU["err16"] = "недопустимая директива"
	RU["err17"] = "Z компоненты разрешены только в AC режиме"
	RU["err18"] = "Ошибка в задании индуктивной связи"
	RU["err19"] = "отрицательное активное сопротивление"
	RU["err20"] = "недопустимое волновое сопротивление"
	RU["err21"] = "недопустимая постоянная передачи"
	RU["err22"] = "недопустимая длина"

	RU["mode"] = "Режим:"
	RU["results"] = "Результаты:"
	RU["amper"] = "А"
	RU["volt"] = "В"
	RU["watt"] = "Вт"
	RU["var"] = "вар"
	RU["voltamper"] = "В·А"
	RU["deg"] = "&deg;"

	locales["RU"] = RU

	fmt.Println("FoxySim - online circuit simulator")
	fmt.Println("(C) Alexey \"FoxyLab\" Voronin")
	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/", home)
	http.HandleFunc("/calc", sim)
	http.HandleFunc("/help", hlp)
	http.HandleFunc("/ru", ru)
	http.HandleFunc("/en", en)
	http.HandleFunc("/reset", rst)
	http.Handle("/static/", http.StripPrefix("/static", fs))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
