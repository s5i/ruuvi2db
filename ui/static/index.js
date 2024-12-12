async function refresh() {
  setError('');
  let end_time_str = document.getElementById('end_time').value || "now";
  let end_time = NaN;
  if (end_time_str === "now") {
    end_time = Math.floor(Date.now() / 1000);
  }
  if (isNaN(end_time)) {
    end_time = Math.floor(Date.now() / 1000) + parseDuration(end_time_str);
  }
  if (isNaN(end_time)) {
    end_time = Math.floor(new Date(end_time_str) / 1000);
  }
  if (isNaN(end_time)) {
    setError('Bad end time.');
    return;
  }

  let duration = parseDuration(document.getElementById('duration').value || "1d");
  if (duration === undefined) {
    setError('Bad duration.');
    return;
  }

  let aliases = await fetch('/aliases.json').then(resp => { return resp.json() });
  kinds().map((kind) => {
    setGraphStaleness(kind, true);

    return new Promise(async (_resolve, _error) => {
      let resolution = Math.max(Math.floor(10 * duration / graph(kind).scrollWidth), 1);
      let end_time_trunc = end_time - (end_time % duration);
      Promise.all([
        fetch(`/data.json?kind=${kind}&end_time=${end_time_trunc}&duration=${duration}&resolution=${resolution}`).then(resp => { return resp.json() }),
        fetch(`/data.json?kind=${kind}&end_time=${end_time_trunc + duration}&duration=${duration}&resolution=${resolution}`).then(resp => { return resp.json() })
      ]).then((data) => {
        let names = {};
        data = data.flat();

        for (i in data) {
          data[i]['ts'] = new Date(data[i]['ts']);
          let ts = data[i]['ts'] / 1000;
          if (ts < end_time - duration || ts > end_time) {
            delete data[i];
            continue
          }

          for (k in data[i]) {
            if (k == 'ts') {
              continue;
            }

            let name = aliases[k] || k;
            names[name] = true;
            if (name != k) {
              data[i][name] = data[i][k];
              delete data[i][k];
            }
          }
        }

        plot(kind, data, Object.keys(names))
        setGraphStaleness(kind, false);
      });

    });
  });
}

function graph(kind) {
  return Array.from(document.getElementsByClassName("graph")).filter((graph) => { return graph.getAttribute("data-kind") == kind })[0]
}

function kinds() {
  return Array.from(document.getElementsByClassName("graph")).map((graph) => { return graph.getAttribute("data-kind") })
}

function plot(kind, data, tags) {
  c3.generate({
    bindto: "#" + graph(kind).id,
    data: {
      json: data,
      keys: { x: 'ts', value: tags },
    },
    line: {
      connect_null: true
    },
    axis: {
      x: {
        type: 'timeseries',
        tick: {
          format: '%H:%M',
          count: 25,
          culling: false
        },
      }
    },
    grid: {
      y: {
        show: true
      }
    },
    tooltip: {
      format: {
        title: function (x, _) { return x.toLocaleString("sv-SE"); }
      }
    }
  });
}

function setError(error) {
  document.getElementById('error').innerText = error;
}

function setGraphStaleness(kind, isStale) {
  graph(kind).style.backgroundColor = isStale ? "lightgray" : null;
}

function parseDuration(str) {
  var durMap = {
    'w': 604800,
    'd': 86400,
    'h': 3600,
    'm': 60,
    's': 1,
  }
  var result = [...str.matchAll(/([+-])?(\d+w)?(\d+d)?(\d+h)?(\d+m)?(\d+s)?/g)];
  for (match of result) {
    if (match[0] != str) {
      continue;
    }
    var sec = 0;
    var mul = 1;
    if (match[1] && match[1] == '-') {
      mul = -1;
    }
    for (i = 2; i < match.length; i++) {
      if (!match[i]) {
        continue;
      }
      sec += match[i].substring(0, match[i].length - 1) * durMap[match[i].substring(match[i].length - 1, match[i].length)];
    }
    return mul * sec;
  }
}

function init() {
  Array.from(document.getElementsByClassName("graph")).map((graph) => {
    graph.id = "id" + Math.random().toString(16).slice(2);
  })
  Array.from(document.getElementsByTagName("input")).map((input) => {
    input.addEventListener("keyup", function (event) {
      if (event.key === "Enter") {
        refresh();
      }
    })
  });
  refresh();
}

init();
