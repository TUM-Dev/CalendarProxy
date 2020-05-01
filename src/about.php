<html lang="en">
<head>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/github-fork-ribbon-css/0.2.3/gh-fork-ribbon.min.css"/>
    <title>TUM Calendar Proxy - clean, structured calendar entries</title>
</head>
<body>
<a class="github-fork-ribbon" href="https://github.com/TUM-Dev/CalendarProxy/" data-ribbon="Contribute on GitHub" title="Contribute on GitHub">Contribute on GitHub</a>

<h1>TUM Calendar Proxy</h1>
<img src="showcase.png" alt="New and old representation of the calender"/>

<h2>About</h2>
<p>Nice and easy proxy to remove some clutter from the TUM online iCal export. E.g.:</p>
<ul>
    <li>Shorten Lesson Names like 'Grundlagen Betriebssysteme und Systemsoftware' → 'GBS'</li>
    <li>Adds locations, which are understood by Google Maps / Google Now</li>
    <li>Replaces 'Tutorübung' with 'TÜ'</li>
    <li>Remove event duplicates due to multiple rooms</li>
</ul>

<h2>HowTo</h2>
<ol>
    <li>Grab the URL from the <a href="https://campus.tum.de/tumonline/wbKalender.wbPerson">TUMO calendar</a> via the 'Veröffentlichen' button</li>
    <li>Copy 'n' Paste the query string (everything after the ? sign, e.g. "?pStud=ABCDEF&pToken=XYZ") and append it to this url:
        <a href="#dontclickme">https://cal.bruck.me/</a> so it looks like this: <a href="#dontclickme">https://cal.bruck.me/?pStud=ABCDEF&pToken=XYZ</a>
    </li>
    <li>???</li>
    <li>Profit!</li>
    <li>Go to Google Calendar (or similar) and import the resulting url</li>
</ol>
<h2></h2>


<h3>Contribute / Suggest</h3>
If you want to suggest something create an issue at <a href="https://github.com/TUM-Dev/CalendarProxy/issues">Github</a>

<br/>
<br/>

<span style="font-size:10px;color:#aaa;">Version v1.2 - <a href="https://github.com/kordianbruck/TumCalProxy/commits/master">Changelog</a></span>

</body>
</html>
