<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">
<html>
<head>
<script type="text/javascript" src="https://vk.com/js/api/share.js?95" charset="windows-1251"></script>
<link rel="stylesheet" type="text/css" href="static/style.css" />
<meta http-equiv="content-type" content="text/html; charset=utf-8">
<link rel="icon" type="image/png" href="http://foxylab.com:7777/static/favicon.png" />
<link rel="shortcut icon" type="image/png" href="http://foxylab.com:7777/static/favicon.png" />
<meta name="description" content="Circuit simulator FoxySim">
<meta name="keywords" content="Alexey V.Voronin, Алексей Воронин, Гомель, Беларусь, БелГУТ, Белорусский государственный университет транспорта, электротехника, ТОЭ, теоретические основы электротехники">
<meta name="robots" content="noarchive">
<meta name="author" content="Alexey FoxyLab Voronin">
<meta name="viewport" content="width=device-width">
<meta name="robots" content="noarchive">
<meta name="googleBOT" content="noarchive">
<title>FoxySim -  онлайн-симулятор ЭЦ</title>
</head>
<body>
<!-- Yandex.Metrika counter -->
<script type="text/javascript" >
    (function (d, w, c) {
        (w[c] = w[c] || []).push(function() {
            try {
                w.yaCounter45842049 = new Ya.Metrika({
                    id:45842049,
                    clickmap:true,
                    trackLinks:true,
                    accurateTrackBounce:true,
                    ut:"noindex"
                });
            } catch(e) { }
        });

        var n = d.getElementsByTagName("script")[0],
            s = d.createElement("script"),
            f = function () { n.parentNode.insertBefore(s, n); };
        s.type = "text/javascript";
        s.async = true;
        s.src = "https://mc.yandex.ru/metrika/watch.js";

        if (w.opera == "[object Opera]") {
            d.addEventListener("DOMContentLoaded", f, false);
        } else { f(); }
    })(document, window, "yandex_metrika_callbacks");
</script>
<noscript><div><img src="https://mc.yandex.ru/watch/45842049?ut=noindex" style="position:absolute; left:-9999px;" alt="" /></div></noscript>
<!-- /Yandex.Metrika counter -->
<p>
<table style="width:300px; margin: 0px;" cellspacing="0">
<tr valign="middle">
<td><img src="static/favicon.png" alt="" /></td>
<td><b><i><strong style="font-size: 12px;">FoxySim - онлайн-симулятор ЭЦ</strong></i></b></td><td align="right"><a href="ru"><img src="static/ru.png" alt="" /></a>&nbsp;<a href="en"><img src="static/en.png" alt="" /></a></td>
</tr>
</table>
<table style="width:300px; margin: 0px" cellspacing="0">
<tr>
<td align="left"><font color="blue">Список соединений:</font>&nbsp;
<a href="help"><img src="static/help.png" alt="" /></a>
&nbsp;&nbsp;
<a href="https://acdc.foxylab.com/foxysim" target="_blank"><img src="static/acdc.png" alt="" /></a>&nbsp;&nbsp;
</td>
<td align="right" valign="middle"><form  action="/reset" method="post" style="display: inline-block; margin: 0;"><input type="submit" value="Стереть"></form></td>
</tr>
</table>
<form action="/calc" method="post">
<textarea style="width:300px;" name="netlist" rows="16" name="netlist" id="netlist" value="" class="textfield" value="" spellcheck="false">{{.Net}}</textarea>
<br/><input type="submit" value="Пуск!">
<input type="hidden" name="security" value="{{.Sec}}">
</form>
</p>
<p class="small">
Ваш IP-адрес:&nbsp;<font color="blue">{{.Ip}}</font><br/> 
<table style="width:300px; margin: 0px; margin-top: -5px;">
<tr>
<td  class="small">
<a href="https://t.me/foxysim"><img src="static/telegram.png" alt="" /></a>&nbsp;&nbsp;&nbsp;<a href="https://belsut.foxylab.com" target="_blank"><img src="static/belsut.png" alt="" /></a><br/>
<img src="static/foxylab_micro.png" alt="" />&nbsp;&copy; <i>Alexey "FoxyLab" Voronin</i><br />
<img src="static/golang_micro.png" alt="" />&nbsp;FoxySim powered by golang
</td>
<td align="right" valign="top">
<script type="text/javascript"><!--
document.write(VK.Share.button(false,{type: "custom", text: "<img src=\"https://vk.com/images/share_32.png\" width=\"32\" height=\"32\" />"}));
--></script>
</td>
</tr>
</table>
<i><u>
Этот сайт использует cookies
</u></i>
</p>
<!-- Yandex.Metrika informer -->
<a href="https://metrika.yandex.by/stat/?id=45842049&amp;from=informer"
target="_blank" rel="nofollow"><img src="https://informer.yandex.ru/informer/45842049/1_1_FFFFFFFF_EFEFEFFF_0_uniques"
style="width:80px; height:15px; border:0;" alt="Яндекс.Метрика" title="Яндекс.Метрика: данные за сегодня (уникальные посетители)" class="ym-advanced-informer" data-cid="45842049" data-lang="ru" /></a>
<!-- /Yandex.Metrika informer -->
<br/>
<a href="https://info.flagcounter.com/zwt2"><img src="https://s01.flagcounter.com/count/zwt2/bg_FFFFFF/txt_000000/border_CCCCCC/columns_4/maxflags_12/viewers_3/labels_0/pageviews_0/flags_0/percent_0/" alt="Flag Counter" border="0"></a>
</body>
</html>