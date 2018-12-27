<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">
<html>
<head>
<link rel="stylesheet" type="text/css" href="static/style.css" />
<meta http-equiv="content-type" content="text/html; charset=utf-8">
<link rel="icon" type="image/png" href="http://foxylab.com:7777/static/favicon.png" />
<link rel="shortcut icon" type="image/png" href="http://foxylab.com:7777/static/favicon.png" />
<meta name="description" content="Circuit simulator FoxySim">
<meta name="keywords" content="Alexey V.Voronin, Алексей Воронин, Гомель, Беларусь, БелГУТ, Белорусский государственный университет транспорта, электротехника, ТОЭ, теоретические основы электротехники">
<meta name="robots" content="noarchive">
<meta name="author" content="Alexey V. Voronin @ FoxyLab">
<meta name="viewport" content="width=device-width">
<meta name="robots" content="noarchive">
<meta name="googleBOT" content="noarchive">
<title>FoxySim -  онлайн-симулятор ЭЦ</title>
</head>
<body>
<p>
<table>
<tr valign="middle">
<td><img src="static/favicon.png" alt="" /></td>
<td><b><i>FoxySim -  онлайн-симулятор ЭЦ</i></b></td>
</tr>
</table>
</p>
<p>
<b>Справка</b><br/>
Подробнее - <a href="https://acdc.foxylab.com/foxysim" target="_blank">https://acdc.foxylab.com/foxysim</a><br/><br/>
<b>Директивы</b>
<table>
<tr>
<td>расчет цепи постоянного тока
</td>
<td><b>.DC</b>
</td>
</tr>
<tr>
<td>расчет цепи синусоидального тока
</td>
<td><b>.AC</b> частота<br/>
лин. частота в Гц - по умолчанию или символ <b>f</b> в конце значения;<br/>
круг. частота в рад/с - символ <b>w</b> в конце значения;<br/>
частота может не задаваться, если в схеме отсутствуют компоненты L и C 
</td>
</tr>
<tr>
<td>задание параметра
</td>
<td><b>.PARAM</b> имя значение<br/>
В списке соединений параметр подставляется в виде <b>{</b>имя<b>}</b> 
</td>
</tr>
<tr>
<td>вывод фаз в градусах<br/>(по умолчанию)
</td>
<td><b>.DEG</b>
</td>
</tr>
<tr>
<td>вывод фаз в радианах
</td>
<td><b>.RAD</b>
</td>
</tr>
<tr>
<td>вывод значений с фиксированной точкой<br/>(по умолчанию)
</td>
<td><b>.FIX число_дес._знаков</b>
<br/>
если число знаков не указано, то выводится шесть десятичных знаков
</td>
</tr>
<tr>
<td>вывод значений в научном формате
</td>
<td><b>.SCI число_знач._цифр</b>
<br/>
если число значащих цифр после точки не указано, то после точки выводится четыре цифры
</td>
</tr>
<tr>
<td>окончание списка соединений
</td>
<td><b>.END</b>
</td>
</tr>
</table><br/><br/>
<b>Компоненты</b>
<table>
<tr>
<td>Источник ЭДС</td>
<td><img src="static/vs_din.png" alt="" /></td><td>DC:&nbsp;<b>V</b>имя&nbsp;N1&nbsp;N2&nbsp;значение<br/>AC:&nbsp;<b>V</b>имя&nbsp;N1&nbsp;N2&nbsp;действ._значение&nbsp;нач._фаза
<br/>
если значение  нач. фазы не задано, то оно принимается равным нулю;<br/>
нач. фаза в градусах - по умолчанию или символ <b>d</b> в конце значения;<br/>
нач. фаза в радианах - символ <b>r</b> в конце значения 
</td>
</tr>
<tr>
<td>Источник тока</td>
<td><img src="static/cs_din.png" alt="" /></td><td>DC:&nbsp;<b>I</b>имя&nbsp;N1&nbsp;N2&nbsp;значение<br/>AC:&nbsp;<b>I</b>имя&nbsp;N1&nbsp;N2&nbsp;действ._значение&nbsp;нач._фаза
<br/>
если значение нач. фазы не задано, то оно принимается равным нулю;<br/>
нач. фаза в градусах - по умолчанию или символ <b>d</b> в конце значения;<br/>
нач. фаза в радианах - символ <b>r</b> в конце значения 
</td>
</tr>
<td>ИНУН</td>
<td><img src="static/vcvs_din.png" alt="" /></td><td>DC:&nbsp;<b>E</b>имя&nbsp;N1&nbsp;N2&nbsp;коэф._передачи<br/>AC:&nbsp;<b>E</b>имя&nbsp;N1&nbsp;N2&nbsp;модуль_коэф._передачи&nbsp;фаза_коэф._передачи
<br/>
если значение фазы коэф. передачи не задано, то оно принимается равным нулю;<br/>
фаза коэф. передачи в градусах - по умолчанию или символ <b>d</b> в конце значения;<br/>
фаза коэф. передачи в радианах - символ <b>r</b> в конце значения 
</td>
</tr>
<tr>
<td>ИТУТ</td>
<td><img src="static/cccs_din.png" alt="" /></td><td>DC:&nbsp;<b>F</b>имя&nbsp;N1&nbsp;N2&nbsp;коэф._передачи<br/>AC:&nbsp;<b>F</b>имя&nbsp;N1&nbsp;N2&nbsp;модуль_коэф._передачи&nbsp;фаза_коэф._передачи
<br/>
если значение фазы коэф. передачи не задано, то оно принимается равным нулю;<br/>
фаза коэф. передачи в градусах - по умолчанию или символ <b>d</b> в конце значения;<br/>
фаза коэф. передачи в радианах - символ <b>r</b> в конце значения</td>
</tr>
<tr>
<td>ИТУН</td>
<td><img src="static/vccs_din.png" alt="" /></td><td>DC:&nbsp;<b>G</b>имя&nbsp;N1&nbsp;N2&nbsp;коэф._передачи<br/>AC:&nbsp;<b>G</b>имя&nbsp;N1&nbsp;N2&nbsp;модуль_коэф._передачи&nbsp;фаза_коэф._передачи
<br/>
если значение фазы коэф. передачи не задано, то оно принимается равным нулю;<br/>
фаза коэф. передачи в градусах - по умолчанию или символ <b>d</b> в конце значения;<br/>
фаза коэф. передачи в радианах - символ <b>r</b> в конце значения</td>
</tr>
<tr>
<td>ИНУТ</td>
<td><img src="static/ccvs_din.png" alt="" /></td><td>DC:&nbsp;<b>H</b>имя&nbsp;N1&nbsp;N2&nbsp;коэф._передачи<br/>AC:&nbsp;<b>H</b>имя&nbsp;N1&nbsp;N2&nbsp;модуль_коэф._передачи&nbsp;фаза_коэф._передачи
<br/>
если значение фазы коэф. передачи не задано, то оно принимается равным нулю;<br/>
фаза коэф. передачи в градусах - по умолчанию или символ <b>d</b> в конце значения;<br/>
фаза коэф. передачи в радианах - символ <b>r</b> в конце значения</td>
</tr>
<tr>
<td>Резистор</td>
<td><img src="static/res_din.png" alt="" /></td><td><b>R</b>имя&nbsp;N1&nbsp;N2&nbsp;значение</td>
</tr>
<tr>
<td>Катушка<br/>индуктивности</td>
<td><img src="static/ind_din.png" alt="" /></td><td><b>L</b>имя&nbsp;N1&nbsp;N2&nbsp;значение</td>
</tr>
<tr>
<td>Индуктивная<br/>связь</td>
<td><img src="static/coupling_din.png" alt="" /></td><td><b>K</b>имя&nbsp;<b>L</b>имя&nbsp;<b>L</b>имя&nbsp;коэф._связи</td>
</tr>
<tr>
<td>Конденсатор</td>
<td><img src="static/cap_din.png" alt="" /></td><td><b>C</b>имя&nbsp;N1&nbsp;N2&nbsp;значение</td>
</tr>
<tr>
<td>Комплексное<br/>сопротивление</td>
<td><img src="static/z_din.png" alt="" /></td><td>
в экспоненциальной форме:
<br/>
<b>Z</b>имя&nbsp;N1&nbsp;N2&nbsp;модуль&nbsp;фаза
<br/>
если значение фазы не задано, то оно принимается равным нулю;<br/>
фаза в градусах - по умолчанию или символ <b>d</b> в конце значения;<br/>
фаза в радианах - символ <b>r</b> в конце значения 
<br/>
<br/>
в алгебраической форме:
<br/>
<b>Z</b>имя&nbsp;N1&nbsp;N2&nbsp;актив._сопр.&nbsp;реактив._сопр.<b>i</b>
</td>
</tr>
<td>Длинная<br/>линия</td>
<td><img src="static/t.png" alt="" /></td><td></td>
</tr>
<tr>
<td>RG-линия:</td>
<td><img src="static/tr_din.png" alt="" /></td><td>
<b>TR</b>имя&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;характер._сопр.&nbsp;пост._передачи&nbsp;длина</td>
</tr>
<tr>
<td>RGLC-линия:</td>
<td><img src="static/tz_din.png" alt="" /></td><td>
<b>TZ</b>имя&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;характер._сопр.&nbsp;пост._передачи&nbsp;длина
<br/>
характеристическое сопротивление и постоянная передачи указываются в комплексной форме (алгебраической или экспоненциальной)
</td>
</tr>
<tr>
<td>Амперметр</td>
<td><img src="static/amp.png" alt="" /></td><td><b>PA</b>имя&nbsp;N1&nbsp;N2</td>
</tr>
<tr>
<td>Вольтметр</td>
<td><img src="static/volt.png" alt="" /></td><td><b>PV</b>имя&nbsp;N1&nbsp;N2</td>
</tr>
<tr>
<td>Ваттметр</td>
<td><img src="static/watt.png" alt="" /></td><td><b>PW</b>имя&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;N4</td>
</tr>
<tr>
<td>Варметр</td>
<td><img src="static/var.png" alt="" /></td><td><b>PQ</b>имя&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;N4</td>
</tr>
<tr>
<td>Фазометр</td>
<td><img src="static/ph.png" alt="" /></td><td><b>PF</b>имя&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;N4</td>
</tr>
</table>
<br/><br/>
<b>Комментарий:</b></br>
<b>*</b>комментарий</br><br/>
<i>Примеры описания схем</i><br/>
<table>
<tr>
<td>Цепь постоянного тока<br/><img src="static/cir_ex_din.png" alt="" /></td>
</tr>
<tr>
<td>
<pre>
.DC
V1 1 0 10
R1 1 2 5
R2 2 0 15
R3 2 3 20
V2 3 0 30
I1 2 0 5
.END
</pre>
</td>
</tr>
<tr>
<td>Цепь синусоидального тока<br/><img src="static/rlc_din.png" alt="" /></td>
</tr>
<tr>
<td>
<pre>
.AC 50
V1 1 0 100 0
PW1 1 2 1 0
PQ1 2 3 2 0
PF1 3 4 3 0
PA1 4 5
PV1 1 0
R1 5 6 50
L1 6 7 100m
C1 7 0 80u
.END
</pre>
</td>
</tr>
</table>
<table style="width:300px;">
<tr>
<td><a href="/">&lt;&lt;&lt; Назад</a>
</td>
</tr>
<tr>
<td  class="small">
<img src="static/foxylab_micro.png" alt="" />&nbsp;&copy; <i>Alexey &quot;FoxyLab&quot; Voronin</i><br />
<img src="static/golang_micro.png" alt="" />&nbsp;FoxySim powered by golang
</td>
</tr>
</table>
</p>
</body>
</html>