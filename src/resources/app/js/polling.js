'use strict';

let polling;
let node;
let logTable;
let logChart;
let stateChart;
let resultChart;
let aiLossChart;
let aiHeatmap;
let currentPage;
let resultChartEnt = "";
let numValEntList = {};
let pane;

function showPage(mode) {
  if(pane) {
    return;
  }
  const pages = ["log", "state", "result", "ai"];
  pages.forEach(p => {
    if (mode == p) {
      $("#" + p + "_page").removeClass("hidden");
      $("#" + p).addClass("active");
    } else {
      $("#" + p + "_page").addClass("hidden");
      $("#" + p).removeClass("active");
    }
  });
  if (mode == "ai") {
    $('.toolbar-actions input').addClass("hidden");
    $('.toolbar-actions span').addClass("hidden");
    $('.toolbar-actions button.get').addClass("hidden");
    $('.toolbar-actions button.clear').addClass("hidden");
  } else {
    $('.toolbar-actions input').removeClass("hidden");
    $('.toolbar-actions span').removeClass("hidden");
    $('.toolbar-actions button.get').removeClass("hidden");
    $('.toolbar-actions button.clear').removeClass("hidden");
  }
  if (mode == "result") {
    $('.toolbar-actions button.setting').removeClass("hidden");
  } else {
    $('.toolbar-actions button.setting').addClass("hidden");
  }
  currentPage = mode;
}

function makeLogTable() {
  const opt = {
    dom: 'lBfrtip',
    buttons: [
      {
        extend: 'copyHtml5',
        text: '<i class="fas fa-copy"></i>',
        titleAttr: 'Copy'
      },
      {
        extend: 'excelHtml5',
        text: '<i class="fas fa-file-excel"></i>',
        titleAttr: 'Excel'
      },
      {
        extend: 'csvHtml5',
        text: '<i class="fas fa-file-csv"></i>',
        titleAttr: 'CSV'
      }
    ],
    "order": [[1, "desc"]],
    "paging": true,
    "info": false,
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal": "",
      "emptyTable": "表示するログがありません。",
      "info": "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty": "",
      "infoFiltered": "(全 _MAX_ 件)",
      "infoPostFix": "",
      "thousands": ",",
      "lengthMenu": "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing": "処理中...",
      "search": "フィルター:",
      "zeroRecords": "一致するログがありません。",
      "paginate": {
        "first": "最初",
        "last": "最後",
        "next": "次へ",
        "previous": "前へ"
      },
      "aria": {
        "sortAscending": ": 昇順でソート",
        "sortDescending": ": 降順でソート"
      }
    },
  }
  logTable = $('#log_table').DataTable(opt);
}

function makeLogChart() {
  const option = {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [{
      offset: 0,
      color: '#4b5769'
    }, {
      offset: 1,
      color: '#404a59'
    }]),
    tooltip: {
      trigger: 'axis',
      formatter: function (params) {
        const p = params[0];
        return p.name + ' : ' + p.value[1];
      },
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: "5%",
      right: "5%",
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: '日時',
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: "8px",
        formatter: function (value, index) {
          var date = new Date(value);
          return echarts.format.formatTime('MM/dd hh:mm', date)
        }
      },
      axisLine: {
        lineStyle: {
          color: '#ccc'
        }
      },
      splitLine: {
        show: false
      },
    },
    yAxis: {
      name: '件数',
      type: 'value',
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc'
        }
      },
    },
    series: [{
      type: 'bar',
      color: "#1f78b4",
      large: true,
      data: [],
    }]
  };
  logChart = echarts.init(document.getElementById('log_chart'));
  logChart.setOption(option);
}

function makeStateChart() {
  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [{
      offset: 0,
      color: '#4b5769'
    }, {
      offset: 1,
      color: '#404a59'
    }]),
    color: ["#e31a1c", "#fb9a99", "#dfdf22", "#33a02c", "#999"],
    legend: {
      orient: "vertical",
      top: 50,
      right: 10,
      textStyle: {
        fontSize: 10,
        color: "#ccc",
      },
      data: ['重度', '軽度', '注意', '正常', '不明']
    },
    grid: {
      top: '3%',
      left: '7%',
      right: '10%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      name: '件数',
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc'
        }
      }
    },
    yAxis: {
      type: 'category',
      name: '時間帯',
      axisLine: {
        show: false,
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 8,
        margin: 2,
      },
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      data: []
    },
    series: [
      {
        name: '重度',
        type: 'bar',
        stack: '回数',
        data: []
      },
      {
        name: '軽度',
        type: 'bar',
        stack: '回数',
        data: []
      },
      {
        name: '注意',
        type: 'bar',
        stack: '回数',
        data: []
      },
      {
        name: '正常',
        type: 'bar',
        stack: '回数',
        data: []
      },
      {
        name: '不明',
        type: 'bar',
        stack: '回数',
        data: []
      },
    ],
  };
  stateChart = echarts.init(document.getElementById('state_chart'));
  stateChart.setOption(option);
}

function makeResultChart() {
  const option = {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [{
      offset: 0,
      color: '#4b5769'
    }, {
      offset: 1,
      color: '#404a59'
    }]),
    tooltip: {
      trigger: 'axis',
      formatter: function (params) {
        const p = params[0];
        return p.name + ' : ' + p.value[1];
      },
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: "10%",
      right: "5%",
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: '日時',
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: "8px",
        formatter: function (value, index) {
          var date = new Date(value);
          return echarts.format.formatTime('MM/dd hh:mm', date)
        }
      },
      axisLine: {
        lineStyle: {
          color: '#ccc'
        }
      },
      splitLine: {
        show: false
      },
    },
    yAxis: {
      type: 'value',
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc'
        }
      },
    },
    series: [{
      color: "#1f78b4",
      type: 'line',
      showSymbol: false,
      hoverAnimation: false,
      data: [],
    }]
  };
  resultChart = echarts.init(document.getElementById('result_chart'));
  resultChart.setOption(option);
}

function makeAILossChart() {
  const option = {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [{
      offset: 0,
      color: '#4b5769'
    }, {
      offset: 1,
      color: '#404a59'
    }]),
    tooltip: {
      trigger: 'axis',
      formatter: function (params) {
        const p = params[0];
        return p.name + ' : ' + p.value[1];
      },
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: "10%",
      right: "5%",
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: '時刻',
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: "8px",
        color: '#ccc',
        formatter: function (value, index) {
          var date = new Date(value);
          return echarts.format.formatTime('hh:mm:ss', date)
        }
      },
      splitLine: {
        show: false
      },
      axisLine: {
        lineStyle: {
          color: '#ccc'
        }
      }
    },
    yAxis: {
      type: 'value',
      name: '誤差',
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc'
        }
      }
    },
    series: [{
      color: "#1f78b4",
      type: 'line',
      showSymbol: false,
      hoverAnimation: false,
      data: [],
    }]
  };
  aiLossChart = echarts.init(document.getElementById('ai_loss_chart'));
  aiLossChart.setOption(option);
}

function makeAIHeatmap() {
  var hours = ['0時', '1時', '2時', '3時', '4時', '5時', '6時',
    '7時', '8時', '9時', '10時', '11時',
    '12時', '13時', '14時', '15時', '16時', '17時',
    '18時', '19時', '20時', '21時', '22時', '23時'];

  const option = {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [{
      offset: 0,
      color: '#4b5769'
    }, {
      offset: 1,
      color: '#404a59'
    }]),
    grid: {
      left: "10%",
      right: "5%",
      top: 30,
      buttom: 0,
    },
    tooltip: {
      trigger: 'item',
      formatter: function (params) {
        return params.name + ' ' + params.data[1] + '時 : ' + params.data[2];
      },
      axisPointer: {
        type: 'shadow'
      }
    },
    xAxis: {
      type: 'category',
      name: '日付',
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc'
        }
      },
      data: []
    },
    yAxis: {
      type: 'category',
      name: '時間帯',
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc'
        }
      },
      data: hours
    },
    visualMap: {
      min: 40,
      max: 100,
      textStyle: {
        color: '#ccc',
        fontSize: 8
      },
      calculable: true,
      realtime: false,
      inRange: {
        color: ['#313695', '#4575b4', '#74add1', '#abd9e9', '#e0f3f8', '#ffffbf', '#fee090', '#fdae61', '#f46d43', '#d73027', '#a50026']
      }
    },
    series: [{
      name: 'Score',
      type: 'heatmap',
      data: [],
      emphasis: {
        itemStyle: {
          borderColor: '#ccc',
          borderWidth: 1
        }
      },
      progressive: 1000,
      animation: false
    }]
  };
  aiHeatmap = echarts.init(document.getElementById('ai_heatmap'));
  aiHeatmap.setOption(option);
  aiHeatmap.on('dblclick', function (params) {
    const d = params.name + ' ' + params.data[1] + ":00:00";
    $(".toolbar-actions input[name=start]").val(moment(d).subtract(2, "h").format("Y-MM-DDTHH:00"));
    $(".toolbar-actions input[name=end]").val(moment(d).add(2, "h").format("Y-MM-DDTHH:00"));
    showPage("result");
    $('.toolbar-actions button.get').click();
    logChart.resize();
  });
}

function setupTimeVal() {
  $(".toolbar-actions input[name=start]").val(moment().subtract(12, "h").format("Y-MM-DDTHH:00"));
  $(".toolbar-actions input[name=end]").val(moment().add(1, "h").format("Y-MM-DDTHH:00"));
}

function clearData() {
  logTable.clear();
  logTable.draw();
  logChart.setOption({
    series: [{
      data: []
    }]
  });
  logChart.resize();
  const optState = {
    yAxis: {
      data: [],
    },
    series: [
      { data: [] },
      { data: [] },
      { data: [] },
      { data: [] },
      { data: [] },
    ]
  };
  stateChart.setOption(optState);
  stateChart.resize();
  resultChart.setOption({
    yAxis: {
      name: '',
    },
    series: [{
      data: []
    }]
  });
  resultChart.resize();
  aiLossChart.setOption({
    series: [{
      data: []
    }]
  });
  aiLossChart.resize();
  aiHeatmap.setOption({
    xAxis: {
      data: []
    },
    series: [{
      data: []
    }]
  });
  aiHeatmap.resize();
  resultChartEnt = "";
}

document.addEventListener('astilectron-ready', function () {
  showPage("log");
  setupTimeVal();
  makeLogTable();
  makeLogChart();
  makeStateChart();
  makeResultChart();
  makeAILossChart();
  makeAIHeatmap();
  logChart.resize();
  stateChart.resize();
  resultChart.resize();
  aiLossChart.resize();
  aiHeatmap.resize();
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "setParams":
        if (message.payload && message.payload.Polling) {
          polling = message.payload.Polling;
          node = message.payload.Node;
          if (polling.LogMode == 3) {
            $('#ai').removeClass('hidden');
            $('.toolbar-actions button.ai').removeClass('hidden');
          } else {
            $('#ai').addClass('hidden');
            $('.toolbar-actions button.ai').addClass('hidden');
          }
          setWindowTitle(node.Name, polling.Name);
          clearData();
          setupTimeVal();
          showPage("log");
          $('.toolbar-actions button.get').click();
          logChart.resize();
        }
        return { name: "setParams", payload: "ok" };
      case "error":
        setTimeout(() => {
          dialog.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
  $('#log').click(() => {
    showPage("log");
    logChart.resize();
  });
  $('#state').click(() => {
    showPage("state");
    stateChart.resize();
  });
  $('#chart').click(() => {
    showPage("result");
    resultChart.resize();
  });
  $('#ai').click(() => {
    if(pane){
      return;
    }
    showPage("ai");
    updateAIPage();
  });
  $('.toolbar-actions button.close').click(() => {
    astilectron.sendMessage({ name: "close", payload: "" }, message => {
    });
  });
  $('.toolbar-actions button.clear').click(() => {
    if(pane){
      return;
    }
    if (!confirmDialog("ログクリア", "ポーリングログをクリアしますか?")) {
      return;
    }
    astilectron.sendMessage({ name: "clear", payload: polling.ID }, message => {
      if(pane){
        return;
      }
      clearData();
      showPage("log");
    });
  });
  $('.toolbar-actions button.get').click(() => {
    if(pane){
      return;
    }
    updateData();
  });
  $('.toolbar-actions button.setting').click(() => {
    createChartSettingPane();
  });
  $('.toolbar-actions button.ai').click(() => {
    astilectron.sendMessage({ name: "doai", payload: polling.ID }, message => {
      if (message && message.payload == "ok") {
        dialog.showMessageBox({ message: "AI再分析開始しました。", title: "AI分析" });
      } else {
        dialog.showErrorBox("エラー", "AI再分析できません。");
      }
    });
  });
});

function updateData() {
  const params = {
    PollingID: polling.ID,
    StartTime: $(".toolbar-actions input[name=start]").val(),
    EndTime: $(".toolbar-actions input[name=end]").val()
  }
  $('.toolbar-actions button.get').prop("disabled", true);
  $('#wait').removeClass("hidden");
  astilectron.sendMessage({ name: "get", payload: params }, message => {
    $('#wait').addClass("hidden");
    $('.toolbar-actions button.get').prop("disabled", false);
    const logs = message.payload;
    if (typeof logs === "string") {
      setTimeout(() => {
        dialog.showErrorBox("エラー", message.payload);
      }, 100);
      return;
    }
    if (logs.length > 0) {
      for(let i=0;i < logs.length ;i++){
        if(logs[i].State == "normal" || logs[i].State == "repair" ){
          setDataList(logs[i].StrVal);
          break;
        }
      }
    }
    const dataTimeLine = [];
    const dataResult = [];
    const optState = {
      yAxis: {
        data: [],
      },
      series: [
        { data: [] },
        { data: [] },
        { data: [] },
        { data: [] },
        { data: [] },
      ]
    };
    const stateCount = {
      high: 0,
      low: 0,
      warn: 0,
      normal: 0,
      unknown: 0,
    }
    let intCount = 5;
    let intState = 1;
    if (logs.length > 2) {
      const dh = (logs[logs.length - 1].Time - logs[0].Time) / (1000 * 1000 * 1000 * 3600);
      intState = Math.round(dh / 4);
      if (intState < 1) {
        intState = 1;
      }
      if (intState > 24) {
        intCount = 60;
      }
    }
    let count = 0;
    let ctc;
    let cts;
    let dp = getDispParams();
    logTable.clear();
    logs.forEach(l => {
      const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
      const state = getStateHtml(l.State)
      logTable.row.add([state, ts, l.NumVal, l.StrVal]);
      let numVal = l.NumVal;
      if (resultChartEnt != "") {
        numVal = getNumVal(resultChartEnt, l.StrVal);
      }
      numVal *= dp.mul;
      dataResult.push({
        name: ts,
        value: [new Date(l.Time / (1000 * 1000)), numVal],
      });
      const newCtc = Math.floor(l.Time / (1000 * 1000 * 1000 * 60 * intCount));
      if (!ctc) {
        ctc = newCtc;
      }
      if (ctc != newCtc) {
        let t = new Date(ctc * 60 * intCount * 1000);
        dataTimeLine.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm:ss', t),
          value: [t, count]
        });
        ctc = newCtc;
        count = 0;
      }
      count++;
      const newCts = Math.floor(l.Time / (1000 * 1000 * 1000 * 60 * 60 * intState));
      if (!cts) {
        cts = newCts;
      }
      if (cts != newCts) {
        let t = new Date(cts * 60 * 60 * 1000 * intState);
        optState.yAxis.data.push(echarts.format.formatTime('MM/dd hh:mm', t));
        optState.series[0].data.push(stateCount.high);
        optState.series[1].data.push(stateCount.low);
        optState.series[2].data.push(stateCount.warn);
        optState.series[3].data.push(stateCount.normal);
        optState.series[4].data.push(stateCount.unknown);
        cts = newCts
        stateCount.high = 0;
        stateCount.low = 0;
        stateCount.warn = 0;
        stateCount.normal = 0;
        stateCount.unknown = 0;
      }
      switch (l.State) {
        case "high":
          stateCount.high++;
          break;
        case "low":
          stateCount.low++;
          break;
        case "warn":
          stateCount.warn++;
          break;
        case "normal":
        case "repair":
          stateCount.normal++;
          break;
        default:
          stateCount.unknown++;
          break;
      }
    });
    if (count > 0) {
      let t = new Date(ctc * 60 * intCount * 1000);
      dataTimeLine.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm:ss', t),
        value: [t, count]
      });
      t = new Date(cts * 60 * 60 * 1000 * intState);
      optState.yAxis.data.push(echarts.format.formatTime('MM/dd hh:mm', t));
      optState.series[0].data.push(stateCount.high);
      optState.series[1].data.push(stateCount.low);
      optState.series[2].data.push(stateCount.warn);
      optState.series[3].data.push(stateCount.normal);
      optState.series[4].data.push(stateCount.unknown);
    }
    logTable.draw();
    logChart.setOption({
      series: [{
        data: dataTimeLine
      }]
    });
    logChart.resize();
    stateChart.setOption(optState);
    stateChart.resize();
    resultChart.setOption({
      yAxis: {
        name: dp.axis
      },
      series: [{
        data: dataResult
      }]
    });
    resultChart.resize();
  });
}

function setWindowTitle(n, p) {
  const t = "ポーリング分析 - " + n + " - " + p;
  $("title").html(t);
  $("h1.title").html(t);
}

function setDataList(s) {
  numValEntList = { "数値データ": "" };
  try {
    JSON.parse(s, (k, v) => {
      if (k != "" && parseFloat(v) != NaN) {
        numValEntList[getChartModeName(k)] = k
      }
    });
  } catch (e) {
    console.log(e);
  }
}

function getNumVal(ent, s) {
  let ret = 0.0;
  try {
    JSON.parse(s, (k, v) => {
      if (k == ent) {
        const nv = parseFloat(v);
        if (nv != NaN) {
          ret = nv;
        }
      }
    });
  } catch (e) {
    console.log(e);
  }
  return ret;
}

function getDispParams() {
  let resultChartMode = "";
  if (resultChartEnt == "") {
    // 数値の場合はポーリングの種類から選ぶ
    switch (polling.Type) {
      case "ping":
        if (polling.Polling == "line") {
          resultChartMode = "speed";
          break;
        }
      case "tcp":
      case "http":
      case "https":
      case "dns":
      case "ntp":
        resultChartMode = "rtt"
        break;
      case "sysloguser":
        resultChartMode = "successRate"
        break;
      case "snmp":
        const r = snmpChartDispInfo();
        if (r) {
          return r;
        }
      default:
        resultChartMode = "none"
        break;
    }
  } else {
    resultChartMode = resultChartEnt;
  }
  const r = chartDispInfo[resultChartMode];
  if (r) {
    return r;
  }
  return {
    mul: 1.0,
    axis: ""
  }
}

function getChartModeName(mode) {
  const r = chartDispInfo[mode];
  if (r) {
    return r.axis;
  }
  return(mode);
}

function snmpChartDispInfo(){
  const a = polling.Polling.split("|");
  if(a.length < 4 ) {
    return undefined;
  }
  let p = a[a.length-1].split(",");
  if (p.length != 2){
    return undefined;
  }
  return {
    mul: 1.0 * p[0],
    axis: p[1]
  }
}

function updateAIPage() {
  astilectron.sendMessage({ name: "getai", payload: polling.ID }, message => {
    let aiData = message.payload;
    if (!aiData || !aiData.ScoreData || aiData.ScoreData.length < 1) {
      setTimeout(() => {
        dialog.showErrorBox("AI分析", "該当する分析データがありません。");
      }, 100);
      aiData = {
        ScoreData: [],
        LossData: []
      };
    }
    const lossChartData = [];
    aiData.LossData.forEach(e => {
      const t = new Date(e[0]);
      lossChartData.push({
        name: echarts.format.formatTime('hh:mm:ss', t),
        value: [t, e[1]]
      });
    });
    aiLossChart.setOption({
      series: [{
        data: lossChartData
      }]
    });
    aiLossChart.resize();
    const heatmapX = [];
    const heatmapVal = [];
    let nD = 0;
    let x = -1;
    aiData.ScoreData.forEach(e => {
      const t = new Date(e[0] * 1000);
      if (nD != t.getDate()) {
        heatmapX.push(echarts.format.formatTime('yyyy/MM/dd', t))
        nD = t.getDate()
        x++;
      }
      heatmapVal.push([x, t.getHours(), e[1]])
    });
    aiHeatmap.setOption({
      xAxis: {
        data: heatmapX
      },
      series: [{
        data: heatmapVal
      }]
    });
    aiHeatmap.resize();
  });
}

function createChartSettingPane() {
  if (pane) {
    return;
  }
  const p = {
    Ent: resultChartEnt,
  };
  pane = new Tweakpane({
    title: "グラフ設定",
  });
  pane.addInput(p, 'Ent', {
    label: "表示項目",
    options: numValEntList,
  });
  pane.addButton({
    title: '取消',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: '適用',
  }).on('click', (value) => {
    resultChartEnt = p.Ent;
    pane.dispose();
    pane = undefined;
    $('.toolbar-actions button.get').click();
  });
}

