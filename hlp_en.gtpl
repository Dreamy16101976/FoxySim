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
<title>FoxySim - online circuit simulator</title>
</head>
<body>
<p>
<table>
<tr valign="middle">
<td><img src="static/favicon.png" alt="" /></td>
<td><b><i>FoxySim - online circuit simulator</i></b></td>
</tr>
</table>
</p>
<p>
<b>Help</b><br/><br/>
<b>Directives</b>
<table>
<tr>
<td>DC circuit analysis
</td>
<td><b>.DC</b>
</td>
</tr>
<tr>
<td>AC circuit analysis
</td>
<td><b>.AC</b> frequency
<br/>
regular frequency in Hz - by default or the symbol <b>f</b> at the end of the value;<br/>
angular frequency in rad/s - the symbol <b>w</b> at the end of the value;<br/>
the frequency may not be specified if the circuit do not exist components L and C
</td>
</tr>
<tr>
<td>define parameter
</td>
<td><b>.PARAM</b> name value<br/>
In the netlist, the parameter is substituted in the form <b>{</b>name<b>}</b> 
</td>
</tr>
<tr>
<td>printing phases in degrees<br/>(by default)
</td>
<td><b>.DEG</b>
</td>
</tr>
<tr>
<td>printing phases in radians
</td>
<td><b>.RAD</b>
</td>
</tr>
<tr>
<td>printing values in fixed-point notation<br/>(by default)
</td>
<td><b>.FIX number_of_decimal_places</b>
<br/>
if number of decimal places is not specified, displays six decimal places
</td>
</tr>
<tr>
<td>printing values in scientific notation
</td>
<td><b>.SCI number_of_significant_digits</b>
<br/>
if number of significant digits is not specified, displays four significant digits after decimal point
</td>
</tr>
<tr>
<td>end of netlist
</td>
<td><b>.END</b>
</td>
</tr>
</table><br/><br/>
<b>Components</b>
<table>
<tr>
<td>Voltage source</td>
<td><img src="static/vs_ansi.png" alt="" /></td><td>DC:&nbsp;<b>V</b>name&nbsp;N1&nbsp;N2&nbsp;value<br/>AC:&nbsp;<b>V</b>name&nbsp;N1&nbsp;N2&nbsp;RMS_value&nbsp;phase
<br/>
phase by default equal to zero;<br/>
phase in degrees - by default or the symbol <b>d</b> at the end of the value;<br/>
phase in radians - the symbol <b>r</b> at the end of the value
</td>
</tr>
<tr>
<td>Current source</td>
<td><img src="static/cs_ansi.png" alt="" /></td><td>DC:&nbsp;<b>I</b>name&nbsp;N1&nbsp;N2&nbsp;value<br/>AC:&nbsp;<b>I</b>name&nbsp;N1&nbsp;N2&nbsp;RMS_value&nbsp;phase
<br/>
phase by default equal to zero;<br/>
phase in degrees - by default or the symbol <b>d</b> at the end of the value;<br/>
phase in radians - the symbol <b>r</b> at the end of the value
</td>
</tr>
<td>VCVS</td>
<td><img src="static/vcvs_ansi.png" alt="" /></td><td>DC:&nbsp;<b>E</b>name&nbsp;N1&nbsp;N2&nbsp;gain<br/>AC:&nbsp;<b>E</b>name&nbsp;N1&nbsp;N2&nbsp;gain&nbsp;phase
<br/>
phase by default equal to zero;<br/>
phase in degrees - by default or the symbol <b>d</b> at the end of the value;<br/>
phase in radians - the symbol <b>r</b> at the end of the value
</td>
</tr>
<tr>
<td>CCCS</td>
<td><img src="static/cccs_ansi.png" alt="" /></td><td>DC:&nbsp;<b>F</b>name&nbsp;N1&nbsp;N2&nbsp;gain<br/>AC:&nbsp;<b>F</b>name&nbsp;N1&nbsp;N2&nbsp;gain&nbsp;phase
<br/>
phase by default equal to zero;<br/>
phase in degrees - by default or the symbol <b>d</b> at the end of the value;<br/>
phase in radians - the symbol <b>r</b> at the end of the value
</td>
</tr>
<tr>
<td>VCCS</td>
<td><img src="static/vccs_ansi.png" alt="" /></td><td>DC:&nbsp;<b>G</b>name&nbsp;N1&nbsp;N2&nbsp;gain<br/>AC:&nbsp;<b>G</b>name&nbsp;N1&nbsp;N2&nbsp;gain&nbsp;phase
<br/>
phase by default equal to zero;<br/>
phase in degrees - by default or the symbol <b>d</b> at the end of the value;<br/>
phase in radians - the symbol <b>r</b> at the end of the value
</td>
</tr>
<tr>
<td>CCVS</td>
<td><img src="static/ccvs_ansi.png" alt="" /></td><td>DC:&nbsp;<b>H</b>name&nbsp;N1&nbsp;N2&nbsp;gain<br/>AC:&nbsp;<b>H</b>name&nbsp;N1&nbsp;N2&nbsp;gain&nbsp;phase
<br/>
phase by default equal to zero;<br/>
phase in degrees - by default or the symbol <b>d</b> at the end of the value;<br/>
phase in radians - the symbol <b>r</b> at the end of the value
</td>
</tr>
<tr>
<td>Resistor</td>
<td><img src="static/res_ansi.png" alt="" /></td><td><b>R</b>name&nbsp;N1&nbsp;N2&nbsp;value</td>
</tr>
<tr>
<td>Inductance</td>
<td><img src="static/ind_ansi.png" alt="" /></td><td><b>L</b>name&nbsp;N1&nbsp;N2&nbsp;value</td>
</tr>
<tr>
<td>Inductive coupling</td>
<td><img src="static/coupling_ansi.png" alt="" /></td><td><b>K</b>name&nbsp;<b>L</b>name&nbsp;<b>L</b>name&nbsp;coeff.</td>
</tr>
<tr>
<td>Capacitance</td>
<td><img src="static/cap_din.png" alt="" /></td><td><b>C</b>name&nbsp;N1&nbsp;N2&nbsp;value</td>
</tr>
<tr>
<td>Complex<br/>impedance</td>
<td><img src="static/z_ansi.png" alt="" /></td><td>
in exponential form:
<br/>
<b>Z</b>name&nbsp;N1&nbsp;N2&nbsp;magnitude&nbsp;phase
<br/>
phase by default equal to zero;<br/>
phase in degrees - by default or the symbol <b>d</b> at the end of the value;<br/>
phase in radians - the symbol <b>r</b> at the end of the value
<br/>
<br/>
in cartesian form:
<br/>
<b>Z</b>name&nbsp;N1&nbsp;N2&nbsp;resistance&nbsp;reactance<b>i</b>
</td>
</tr>
<td>Transmission<br/>line</td>
<td><img src="static/t.png" alt="" /></td><td></td>
</tr>
<tr>
<td>RG-line:</td>
<td><img src="static/tr_ansi.png" alt="" /></td><td>
<b>TR</b>name&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;charact._impedance&nbsp;propagation_constant&nbsp;length</td>
</tr>
<tr>
<td>RGLC-line:</td>
<td><img src="static/tz_ansi.png" alt="" /></td><td>
<b>TZ</b>name&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;charact._impedance&nbsp;propagation_constant&nbsp;length
<br/>
characteristic impedance and propagation constant is given in complex form (cartesian or exponential)
</td>
</tr>
<tr>
<td>Ampermeter</td>
<td><img src="static/amp.png" alt="" /></td><td><b>PA</b>name&nbsp;N1&nbsp;N2</td>
</tr>
<tr>
<td>Voltmeter</td>
<td><img src="static/volt.png" alt="" /></td><td><b>PV</b>name&nbsp;N1&nbsp;N2</td>
</tr>
<tr>
<td>Wattmeter</td>
<td><img src="static/watt.png" alt="" /></td><td><b>PW</b>name&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;N4</td>
</tr>
<tr>
<td>Varmeter</td>
<td><img src="static/var.png" alt="" /></td><td><b>PQ</b>name&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;N4</td>
</tr>
<tr>
<td>Phasemeter</td>
<td><img src="static/ph.png" alt="" /></td><td><b>PF</b>name&nbsp;N1&nbsp;N2&nbsp;N3&nbsp;N4</td>
</tr>
</table>
<br/><br/>
<b>Comments<b><br/>
<b>*<b>comment<br/><br/>
<i>Examples of circuit description</i><br/>
<table>
<tr>
<td>DC circuit<br/><img src="static/cir_ex_ansi.png" alt="" /></td>
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
<td>AC circuit<br/><img src="static/rlc_ansi.png" alt="" /></td>
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
<td><a href="/">&lt;&lt;&lt; Back</a>
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