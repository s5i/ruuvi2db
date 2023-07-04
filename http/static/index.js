function refresh() {
  document.getElementById('error').innerText = '';
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
    document.getElementById('error').innerText = 'Bad end time.';
    return;
  }

  let duration = parseDuration(document.getElementById('duration').value || "1d");
  if (duration === undefined) {
    document.getElementById('error').innerText = 'Bad duration.';
    return;
  }

  let end_time_trunc = end_time - (end_time % 3600);
  if (end_time_trunc != end_time) {
    end_time_trunc += 3600;
  }

  let start_time = end_time - duration;
  let start_time_trunc = start_time - ((start_time) % 3600);

  let kinds = ["temperature", "humidity", "pressure", "battery"];
  kinds.map((kind) => {
    document.getElementById('graph-' + kind).style.backgroundColor = "lightgray";
  });

  fetch('/tags.json').then(resp => { return resp.json() }).then(tags => {
    let updateKind = (kind) => {
      return new Promise(async (resolve, _) => {
        let promises = [];
        for (let et = end_time_trunc; et > start_time_trunc; et -= 3600) {
          promises.push(fetch(`/data.json?kind=${kind}&end_time=${et}&duration=3600`).then(resp => { return resp.json() }));
        }
        Promise.all(promises).then((values) => {
          let preFilterData = values.flat();
          let skip = Math.max(Math.floor(preFilterData.length / 250), 1);
          let begin = preFilterData.length % skip;
          let data = [];
          for (let i = begin; i < preFilterData.length; i += skip) {
            preFilterData[i]['ts'] = new Date(preFilterData[i]['ts']);
            let ts = Math.floor(preFilterData[i]['ts'] / 1000);
            if (ts < start_time || ts > end_time) {
              continue;
            }
            data.push(preFilterData[i]);
          }

          for (i in data) {
            data[i]['ts'] = new Date(data[i]['ts']);
          }
          graph(kind, data, tags)
          document.getElementById('graph-' + kind).style.backgroundColor = null;
        });

      });
    };
    kinds.map((kind) => updateKind(kind));
  });
}

function graph(kind, data, tags) {
  c3.generate({
    bindto: '#graph-' + kind,
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

function bindEnter() {
  document.getElementById("end_time").addEventListener("keyup", function (event) {
    if (event.key === "Enter") {
      refresh();
    }
  });
  document.getElementById("duration").addEventListener("keyup", function (event) {
    if (event.key === "Enter") {
      refresh();
    }
  });
}

bindEnter();
refresh();
