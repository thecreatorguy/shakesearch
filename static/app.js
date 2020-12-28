const Controller = {
  search: (opts) => {
    // Parse options
    Object.assign(Controller, opts)
    document.getElementById("main").classList.add("searched");

    document.getElementById("search-results").style.display = "initial";

    const params = Object.entries({
      q: Controller.query,
      page: Controller.page,
      length: 8,
    }).map(([key, val]) => `${key}=${val}`).join("&");

    fetch(`/search?${params}`).then((response) => {
      response.json().then((results) => {
        Controller.updateResults(results);
      });
    });
  },

  updateResults: (resObj) => {

    // Add the pagination results
    const pageControls = document.getElementById("page-controls");
    pageControls.innerHTML = "";

    maxPage = Math.floor(resObj.total / resObj.length);
    buttonNums = Array.from(new Set([0, maxPage, resObj.page - 1, resObj.page, resObj.page + 1]))
      .sort((a, b) => a - b).filter(x => x >= 0 && x <= maxPage);
    for (let i = 0; i < buttonNums.length; i++) {
      let b = document.createElement("button")
      b.innerHTML = buttonNums[i] + 1;
      if (buttonNums[i] == resObj.page) {
        b.setAttribute("readonly", true);
        b.setAttribute("disabled", true);
      } else {
        b.addEventListener("click", ev => {
          ev.preventDefault();
          Controller.search({page: buttonNums[i]});
        })
      }
      pageControls.append(b)
      
      if (i < buttonNums.length - 1 && buttonNums[i] < buttonNums[i+1] - 1) {
        pageControls.append("...")
      }
    }

    // Render the results
    const rows = [];
    for (let result of resObj.results) {
      rows.push(`<tr><td>${result.work}</td><td onclick="Controller.preview('${result.id}');">${result.fragments.join(' ... ')}</td></tr>`);
    }

    const table = document.getElementById("table-body");
    table.innerHTML = `<tr><th>Work</th><th>Result</th></tr>${rows.join('')}`;
  },

  preview: (id) => {
    fetch(`/preview?id=${id}`).then((response) => {
      response.json().then((results) => {
        document.getElementById("preview").innerText = results.preview;
      });
    });
  }
};
Controller.query = ""
Controller.page = 0
Controller.pageLength = 10

const form = document.getElementById("form");
form.addEventListener("submit", ev => {
  ev.preventDefault()
  const form = document.getElementById("form");
  const data = Object.fromEntries(new FormData(form));
  Controller.search({query: data.query, page: 0});
});
document.getElementById("search-results").style.display = "none";
