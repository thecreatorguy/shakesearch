const Controller = {
  search: (opts) => {
    // Parse options
    this.query = opts.query || this.query || ""
    this.page = opts.page || this.page || 0
    this.pageLength = opts.pageLength || this.pageLength || 10

    document.getElementById("search-results").style.display = "initial";

    const params = Object.entries({
      q: this.query,
      page: this.page,
      length: this.pageLength,
    }).map(([key, val]) => `${key}=${val}`).join("&");

    const response = fetch(`/search?${params}`).then((response) => {
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
      console.log(buttonNums)
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
    // TODO: parse this for newlines, markdown
    const rows = [];
    for (let result of resObj.results) {
      rows.push(`<tr>${result}<tr/>`);
    }

    const table = document.getElementById("table-body");
    table.innerHTML = rows;
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", ev => {
  ev.preventDefault()
  const form = document.getElementById("form");
  const data = Object.fromEntries(new FormData(form));
  Controller.search({query: data.query});
});
document.getElementById("search-results").style.display = "none";
