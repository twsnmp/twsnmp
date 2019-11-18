'use strict';

let currentPage = "";
let nodes;
let polling;
let logTable;
let syslogTable;
let trapTable;
let netflowTable;
let ipfixTable;
let logChart;
let syslogChart;
let trapChart;
let netflowChart;
let ipfixChart;
let pane;
let pollingList;
const searchHistory = [];

function setupPollingPage() {
  astilectron.sendMessage({ name: "getLogPollings", payload: "" }, message => {
    if (message.payload === "ng") {
      astilectron.showErrorBox("ログ表示", "ログポーリングを取得できません。");
      return;
    }
    polling.rows().remove();
    pollingList = {};
    for (let i = message.payload.length - 1; i >= 0; i--) {
      const p = message.payload[i];
      const lt = moment(p.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
      const level = getStateHtml(p.Level);
      const state = getStateHtml(p.State);
      polling.row.add([state, p.Name, level, p.Type, p.Polling, lt, p.ID]);
      pollingList[p.ID] = p;
    }
    polling.draw();
    setPollingBtns(false);
  });
  astilectron.sendMessage({ name: "getNodes", payload: "" }, message => {
    if (message.payload === "ng") {
      astilectron.showErrorBox("ログ表示", "ノードを取得できません。");
      return;
    }
    nodes = message.payload;
  });
}

function searchLog() {
  const filter = {
    StartTime: $(".log_btns input[name=start]").val(),
    EndTime: $(".log_btns input[name=end]").val(),
    Filter: $(".log_btns input[name=filter]").val(),
    LogType: currentPage
  }
  astilectron.sendMessage({ name: "searchLog", payload: filter }, message => {
    if (message.payload == "ng") {
      astilectron.showErrorBox("ログ表示", "ログを取得できません。");
      return;
    }
    if(filter.Filter &&  !searchHistory.includes(filter.Filter)){
      searchHistory.push(filter.Filter);
    } 
    switch (currentPage) {
      case "log":
        showLog(message.payload);
        break;
      case "syslog":
        showSyslog(message.payload);
        break;
      case "trap":
        showTrap(message.payload);
        break;
      case "netflow":
        showNetflow(message.payload);
        break;
      case "ipfix":
        showIpfix(message.payload);
        break;
      default:
        astilectron.showErrorBox("ログ表示", "内部エラー表示内容の不整合");
    }
  });
}

function showLog(logList) {
  const data = [];
  let count = 0;
  let ctm;
  logTable.rows().remove();
  for (let i = logList.length - 1; i >= 0; i--) {
    const l = logList[i]
    const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
    const lvl = getStateHtml(l.Level)
    logTable.row.add([lvl, ts, l.Type, l.NodeName, l.Event]);
    if(!ctm ) {
      ctm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
      count++;
      continue;
    }
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
    if (ctm != newCtm) {
      let t = new Date(ctm * 60 * 1000);
      data.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm:ss', t),
        value: [t,count]
      });
      ctm--;
      for(;ctm > newCtm;ctm--) {
        t = new Date(ctm * 60 * 1000);
        data.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm:ss', t),
          value: [t,0]
        });
      }
      count=0;
    }
    count++;
  }
  logTable.draw();
  logChart.setOption({
    series: [{
      data: data
    }]
  });
  logChart.resize();
}

function showSyslog(logList) {
  const data = [];
  let count = 0;
  let ctm;
  syslogTable.rows().remove();
  for (let i = logList.length - 1; i >= 0; i--) {
    const l = logList[i]
    if (!l) {
      continue;
    }
    const ll = JSON.parse(l.Log)
    if (!ll.content) {
      continue;
    }
    const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
    syslogTable.row.add([ts, getSeverityHtml(ll.severity), getFacilityName(ll.facility), ll.hostname, ll.content]);
    if(!ctm ) {
      ctm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
      count++;
      continue;
    }
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
    if (ctm != newCtm) {
      let t = new Date(ctm * 60 * 1000);
      data.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
        value: [t,count]
      });
      ctm--;
      for(;ctm > newCtm;ctm--) {
        t = new Date(ctm * 60 * 1000);
        data.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
          value: [t,0]
        });
      }
      count=0;
    }
    count++;
  }
  syslogTable.draw();
  syslogChart.setOption({
    series: [{
      data: data
    }]
  });
  syslogChart.resize();
}

function showTrap(logList) {
  const data = [];
  let count = 0;
  let ctm;
  trapTable.rows().remove();
  for (let i = logList.length - 1; i >= 0; i--) {
    const l = logList[i]
    const ll = JSON.parse(l.Log)
    if (!ll.FromAddress) {
      continue;
    }
    const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
    trapTable.row.add([ts,
      ll.FromAddress,
      ll.GenericTrap, ll.SpecificTrap,
      ll.Enterprise,
      ll.Variables
    ]);
    if(!ctm ) {
      ctm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
      count++;
      continue;
    }
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
    if (ctm != newCtm) {
      let t = new Date(ctm * 60 * 1000);
      data.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
        value: [t,count]
      });
      ctm--;
      for(;ctm > newCtm;ctm--) {
        t = new Date(ctm * 60 * 1000);
        data.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
          value: [t,0]
        });
      }
      count=0;
    }
    count++;
  }
  trapTable.draw();
  trapChart.setOption({
    series: [{
      data: data
    }]
  });
  trapChart.resize(); 
}

function showNetflow(logList) {
  const data = [];
  let count = 0;
  let ctm;
  netflowTable.rows().remove();
  for (let i = logList.length - 1; i >= 0; i--) {
    const l = logList[i]
    const ll = JSON.parse(l.Log)
    if (!ll.srcAddr) {
      continue;
    }
    const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
    netflowTable.row.add([
      ts,
      ll.srcAddr, ll.srcPort,
      ll.dstAddr, ll.dstPort,
      ll.protocolStr, ll.tcpflagsStr,
      ll.packets, ll.bytes, (ll.last - ll.first) / 100.0
    ]);
    if(!ctm ) {
      ctm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
      count++;
      continue;
    }
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
    if (ctm != newCtm) {
      let t = new Date(ctm * 60 * 1000);
      data.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
        value: [t,count]
      });
      ctm--;
      for(;ctm > newCtm;ctm--) {
        t = new Date(ctm * 60 * 1000);
        data.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
          value: [t,0]
        });
      }
      count=0;
    }
    count++;
  }
  netflowTable.draw();
  netflowChart.setOption({
    series: [{
      data: data
    }]
  });
  netflowChart.resize();  
}

function showIpfix(logList) {
  const data = [];
  let count = 0;
  let ctm;
  ipfixTable.rows().remove();
  for (let i = logList.length - 1; i >= 0; i--) {
    const l = logList[i]
    const ll = JSON.parse(l.Log)
    if (!ll.flowStartSysUpTime) {
      continue;
    }
    let srcAddr = ll.sourceIPv4Address || ll.sourceIPv6Address;
    let dstAddr = ll.destinationIPv6Address || ll.destinationIPv4Address;
    const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
    ipfixTable.row.add([
      ts,
      srcAddr, ll.sourceTransportPort,
      dstAddr, ll.destinationTransportPort,
      ll.protocolIdentifier == 6 ? "tcp" : ll.protocolIdentifier == 17 ? "udp" : ll.protocolIdentifier == 1 ? "icmp" : ll.protocolIdentifier,
      ll.tcpControlBits,
      ll.packetDeltaCount, ll.octetDeltaCount,
      (ll.flowEndSysUpTime - ll.flowStartSysUpTime) / 100.0
    ]);
    if(!ctm ) {
      ctm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
      count++;
      continue;
    }
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
    if (ctm != newCtm) {
      let t = new Date(ctm * 60 * 1000);
      data.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
        value: [t,count]
      });
      ctm--;
      for(;ctm > newCtm;ctm--) {
        t = new Date(ctm * 60 * 1000);
        data.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
          value: [t,0]
        });
      }
      count=0;
    }
    count++;
  }
  ipfixTable.draw();
  ipfixChart.setOption({
    series: [{
      data: data
    }]
  });
  ipfixChart.resize();
}

function showPage(mode) {
  const pages = ["polling", "log", "syslog", "trap", "netflow", "ipfix"];
  pages.forEach(p => {
    if (mode == p) {
      $("#" + p + "_page").removeClass("hidden");
      $("#" + p).addClass("active");
    } else {
      $("#" + p + "_page").addClass("hidden");
      $("#" + p).removeClass("active");
    }
  });
  if (mode == "polling") {
    $(".toolbar-footer .log_btns").addClass("hidden");
    $(".toolbar-footer .polling_btns").removeClass("hidden");
  } else {
    $(".toolbar-footer .log_btns").removeClass("hidden");
    $(".toolbar-footer .polling_btns").addClass("hidden");
  }
  currentPage = mode;
}


function setPollingBtns(show) {
  const btns = ["edit", "delete", "poll", "show"];
  btns.forEach(b => {
    if (!show) {
      $('.polling_btns button.' + b).addClass("hidden");
    } else {
      $('.polling_btns button.' + b).removeClass("hidden");
    }
  });
}

function makePollingTable() {
  polling = $('#polling_table').DataTable({
    "paging": false,
    "info": false,
    "ordering": false,
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal": "",
      "emptyTable": "ポーリングがありません。",
      "thousands": "",
      "loadingRecords": "読み込み中...",
      "processing": "処理中...",
      "search": "検索:",
      "zeroRecords": "一致するポーリングがありません。",
      "aria": {
        "sortAscending": ": 昇順でソート",
        "sortDescending": ": 降順でソート"
      }
    },
  });
  $('#polling_table tbody').on('click', 'tr', function () {
    if ($(this).hasClass('selected')) {
      $(this).removeClass('selected');
      setPollingBtns(false);
    } else {
      polling.$('tr.selected').removeClass('selected');
      $(this).addClass('selected');
      setPollingBtns(true);
    }
  });
  $('.polling_btns button.delete').click(function () {
    const r = polling.row('.selected');
    if (!r) {
      return;
    }
    const d = r.data();
    if (!d || d.length < 7) {
      return;
    }
    const id = d[6];
    if (!pollingList[id]) {
      return;
    }
    if (!confirm(`${pollingList[id].Name}を削除しますか?`)) {
      return;
    }
    astilectron.sendMessage({ name: "deletePolling", payload: id }, message => {
      if (message.payload != "ok") {
        astilectron.showErrorBox("ポーリング削除", "削除できません。");
        return;
      }
      r.remove().draw(false);
    });
  });
  $('.polling_btns button.poll').click(function () {
    const r = polling.row('.selected');
    if (!r) {
      return;
    }
    const d = r.data();
    if (!d || d.length < 7) {
      return;
    }
    const id = d[6];
    if (!pollingList[id]) {
      return;
    }
    astilectron.sendMessage({ name: "pollNow", payload: id }, message => {
      if (message.payload != "ok") {
        astilectron.showErrorBox("ポーリング確認", "ポーリングの再実行に失敗しました。");
        return;
      }
      setTimeout(()=> {
        setupPollingPage();
        showPage("polling");
      },1000);
    });
  });
  $('.polling_btns button.edit').click(function () {
    const r = polling.row('.selected');
    if (!r) {
      return;
    }
    const d = r.data();
    if (!d || d.length < 7) {
      return;
    }
    const id = d[6];
    if (!pollingList[id]) {
      return;
    }
    createEditPollingPane(id);
  });
  $('.polling_btns button.add').click(function () {
    createEditPollingPane("");
  });
  $('.polling_btns button.show').click(function () {
    const r = polling.row('.selected');
    if (!r) {
      return;
    }
    const d = r.data();
    if (!d || d.length < 7) {
      return;
    }
    const id = d[6];
    if (!pollingList[id]) {
      return;
    }
    astilectron.sendMessage({ name: "showPolling", payload: id }, message => {
      if (message.payload != "ok") {
        astilectron.showErrorBox("ノード情報", "ログを取得できません。");
        return;
      }
    });
  });
}

function makeLogTables() {
  const logOpt = {
    "paging": true,
    "info": false,
    "order": [[1, "desc"]],
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
      "search": "検索:",
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
  };
  logTable = $('#log_table').DataTable(logOpt);
  logOpt.order = [[0, "desc"]];
  syslogTable = $('#syslog_table').DataTable(logOpt);
  trapTable = $('#trap_table').DataTable(logOpt);
  netflowTable = $('#netflow_table').DataTable(logOpt);
  ipfixTable = $('#ipfix_table').DataTable(logOpt);
}

function makeCharts() {
  const option = {
    title: {
      show: false,
    },
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
      right:"5%",
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      axisLabel:{
        fontSize: "8px",
        formatter: function (value, index) {
          var date = new Date(value);
          return echarts.format.formatTime('MM/dd hh:mm', date)
        }
      },
      splitLine: {
        show: false
      },
    },
    yAxis: {
      type: 'value',
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
  syslogChart = echarts.init(document.getElementById('syslog_chart'));
  syslogChart.setOption(option);
  trapChart = echarts.init(document.getElementById('trap_chart'));
  trapChart.setOption(option);
  netflowChart = echarts.init(document.getElementById('netflow_chart'));
  netflowChart.setOption(option);
  ipfixChart = echarts.init(document.getElementById('ipfix_chart'));
  ipfixChart.setOption(option);
}

function setupTimeVal() {
  $(".log_btns input[name=start]").val(moment().subtract(1, "h").format("Y-MM-DDTHH:00"));
  $(".log_btns input[name=end]").val(moment().add(1,"h").format("Y-MM-DDTHH:00"));
}

document.addEventListener('astilectron-ready', function () {
  makePollingTable();
  makeLogTables();
  makeCharts();
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "show":
        setTimeout(()=>{
          setupPollingPage();
          showPage("polling");
          setupTimeVal();
        },1000);
        return { name: "show", payload: "ok" };
      case "error":
        setTimeout(() => {
          astilectron.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
  $('.toolbar-actions button.close').click(() => {
    astilectron.sendMessage({ name: "close", payload: "" }, message => {
    });
  });
  $('#polling').click(() => {
    if (pane) {
      return true;
    }
    setupPollingPage();
    showPage("polling");
  });
  $('#log').click(() => {
    if (pane) {
      return true;
    }
    showPage("log");
    logChart.resize();
  });
  $('#syslog').click(() => {
    if (pane) {
      return true;
    }
    showPage("syslog");
    syslogChart.resize();
  });
  $('#trap').click(() => {
    if (pane) {
      return true;
    }
    showPage("trap");
    trapChart.resize();
  });
  $('#netflow').click(() => {
    if (pane) {
      return true;
    }
    showPage("netflow");
    netflowChart.resize();
  });
  $('#ipfix').click(() => {
    if (pane) {
      return true;
    }
    showPage("ipfix");
    ipfixChart.resize();
  });
  $('.log_btns button.search').click(function () {
    searchLog();
  });
  const sh = function() {
    return function findMatches(q, cb) {
      let matches, substrRegex;  
      // an array that will be populated with substring matches
      matches = [];
      // regex used to determine if a string contains the substring `q`
      substrRegex = new RegExp(q, 'i');
      // iterate through the pool of strings and for any string that
      // contains the substring `q`, add it to the `matches` array
      $.each(searchHistory, function(i, str) {
        if (substrRegex.test(str)) {
          matches.push(str);
        }
      });  
      cb(matches);
    };
  };
  
  $('.log_btns input[name=filter]').typeahead({
    hint: true,
    highlight: true,
    minLength: 1
  },
  {
    name: 'SearchHistory',
    source: sh()
  });  
});

function createEditPollingPane(id) {
  if (pane) {
    pane.dispose();
    pane = undefined;
  }
  let p;
  if (pollingList[id]) {
    p = pollingList[id];
  } else {
    p = {
      ID: "",
      Name: "",
      NodeID: "",
      Type: "syslog",
      Polling: "",
      Level: "low",
      PollInt: 60,
      Timeout: 1,
      Retry: 1,
      LastTime: 0,
      LastResult: "",
      State: "",
    };
  }
  pane = new Tweakpane({
    title: id === "" ? "ログ監視" : "ログ監視編集",
  });
  pane.addInput(p, 'Name', { label: "名前" });
  pane.addInput(p, 'Type', {
    label: "種別",
    options: {
      "SYSLOG":  "syslog",
      "TRAP":    "trap",
      "NetFlow": "netflow",
      "IPFIX":   "ipfix",
    },
  });
  pane.addInput(p, 'NodeID', {
    label: "関連ノード",
    options: nodes
  });
  pane.addInput(p, 'Level', {
    label: "レベル",
    options: {
      "重度": "high",
      "軽度": "low",
      "警告": "warn",
      "情報": "info",
    },
  });
  pane.addInput(p, 'Polling', { label: "定義" });
  pane.addInput(p, 'PollInt', {
    label: "間隔",
    min: 60,
    max: 3600,
    step: 10,
  });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: 'Save',
  }).on('click', (value) => {
    // Check Values
    if (p.Name == "") {
      astilectron.showErrorBox("ポーリング編集", "名前を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "savePolling", payload: p }, message => {
      if (message.payload !== "ok") {
        astilectron.showErrorBox("ポーリング編集", "保存に失敗しました。");
        return;
      }
      setupPollingPage();
      showPage("polling");
    });
    pane.dispose();
    pane = undefined;
  });
}

function setWindowTitle(n) {
  const t = "ログ表示 - " + n;
  $("title").html(t);
  $("h1.title").html(t);
}
