function getData(end_time, duration) {
  var data = Papa.parse(`/csv?end_time=${end_time}&duration=${duration}`, {
    download: true,
    delimiter: ",",
    header: true,
    skipEmptyLines: true,
    complete: function(results) {
      refreshGraphs(results.data, end_time, duration);
    },
  });
}

function refreshGraphs(data, end_time, duration) {
  var names = {};
  var datapoints = {};
  var columns = {};
  var dimensions = ['temperature', 'humidity', 'pressure', 'battery'];

  for (row of data) {
    names[row.name] = true;

    if (!datapoints[row.timestamp]) {
      datapoints[row.timestamp] = {};
      for (d of dimensions) {
        datapoints[row.timestamp][d] = {};
      }
    }
    for (d of dimensions) {
      datapoints[row.timestamp][d][row.name] = row[d];
    }
  }

  for (d of dimensions) {
    columns[d] = [['ts']];
    for (name in names) {
      columns[d].push([name]);
    }

    for (ts in datapoints) {
      var date = new Date(0);
      date.setUTCSeconds(ts);

      for (i in columns[d]) {
        if (columns[d][i][0] == 'ts') {
          columns[d][i].push(date);
          continue;
        }
        columns[d][i].push(
          datapoints[ts][d][columns[d][i][0]] || null
        );
      }
    }

    var minX = new Date(0);
    minX.setUTCSeconds(end_time - duration);

    var maxX = new Date(0);
    maxX.setUTCSeconds(end_time);

    c3.generate({
        bindto: '#graph-' + d,
        data: {
            x: 'ts',
            columns: columns[d]
        },
        line: {
          connect_null: true
        },
        axis: {
            x: {
                type: 'timeseries',
                tick: {
                    format: '%Y-%m-%d %H:%M'
                },
                max: maxX,
                min: minX
            }
        },
        grid: {
            y: {
                show: true
            }
        }
    });
  }
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
    return sec;
  }
}

function refresh() {
  var end_time = document.getElementById('end_time').value || "now";
  var duration = document.getElementById('duration').value || "1d";
  var t = Date.now();
  if (end_time != "now") {
    t = new Date(end_time);
  }
  getData(Math.floor(t / 1000), parseDuration(duration));
}

function bindEnter() {
  document.getElementById("end_time").addEventListener("keyup", function(event) {
    if (event.key === "Enter") {
      refresh();
    }
  });
  document.getElementById("duration").addEventListener("keyup", function(event) {
    if (event.key === "Enter") {
      refresh();
    }
  });
}

bindEnter();
refresh();
