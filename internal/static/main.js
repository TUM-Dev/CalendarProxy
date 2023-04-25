const filteredCourses = {};

function getAndCheckCalLink() {
    let input = document.getElementById("tumCalLink")
    input.removeAttribute("class")
    if (!input.value.match(/https:\/\/campus.tum.de\/tumonlinej\/.{2}\/termin\/ical\?(pStud|pPers)=[0-9,A-Z]*&pToken=[0-9,A-Z]*/i)) {
        input.setAttribute("class", "invalid")
        return undefined;
    }

    return input.value;
}

function setCopyButton(state /* copied | reset */) {
    const btn = document.getElementById("generateLinkBtn");

    const isCopiedState = state === "copied";
    btn.innerText = isCopiedState ? "copied!" : "Generate & Copy";
    btn.setAttribute("style", `background-color: ${isCopiedState ? "#4CAF50" : "#007cea"};`);
}

function generateLink() {
    const calLink = getAndCheckCalLink();
    if (!calLink)
        return;

    let adjustedLink = calLink.replace(/https:\/\/campus.tum.de\/tumonlinej\/.{2}\/termin\/ical/i, "https://cal.tum.app").replace("\t", "");
    for (const [tag, shouldFilter] of Object.entries(filteredCourses)) {
        if (shouldFilter)
            adjustedLink += "&filterTag=" + tag;
    }

    copyToClipboard(adjustedLink);
    setCopyButton("copied")
}

function reloadCourses() {
    const calLink = getAndCheckCalLink();
    if (!calLink)
        return;

    // includes pStud and pToken
    const queryParams = new URLSearchParams(new URL(calLink).search);
    
    const url = new URL("api/courses", window.location.origin);
    url.search = queryParams;

    fetch(url)
        .then(response => {
            if (response.ok)
                return response.json();

            throw new Error(`Failed to fetch courses: ${response.text()}`);
        })
        .then(courses => {
            // add checkboxes for each course in courseFilterList
            const courseFilterList = document.getElementById("courseFilterList");
            courseFilterList.innerHTML = "";

            // courses is a dictionary of lecture tag -> lecture name pairs
            for (const [tag, name] of Object.entries(courses)) {
                const li = document.createElement("li");
                const input = document.createElement("input");
                input.type = "checkbox";
                input.id = tag;
                input.checked = !filteredCourses[tag];
                input.onchange = () => {
                    filteredCourses[tag] = !input.checked;
                    setCopyButton("reset");
                };
                li.appendChild(input);
                li.appendChild(document.createTextNode(name));
                courseFilterList.appendChild(li);
            }

            // hide or un-hide course filter section depending on whether
            // courses were found
            document.getElementById("courseFilterDiv").hidden = Object.keys(courses).length === 0;
        })
        .catch(err => {
            console.log(err);
            document.getElementById("courseFilterDiv").hidden = true;
        });
}

function copyToClipboard(text) {
    const dummy = document.createElement("textarea");
    document.body.appendChild(dummy);
    dummy.value = text;
    dummy.select();
    document.execCommand("copy");
    document.body.removeChild(dummy);
}
