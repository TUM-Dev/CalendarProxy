function generateLink() {
    let input = document.getElementById("tumCalLink")
    input.removeAttribute("class")
    if (!input.value.match(/https:\/\/campus.tum.de\/tumonlinej\/.{2}\/termin\/ical\?pStud=[0-9,A-Z]*&pToken=[0-9,A-Z]*/i)) {
        input.setAttribute("class", "invalid")
        return
    }
    copyToClipboard(input.value.replace(/https:\/\/campus.tum.de\/tumonlinej\/.{2}\/termin\/ical/i, "https://cal.bruck.me").replace("\t", ""))
    let btn = document.getElementById("generateLinkBtn")
    btn.innerText = "copied!"
    btn.setAttribute("style", "background-color: #4CAF50;")
}


function copyToClipboard(text) {
    const dummy = document.createElement("textarea");
    document.body.appendChild(dummy);
    dummy.value = text;
    dummy.select();
    document.execCommand("copy");
    document.body.removeChild(dummy);
}
