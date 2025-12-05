const hiddenCourses = new Set();
let originalLink = null;

function getAndCheckCalLink() {
    let input = document.getElementById("tumCalLink")
    input.removeAttribute("class")
    let value = input.value;
    if (originalLink !== null) {
        value = originalLink;
    }
    if (!value.match(/https:\/\/campus.tum.de\/tumonlinej\/.{2}\/termin\/ical\?(pStud|pPers)=[0-9,A-Z]*&pToken=[0-9,A-Z]*/i)) {
        input.setAttribute("class", "invalid")
        return undefined;
    }

    return value;
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

    const adjustedLink = new URL(calLink.replace(/https:\/\/campus.tum.de\/tumonlinej\/.{2}\/termin\/ical/i, "https://cal.tum.app").replace("\t", ""));

    // add course hide option
    const queryParams = new URLSearchParams(adjustedLink.search);
    for (const courseName of hiddenCourses) {
          queryParams.append("hide", courseName);
    }

    for (const [id, offset] of startOffsetRecurrences.entries()) {
          if (offset == 0) continue;
          queryParams.append("startOffset", id.toString() + (offset > 0 ? "+" : "") + offset.toString());
    }

    for (const [id, offset] of endOffsetRecurrences.entries()) {
          if (offset == 0) continue;
          queryParams.append("endOffset", id.toString() + (offset > 0 ? "+" : "") + offset.toString());
    }

    adjustedLink.search = queryParams;
    copyToClipboard(adjustedLink.toString());
    setCopyButton("copied");

    originalLink = calLink;
    document.getElementById("tumCalLink").value = adjustedLink.toString();
}

function reloadCourses() {
    originalLink = null;

    const calLink = getAndCheckCalLink();
    if (!calLink)
        return;

    // includes pStud and pToken
    const queryParams = new URLSearchParams(new URL(calLink).search);
    
    const url = new URL("api/courses", window.location.origin);
    url.search = queryParams;

    fetch(url)
        .then(response => {
            if (response.ok) {
                return response.json();
            }

            throw new Error(`Failed to fetch courses: ${response.text()}`);
        })
        .then(courses => {
            // add checkboxes for each course in courseAdjustList
            const courseAdjustList = document.getElementById("courseAdjustList");
            courseAdjustList.innerHTML = "";

            for (const [key, course] of Object.entries(courses)) {
                const li = document.createElement("li");
                const input = document.createElement("input");
                const recurrences = document.createElement("ul");
                input.type = "checkbox";
                input.id = course.summary;
                input.checked = !hiddenCourses.has(key);
                input.onchange = () => {
                    if (input.checked) {
                      hiddenCourses.delete(key);
                    } else {
                      hiddenCourses.add(key);
                    }
                    setCopyButton("reset");
                };
                
                for (const recurrence of Object.values(course.recurrences)) {
                  const recLi = document.createElement("li");
                  recLi.id = recurrence.recurringId;

                  const startDate = new Date(recurrence.dtStart);
                  const endDate = new Date(recurrence.dtEnd);
                  const dayOfWeek = new Intl.DateTimeFormat("de-DE", {weekday: "long"}).format(startDate);

                  recLi.appendChild(document.createTextNode(`id ${recurrence.recurringId}: `));
                  recLi.appendChild(document.createTextNode(` ${dayOfWeek}: ${startDate.toLocaleTimeString()} - ${endDate.toLocaleTimeString()} `));
                  recurrences.appendChild(recLi);
                }
                li.appendChild(input);
                li.appendChild(document.createTextNode(course.summary));
                li.appendChild(recurrences);
                courseAdjustList.appendChild(li);
            }

            // enable/disable course adjustment section depending on whether
            // courses were found
            document.getElementById("courseAdjustDiv").hidden = Object.keys(courses).length === 0;
        })
        .catch(err => {
            console.log(err);
            document.getElementById("courseAdjustDiv").hidden = true;
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
