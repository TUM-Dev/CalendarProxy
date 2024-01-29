const hiddenCourses = new Set();
const startOffsetRecurrences = new Map();
const endOffsetRecurrences = new Map();

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

    const adjustedLink = new URL(calLink.replace(/https:\/\/campus.tum.de\/tumonlinej\/.{2}\/termin\/ical/i, "https://cal.tum.app").replace("\t", ""));

    // add course hide option
    const queryParams = new URLSearchParams(adjustedLink.search);
    for (const courseName of hiddenCourses) {
          queryParams.append("hide", courseName);
    }

    for (const [id, offset] of startOffsetRecurrences.entries()) {
          queryParams.append("startOffset" + id, offset);
    }

    for (const [id, offset] of endOffsetRecurrences.entries()) {
          queryParams.append("endOffset" + id, offset);
    }

    adjustedLink.search = queryParams;
    copyToClipboard(adjustedLink.toString());
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

                  const startOffsetInput = document.createElement("input");
                  startOffsetInput.type = "number";
                  startOffsetInput.classList.add("offset");
                  startOffsetInput.inputmode = "numeric";
                  startOffsetInput.pattern = "\d*";
                  startOffsetInput.value = recurrence.startOffsetMinutes || 0;
                  startOffsetInput.onchange = () => {
                      startOffsetRecurrences.set(recurrence.recurringId, Number(startOffsetInput.value));
                      setCopyButton("reset");
                  };

                  const endOffsetInput = document.createElement("input");
                  endOffsetInput.type = "number";
                  endOffsetInput.classList.add("offset");
                  endOffsetInput.inputmode = "numeric";
                  endOffsetInput.pattern = "\d*";
                  endOffsetInput.value = recurrence.endOffsetMinutes || 0;
                  endOffsetInput.onchange = () => {
                      endOffsetRecurrences.set(recurrence.recurringId, Number(endOffsetInput.value));
                      setCopyButton("reset");
                  };


                  recLi.appendChild(startOffsetInput);
                  recLi.appendChild(document.createTextNode(`${dayOfWeek}: ${startDate.toLocaleTimeString()} - ${endDate.toLocaleTimeString()}`));
                  recLi.appendChild(endOffsetInput);
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
