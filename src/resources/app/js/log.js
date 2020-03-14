'use strict';

let currentPage = "";
let logTable;
let syslogTable;
let trapTable;
let netflowTable;
let ipfixTable;
let arpTable;
let arpLogTable;
let logChart;
let syslogChart;
let trapChart;
let netflowChart;
let ipfixChart;
let arpLogChart;
const searchHistory = [];

function searchLog() {
  const filter = {
    StartTime: $(".log_btns input[name=start]").val(),
    EndTime: $(".log_btns input[name=end]").val(),
    Filter: $(".log_btns input[name=filter]").val(),
    LogType: currentPage
  }
  astilectron.sendMessage({ name: "searchLog", payload: filter }, message => {
    $('#wait').addClass("hidden");
    if (message.payload == "ng") {
      dialog.showErrorBox("ログ表示", "ログを取得できません。");
      // ログ表示をクリアするため
      message.payload = [];
    } else if (message.payload.length < 1 ) {
      dialog.showErrorBox("ログ表示", "該当するログがありません。");
    } else {
      if(filter.Filter &&  !searchHistory.includes(filter.Filter)){
        searchHistory.push(filter.Filter);
      } 
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
      case "arp":
        showArp(message.payload);
        break;
      default:
        dialog.showErrorBox("ログ表示", "内部エラー表示内容の不整合");
    }
  });
}

function showLog(logList) {
  const data = [];
  let count = 0;
  let ctm;
  logTable.clear();
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
  const dataInfo = [];
  const dataWarn = [];
  const dataError = [];
  let countInfo = 0;
  let countWarn = 0;
  let countError = 0;
  let ctm;
  syslogTable.clear();
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
    syslogTable.row.add([ts, getSeverityHtml(ll.severity), getFacilityName(ll.facility), ll.hostname,ll.tag, ll.content]);
    if(!ctm ) {
      ctm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
      if(ll.severity < 4){
        countError++;
      } else if (ll.severity == 4){
        countWarn++;
      } else {
        countInfo++;
      }
      continue;
    }
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
    if (ctm != newCtm) {
      let t = new Date(ctm * 60 * 1000);
      dataInfo.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
        value: [t,countInfo]
      });
      dataWarn.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
        value: [t,countWarn]
      });
      dataError.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
        value: [t,countError]
      });
      ctm--;
      for(;ctm > newCtm;ctm--) {
        t = new Date(ctm * 60 * 1000);
        dataInfo.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
          value: [t,0]
        });
        dataWarn.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
          value: [t,0]
        });
        dataError.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm', t),
          value: [t,0]
        });
      }
      countInfo=0;
      countWarn=0;
      countError=0;
    }
    if(ll.severity < 4){
      countError++;
    } else if (ll.severity == 4){
      countWarn++;
    } else {
      countInfo++;
    }
  }
  syslogTable.draw();
  syslogChart.setOption({
    series: [
      {
        data: dataInfo
      },
      {
        data: dataWarn
      },
      {
        data: dataError
      },
    ]
  });
  syslogChart.resize();
}

function showTrap(logList) {
  const data = [];
  let count = 0;
  let ctm;
  trapTable.clear();
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
  netflowTable.clear();
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
  ipfixTable.clear();
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

function getArpLogStateHtml(s) {
  if(s == "New"){
    return('<i class="fas fa-plus-circle state state_info"></i>新規');
  }
  if(s == "Change"){
    return('<i class="fas fa-sync state state_high"></i>変化');
  }
  return('<i class="fas fa-check-circle state state_unknown"></i>未定義');
}

function showArp(arpResEnt) {
  const data = [];
  let count = 0;
  let ctm;
  if(!arpResEnt || !arpResEnt.Arps){
    return;
  }
  arpTable.clear();
  for(let i =0;i < arpResEnt.Arps.length;i++ ) {
    arpTable.row.add([arpResEnt.Arps[i].IP, arpResEnt.Arps[i].MAC,arpResEnt.Arps[i].Vendor]);
  }
  arpLogTable.clear();
  for (let i = arpResEnt.Logs.length - 1; i >= 0; i--) {
    const l = arpResEnt.Logs[i]
    if (!l) {
      continue;
    }
    const ll = l.Log.split(',');
    if (ll.length < 2) {
      continue;
    }
    const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
    if(ll.length > 3){
      arpLogTable.row.add([ts, getArpLogStateHtml(ll[0]),  ll[1],ll[3],ll[2]]);
    } else {
      arpLogTable.row.add([ts, getArpLogStateHtml(ll[0]),  ll[1],ll[2],""]);
    }
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
  arpTable.draw();
  arpLogTable.draw();
  arpLogChart.setOption({
    series: [{
      data: data
    }]
  });
  arpLogChart.resize();
}


function showPage(mode) {
  const pages = ["log", "syslog", "trap", "netflow", "ipfix","arp"];
  pages.forEach(p => {
    if (mode == p) {
      $("#" + p + "_page").removeClass("hidden");
      $("#" + p).addClass("active");
    } else {
      $("#" + p + "_page").addClass("hidden");
      $("#" + p).removeClass("active");
    }
  });
  $(".log_btns input[name=filter]").val("");
  currentPage = mode;
  $('.log_btns button.search').click();
}

function makeLogTables() {
  const logOpt = {
    dom: 'lBfrtip',
    buttons: [
      {
        extend:    'copyHtml5',
        text:      '<i class="fas fa-copy"></i>',
        titleAttr: 'Copy'
      },
      {
          extend:    'excelHtml5',
          text:      '<i class="fas fa-file-excel"></i>',
          titleAttr: 'Excel'
      },
      {
          extend:    'csvHtml5',
          text:      '<i class="fas fa-file-csv"></i>',
          titleAttr: 'CSV'
      }
    ],
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
  };
  logTable = $('#log_table').DataTable(logOpt);
  logOpt.order = [[0, "desc"]];
  syslogTable = $('#syslog_table').DataTable(logOpt);
  trapTable = $('#trap_table').DataTable(logOpt);
  netflowTable = $('#netflow_table').DataTable(logOpt);
  ipfixTable = $('#ipfix_table').DataTable(logOpt);
  arpTable = $('#arp_table').DataTable(logOpt);
  arpLogTable = $('#arplog_table').DataTable(logOpt);
}

function makeCharts() {
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
      right:"5%",
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: '日時',
      axisLabel:{
        color:"#ccc",
        fontSize: "8px",
        formatter: function (value, index) {
          var date = new Date(value);
          return echarts.format.formatTime('MM/dd hh:mm', date)
        }
      },
      nameTextStyle:{
        color:"#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle:{
          color: '#ccc'
        }
      },
      splitLine: {
        show: false
      },
    },
    yAxis: {
      type: 'value',
      name: '件数',
      nameTextStyle:{
        color:"#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle:{
          color: '#ccc'
        }
      },
      axisLabel:{
        color:"#ccc",
        fontSize: 8,
        margin: 2,
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
  trapChart = echarts.init(document.getElementById('trap_chart'));
  trapChart.setOption(option);
  netflowChart = echarts.init(document.getElementById('netflow_chart'));
  netflowChart.setOption(option);
  ipfixChart = echarts.init(document.getElementById('ipfix_chart'));
  ipfixChart.setOption(option);
  arpLogChart = echarts.init(document.getElementById('arplog_chart'));
  arpLogChart.setOption(option);
  syslogChart = echarts.init(document.getElementById('syslog_chart'));
  option.series = [
    {
      name: "INFO",
      type: 'bar',
      color: "#1f78b4",
      stack: "count",
      large: true,
      data: [],
    },
    {
      name: "WARN",
      type: 'bar',
      color: "#dfdf22",
      stack: "count",
      large: true,
      data: [],
    },
    {
      name: "ERROR",
      type: 'bar',
      color: "#e31a1c",
      stack: "count",
      large: true,
      data: [],
    },
  ];
  option.legend = {
    textStyle: {
      fontSize: 10,
      color: "#ccc",
    },
    data:["INFO","WARN","ERROR"]
  };
  syslogChart.setOption(option);
}

function setupTimeVal() {
  $(".log_btns input[name=start]").val(moment().subtract(1, "h").format("Y-MM-DDTHH:00"));
  $(".log_btns input[name=end]").val(moment().add(1,"h").format("Y-MM-DDTHH:00"));
}

document.addEventListener('astilectron-ready', function () {
  makeLogTables();
  makeCharts();
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "show":
        setTimeout(()=>{
          setupTimeVal();
          $('#log').click();
        },100);
        return { name: "show", payload: "ok" };
      case "error":
        setTimeout(() => {
          dialog.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
  $('.toolbar-actions button.close').click(() => {
    astilectron.sendMessage({ name: "close", payload: "" }, message => {
    });
  });
  $('#log').click(() => {
    showPage("log");
  });
  $('#syslog').click(() => {
    showPage("syslog");
  });
  $('#trap').click(() => {
    showPage("trap");
  });
  $('#netflow').click(() => {
    showPage("netflow");
  });
  $('#ipfix').click(() => {
    showPage("ipfix");
  });
  $('#arp').click(() => {
    showPage("arp");
  });
  $('.log_btns button.search').click(function () {
    $('#wait').removeClass("hidden");
    setTimeout(()=>{
      searchLog();
    },100);
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
    limit: 200,
    source: sh()
  });  
});
