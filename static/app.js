function markdownItalics(text) {
  let parts = text.split('_');
  let ret = parts[0];
  let i = 1;
  for (; i < parts.length; i++) {
    if (i % 2 == 1) {
      ret += '<em>';
    } else {
      ret += '</em>';
    }
    ret += parts[i];
  }
  if (i % 2 == 0) {
    ret += '</em>';
  }

  return ret;
}

const BASE_URL = JSON.parse(document.getElementById('base-url').innerHTML).BASE_URL

const Controller = {
  search: (opts) => {
    // Parse options
    Object.assign(Controller, opts)
    document.getElementById("main").classList.add("searched");

    document.getElementById("search-results").style.display = "initial";

    const params = Object.entries({
      q: Controller.query,
      page: Controller.page,
      length: 6,
    }).map(([key, val]) => `${key}=${val}`).join("&");

    fetch(`${BASE_URL}/search?${params}`).then((response) => {
      response.json().then((results) => {
        Controller.updateResults(results);
      });
    });
  },

  updateResults: (resObj) => {

    // Add the pagination results
    const pageControls = document.getElementById("page-buttons");
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
    const table = document.getElementById("table-body");
    table.innerHTML = '<tr><th>Work</th><th>Result</th><th></th></tr>';
    for (let result of resObj.results) {
      let row = document.createElement('tr');
      let work = document.createElement('td');
      work.innerHTML = result.work;
      row.append(work);
      let fragment = document.createElement('td');
      fragment.innerHTML = markdownItalics(result.fragments.join(' ... '));
      fragment.addEventListener('click', _ => Controller.preview(result.id));
      row.append(fragment);
      let preview = document.createElement('td');
      let previewButton = document.createElement('button');
      previewButton.innerText = "Preview";
      previewButton.addEventListener('click', _ => Controller.preview(result.id));
      preview.append(previewButton);
      row.append(preview);
      table.append(row);
    }
  },

  preview: (id) => {
    fetch(`${BASE_URL}/preview?id=${id}`).then((response) => {
      response.json().then((results) => {
        document.getElementById("preview").innerHTML = markdownItalics(results.preview);
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
