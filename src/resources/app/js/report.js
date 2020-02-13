'use strict';

let devicesTable;
let deviceChart;
let usersTable;
let userChart;
let flowsTable;
let flowChart;
let serversTable;
let serverChart;
let rulesTable;
let currentPage;
let pane;

function getServiceNames(services) {
  const sns = new Map();
  for(let i = 0; i < services.length;i++){
    const n = getServiceName(services[i]);
    sns.set(n,true);
  }
  return Array.from(sns.keys()).join();
}

function showDevices() {
  devicesTable.clear();
  $('#wait').removeClass("hidden");
  astilectron.sendMessage({ name: "getDevices", payload: "" }, message => {
    let devices =message.payload;
    if (devices == "ng") {
      astilectron.showErrorBox("レポート", "レポートを取得できません。");
      // 表示をクリアするため
      devices = [];
    } else if (devices.length < 1 ) {
      astilectron.showErrorBox("レポート", "該当するデータがありません。");
    }
    const vendorMap = {};
    for (let i = 0 ;i < devices.length;i++) {
      const d = devices[i]
      const ft = moment(d.FirstTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const lt = moment(d.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const score = getScoreHtml(d.Score)
      devicesTable.row.add([score, d.ID, d.Name,d.IP, d.Vendor, ft,lt,d.ID]);
      if (!vendorMap[d.Vendor]){
        vendorMap[d.Vendor] = [0,0,0,0,0,0,0];
      }
      vendorMap[d.Vendor][0]++;
      const si = getScoreIndex(d.Score);
      vendorMap[d.Vendor][si]++; 
    }
    $('#wait').addClass("hidden");
    devicesTable.draw();
    showDeviceChart(vendorMap);
  });
}

function showUsers(){
  usersTable.clear();
  $('#wait').removeClass("hidden");
  astilectron.sendMessage({ name: "getUsers", payload: "" }, message => {
    let users =message.payload;
    if (users == "ng") {
      astilectron.showErrorBox("レポート", "レポートを取得できません。");
      // 表示をクリアするため
      users = [];
    } else if (users.length < 1 ) {
      astilectron.showErrorBox("レポート", "該当するデータがありません。");
    }  
    for (let i = 0 ;i < users.length;i++) {
      const u = users[i]
      const ft = moment(u.FirstTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const lt = moment(u.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const score = getScoreHtml(u.Score)
      usersTable.row.add([score, u.Name,u.Service, ft,lt,u.ID]);
    }
    $('#wait').addClass("hidden");
    usersTable.draw();
  });
}

function showServers(){
  serversTable.clear();
  $('#wait').removeClass("hidden");
  astilectron.sendMessage({ name: "getServers", payload: "" }, message => {
    let servers =message.payload;
    if (users == "ng") {
      astilectron.showErrorBox("レポート", "レポートを取得できません。");
      // 表示をクリアするため
      servers = [];
    } else if (servers.length < 1 ) {
      astilectron.showErrorBox("レポート", "該当するデータがありません。");
    }
    for (let i = 0 ;i < servers.length;i++) {
      const s = servers[i]
      const ft = moment(s.FirstTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const lt = moment(s.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const score = getScoreHtml(s.Score)
      const services = Object.keys(s.Services);
      serversTable.row.add([score, s.Server,s.ServerName,getServiceNames(services),services.length,
        s.Count,s.Bytes,
        s.Loc, ft,lt,services.join(),s.ID]);
    }
    $('#wait').addClass("hidden");
    serversTable.draw();
    showServerChart();
  });
}

function showFlows() {
  $('#wait').removeClass("hidden");
  flowsTable.clear();
  astilectron.sendMessage({ name: "getFlows", payload: "" }, message => {
    let flows = message.payload;
    if ( flows == "ng") {
      astilectron.showErrorBox("レポート", "レポートを取得できません。");
      // 表示をクリアするため
      flows = [];
    } else if (flows.length < 1 ) {
      astilectron.showErrorBox("レポート", "該当するデータがありません。");
    }
    for (let i = 0 ;i < flows.length;i++) {
      const f = flows[i]
      const ft = moment(f.FirstTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const lt = moment(f.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const score = getScoreHtml(f.Score)
      const services = Object.keys(f.Services)
      flowsTable.row.add([score, 
        f.Client,f.ClientName,f.ClientLoc,
        f.Server,f.ServerName,f.ServerLoc,
        getServiceNames(services),services.length,
        f.Count,f.Bytes,
         ft,lt,services.join(),f.ID]);
    }
    $('#wait').addClass("hidden");
    flowsTable.draw();
    showFlowChart();
  });
}

function showRules() {
  $('#wait').removeClass("hidden");
  rulesTable.clear();
  astilectron.sendMessage({ name: "getRules", payload: "" }, message => {
    let rules = message.payload;
    if ( rules == "ng"  ) {
      rules = [];
    }
    for (let i = 0 ;i < rules.length;i++) {
      const r = rules[i]
      rulesTable.row.add([getRulesTypeHtml(r.Type),r.Server,r.ServerName,r.Loc,r.Service,r.ID]);
    }
    $('#wait').addClass("hidden");
    rulesTable.draw();
  });
}

function getScoreHtml(s) {
  if(s > 66  ){
    return('<i class="fas fa-laugh-beam state state_repair"></i>' + Math.floor(s) );
  } else if (s > 50 ) {
    return('<i class="fas fa-smile-beam state state_info"></i>' + Math.floor(s) );
  } else if (s > 42 ) {
    return('<i class="fas fa-grin-beam-sweat state state_warn"></i>' + Math.floor(s) );
  } else if (s > 33){
    return('<i class="fas fa-sad-tear state state_low"></i>' + Math.floor(s) );
  } else if (s <= 0){
    return('<i class="fas fa-question-circle state state_low"></i>--');
  }
  return('<i class="fas fa-angry state state_high"></i>' + Math.floor(s) );
}

function getScoreIndex(s){
  if(s > 66  ){
    return 5;
  } else if (s > 50 ) {
    return 4;
  } else if (s > 42 ) {
    return 3;
  } else if (s > 33){
    return 2;
  } else if (s <= 0){
    return 6
  }
  return 1;
}

function getRulesTypeHtml(t){
  if (t == "allow") {
    return('<i class="fas fa-check-circle state state_info"></i>許可' );
  }
  return('<i class="fas fa-ban state state_high"></i>禁止');
}

function showPage(mode) {
  if(pane) {
    return;
  }
  const pages = ["devices", "users", "servers", "flows","rules"];
  pages.forEach(p => {
    if (mode == p) {
      $("#" + p + "_page").removeClass("hidden");
      $("#" + p).addClass("active");
    } else {
      $("#" + p + "_page").addClass("hidden");
      $("#" + p).removeClass("active");
    }
  });
  currentPage = mode;
  if( mode == "rules" ){
    $("div.report_btns").addClass("hidden");
    $("div.rules_btns").removeClass("hidden");
  } else {
    $("div.report_btns").removeClass("hidden");
    $("div.rules_btns").addClass("hidden");
  }
  setReportBtns(false);
  setRuleDeleteBtns(false);
  switch (mode) {
    case "devices":
      showDevices();
      break;
    case "users":
      showUsers();
      break;
    case "servers":
      showServers();
      break;
    case "flows":
      showFlows();
      break;
    case "rules":
      showRules();
      break;
  }
}

function makeTables() {
  const opt = {
    "paging": true,
    "info": false,
    "pageLength": 10,
    "order": [[0, "asc"]],
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal": "",
      "emptyTable": "表示する情報がありません。",
      "info": "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty": "",
      "infoFiltered": "(全 _MAX_ 件)",
      "infoPostFix": "",
      "thousands": ",",
      "lengthMenu": "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing": "処理中...",
      "search": "検索:",
      "zeroRecords": "一致する情報がありません。",
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
  devicesTable = makeTable('#devices_table',opt,"report");
  usersTable = makeTable('#users_table',opt,"report");
  flowsTable = makeTable('#flows_table',opt,"report");
  serversTable = makeTable('#servers_table',opt,"report");
  rulesTable = makeTable('#rules_table',opt,"rules");
}

function makeTable(id,opt,mode){
  const t = $(id).DataTable(opt);
  $(id +' tbody').on('click', 'tr', function () {
    if ($(this).hasClass('selected')) {
      $(this).removeClass('selected');
      if(mode == "report"){
        setReportBtns(false);
      } else if (mode == "rules") {
        setRuleDeleteBtns(false);
      }
    } else {
      t.$('tr.selected').removeClass('selected');
      $(this).addClass('selected');
      if(mode == "report"){
        setReportBtns(true);
      } else if (mode == "rules") {
        setRuleDeleteBtns(true);
      }
    }
  });
  return t
}

document.addEventListener('astilectron-ready', function () {
  makeTables();
  makeDeviceChart();
  makeServerChart();
  makeFlowChart();
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "show":
        setTimeout(()=>{
          $('#devices').click();
        },100);
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
  $('#devices').click(() => {
    showPage("devices");
  });
  $('#users').click(() => {
    showPage("users");
  });
  $('#flows').click(() => {
    showPage("flows");
  });
  $('#servers').click(() => {
    showPage("servers");
  });
  $('#rules').click(() => {
    showPage("rules");
  });

  $('.report_btns button.reset').click(() => {
    resetReportEnt();
  });

  $('.report_btns button.refresh').click(() => {
    refreshChart();
  });

  $('.report_btns button.delete').click(() => {
    deleteReportEnt();
  });

  $('.report_btns button.add').click(() => {
    addRuleFromReportEnt();
  });

  $('.report_btns button.showloc').click(() => {
    showLoc();
  });

  $('.rules_btns button.delete').click(() => {
    deleteRule();
  });

});

function getSelectedID(t) {
  const r = t.row('.selected');
  if (!r) {
    return undefined;
  }
  const d = r.data();
  if (!d) {
    return undefined;
  }
  const id = d[d.length-1];
  return id
}

function setReportBtns(show){
  const btns = ["delete","add","showloc"];
  btns.forEach( b =>{
    if(!show || (b=="add" && (currentPage=="devices" || currentPage =="users") )) {
      $('.report_btns button.'+ b).addClass("hidden");
    } else {
      $('.report_btns button.'+ b).removeClass("hidden");
    }
  });
}


function setRuleDeleteBtns(show){
  if(!show){
    $('.rules_btns button.delete').addClass("hidden");
  } else {
    $('.rules_btns button.delete').removeClass("hidden");
  }
}

function resetReportEnt() {
  if (!confirm(`信用スコアを再計算しますか?`)) {
    return;
  }
  $('#wait').removeClass("hidden");
  astilectron.sendMessage({ name:"resetReport", payload: currentPage }, message => {
    $('#wait').addClass("hidden");
    if (message.payload != "ok" ) {
      astilectron.showErrorBox("レポート", "信用スコアを再計算できません。");
      return;
    }
    setTimeout(function(){
      showPage(currentPage);
    },100);
  });
}

function refreshChart() {
  switch (currentPage){
    case "servers":
      showServerChart();
    case "flows":
      showFlowChart();
  }
}

function deleteReportEnt() {
  let id;
  let cmd;
  let t;
  switch(currentPage) {
    case "devices":
      id  = getSelectedID(devicesTable);
      cmd = "deleteDevice";
      t = devicesTable;
      break;
    case "users":
      id  = getSelectedID(usersTable);
      cmd = "deleteUser";
      t = usersTable;
      break;
    case "servers":
      id  = getSelectedID(serversTable);
      cmd = "deleteServer";
      t = serversTable;
      break;
    case "flows":
      id  = getSelectedID(flowsTable);
      cmd = "deleteFlow";
      t = flowsTable;
      break;
  }
  if (!id) {
    return;
  }
  if (!confirm(`レポート${id}を削除しますか?`)) {
    return;
  }
  astilectron.sendMessage({ name: cmd, payload: id }, message => {
    if (message.payload != "ok" ) {
      astilectron.showErrorBox("レポート", "削除できません。");
      return;
    }
    const r = t.row('.selected');
    if (r) {
      r.remove().draw(false);
    }
  });
}

function addRuleFromReportEnt() {
  if (pane) {
    return;
  }
  switch(currentPage) {
    case "servers":
      addRuleFromServer();
      break;
    case "flows":
      addRuleFromFlow();
      break;
  }
}

function addRuleFromServer() {
  const r = serversTable.row('.selected');
  if (!r) {
    return;
  }
  const d = r.data();
  addRulePane({
    Type: "allow_service",
    Server: d[1],
    ServerName:d[2],
    Service: d[10],
    Loc: d[7],
  });
}

function addRuleFromFlow() {
  const r = flowsTable.row('.selected');
  if (!r) {
    return;
  }
  const d = r.data();
  addRulePane({
    Type: "denny_service",
    Server: d[4],
    ServerName:d[5],
    Service: d[13],
    Loc: d[6],
  });
}

function addRulePane(e) {
  const a = e.Service.split(",");
  const serviceOpt = {};
  for(let i = 0; i < a.length;i++){
    serviceOpt[a[i]] = a[i];
  }
  if(a.length>0){
    e.Service = a[0];
  }
  pane = new Tweakpane({
    title: "新規ルール"
  });
  pane.addInput(e, 'Type', { 
    label: "種別",
    options: {
      "サーバー限定サービス": "allow_service",
      "禁止サービス" : "denny_service",
      "禁止サーバー" : "denny_server",
      "禁止サーバー&サービス" : "denny_server_service",
      "禁止サーバー位置": "denny_loc",
      "禁止サーバー位置&サービス": "denny_service_loc",
    },
  });
  pane.addMonitor(e, 'Server', { label: "サーバー",interval:60000 });
  pane.addMonitor(e, 'ServerName', { label: "サーバー名",interval:60000 });
  pane.addInput(e, 'Service', { label: "サービス",options: serviceOpt});
  pane.addMonitor(e, 'Loc', { label: "サーバー位置",interval:60000 });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: 'Save',
  }).on('click', (value) => {
    astilectron.sendMessage({ name: "addRule", payload: e }, message => {
      if(message.payload !== "ok") {
        astilectron.showErrorBox("新規ルール", "追加に失敗しました。");
        return;
      }
    });
    pane.dispose();
    pane = undefined;
  });
}

function deleteRule() {
  const id  = getSelectedID(rulesTable);
  if (!id) {
    return;
  }
  if (!confirm(`ルール${id}を削除しますか?`)) {
    return;
  }
  astilectron.sendMessage({ name: "deleteRules", payload: id }, message => {
    if (message.payload != "ok" ) {
      astilectron.showErrorBox("レポート", "ルールを削除できません。");
      return;
    }
    const r = rulesTable.row('.selected');
    if (r) {
      r.remove().draw(false);
    }
  });
}

function showLoc() {
  switch(currentPage) {
    case "servers":
      showLocServer();
      break;
    case "flows":
      showLocFlow();
      break;
  }
}

function showLocServer() {
  const r = serversTable.row('.selected');
  if (!r) {
    return;
  }
  const d = r.data();
  if( !d || d.length < 10){
    return;
  }
  const loc = d[7].split(",");
  if(loc.length < 4|| loc[0] == "LOCAL"){
    return;
  }
  sendShowLoc(loc[1],loc[2]);
}

function showLocFlow() {
  const r = flowsTable.row('.selected');
  if (!r) {
    return;
  }
  const d = r.data();
  if( !d || d.length < 10){
    return;
  }
  let loc = d[6].split(",");
  if(loc.length < 4|| loc[0] == "LOCAL"){
    loc = d[3].split(",");
    if(loc.length < 4|| loc[0] == "LOCAL"){
      return;
    }
  }
  sendShowLoc(loc[1],loc[2]);
}

function sendShowLoc(lat,long){
  const url = `https://www.google.com/maps/search/?api=1&query=${lat},${long}&zoom=12`;
  astilectron.sendMessage({ name: "showLoc", payload: url }, message => {
    if (message.payload != "ok" ) {
      astilectron.showErrorBox("レポート", "位置を表示できません。");
      return;
    }
  });
}

function  makeDeviceChart(){
  const option = {
      backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [{
        offset: 0,
        color: '#4b5769'
      }, {
        offset: 1,
        color: '#404a59'
      }]),
      tooltip : {
          trigger: 'axis',
          axisPointer : {
              type : 'shadow'
          }
      },
      color:[ "#e31a1c","#fb9a99","#dfdf22","#a6cee3","#1f78b4","#999"],
      legend: {
        orient: "vertical",
        top:   50,
        right: 10,
        textStyle:{
          fontSize: 10,
          color: "#ccc",
        },
        data: ['32以下','33-41','42-50','51-66','67以上','調査中']
      },
      grid: {
          top: '3%',
          left: '7%',
          right: '10%',
          bottom: '3%',
          containLabel: true
      },
      xAxis:  {
          type: 'value',
          name: "台数",
          nameTextStyle:{
            color:"#ccc",
            fontSize: 10,
            margin: 2,
          },
          axisLabel:{
            color:"#ccc",
            fontSize: 10,
            margin: 2,
          },
          axisLine: {
            lineStyle:{
              color: '#ccc'
            }
          }
      },
      yAxis: {
          type: 'category',
          axisLine: {
            show:false,
          },
          axisTick:{
            show:false,
          },
          axisLabel:{
            color:"#ccc",
            fontSize: 8,
            margin: 2,
          },  
          data: []
      },
      series: [
        {
          name: '32以下',
          type: 'bar',
          stack: '台数',
          data: []
        },
        {
          name: '33-41',
          type: 'bar',
          stack: '台数',
          data: []
        },
        {
          name: '42-50',
          type: 'bar',
          stack: '台数',
          data: []
        },
        {
          name: '51-66',
          type: 'bar',
          stack: '台数',
          data: []
        },
        {
          name: '67以上',
          type: 'bar',
          stack: '台数',
          data: []
        },
        {
          name: '調査中',
          type: 'bar',
          stack: '台数',
          data: []
        },
      ],
  };
  deviceChart = echarts.init(document.getElementById('device_chart'));
  deviceChart.setOption(option);
}

function showDeviceChart(data) {
  const opt = {
    yAxis:{
      data:[],
    },
    series:[
      {data:[]},
      {data:[]},
      {data:[]},
      {data:[]},
      {data:[]},
      {data:[]}
    ]
  };
  const keys = Object.keys(data);
  keys.sort(function(a,b){
    return data[b][0] -data[a][0];
  });
  let i = keys.length-1;
  if(i > 49 ){
    i = 49
  }
  for(;i >= 0;i--){
    opt.yAxis.data.push(keys[i]);
    for(let j =0; j < 6;j++){
      opt.series[j].data.push(data[keys[i]][j+1]);
    }
  }
  deviceChart.setOption( opt);
  deviceChart.resize();
}

function  makeServerChart(){
  const option = {
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [{
      offset: 0,
      color: '#4b5769'
    }, {
      offset: 1,
      color: '#404a59'
    }]),
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    geo: {
      map: 'world',
      silent: true,
      emphasis: {
          label: {
              show: false,
              areaColor: '#eee'
          }
      },
      itemStyle: {
          borderWidth: 0.2,
          borderColor: '#404a59'
      },
      roam: true
    },
    tooltip: {
      trigger: 'item',
      formatter: function (params) {
        return params.name + ' : ' + params.value[2];
      }
    },
    series: [{
      type: 'scatter',
      coordinateSystem: 'geo',
      label: {
              formatter: '{b}',
              position: 'right',
              color: "#eee",
              show: false
      },
      emphasis: {
              label: {
                  show: true
              }
      },
      symbolSize: 6,
      itemStyle: {
        color: function (params) {
          const s = params.data.value[2];
          if(s > 66  ){
            return "#1f78b4";
          } else if (s > 50 ) {
            return "#a6cee3";
          } else if (s > 42 ) {
            return "#dfdf22";
          } else if (s > 33){
            return "#fb9a99";
          } else if (s <= 0){
            return "#aaa"
          }
          return "#e31a1c";
        }
      },
      data:[]
    }]
  };
  serverChart = echarts.init(document.getElementById('server_chart'));
  serverChart.setOption(option);
}

function getScore(s) {
  const a = s.split("/i>",2);
  if (a.length < 2) {
    return 0;
  }
  return a[1]*1;
}

function showServerChart(data) {
  const opt = {
    series:[
      {data:[]},
    ]
  };
  const locMap = {};
  serversTable.rows({search:"applied"}).every( function ( rowIdx, tableLoop, rowLoop ) {
    if(locMap.length > 10000){
      return;
    }
    const d = this.data();
    if (d[7] == "" || d[7].indexOf("LOCAL") == 0) {
      return;
    }
    const score = getScore(d[0]);
    if(!locMap[d[7]] || locMap[d[7]] > score ) {
      locMap[d[7]] = score
    }
  });
  for(let k in locMap){
    const a = k.split(",")
    if (a.length < 4 || a[0] == "LOCAL" || a[1] == "") {
      continue;
    }
    opt.series[0].data.push({
      name: a[3] + "/" + a[0],
      value: [a[2] * 1.0,a[1] * 1.0,locMap[k]]
    });
  }
  serverChart.setOption(opt);
  serverChart.resize();
}

function  makeFlowChart(){
  const categories =[
    {name:"RU"},
    {name:"CN"},
    {name:"US"},
    {name:"JP"},
    {name:"LOCAL"},
    {name:"Other"}
  ];
  const option = {
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [{
      offset: 0,
      color: '#4b5769'
    }, {
      offset: 1,
      color: '#404a59'
    }]),
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    tooltip: {
      trigger: 'item',
      formatter: function (params) {
        return params.name +":" + params.value;
      }
    },
    legend: [{
      orient: "vertical",
      top:   50,
      right: 20,
      textStyle:{
        fontSize: 10,
        color: "#ccc",
      },
      data:  categories.map(function (a) {
                return a.name;
            })
    }],
    color:[ "#e31a1c","#fb9a99","#dfdf22","#a6cee3","#1f78b4","#999"],
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [
        {
            type: 'graph',
            layout: 'force',
            symbolSize: 6,
            categories: categories,
            roam: true,
            label: {
                show: false
            },
            data: [],
            links: [],
            lineStyle: {
                width: 1,
                curveness: 0
            }
        }
    ]
  };
  flowChart = echarts.init(document.getElementById('flow_chart'));
  flowChart.setOption(option);
}

function getLocCategory(l){
  const a = l.split(",");
  if (a.length< 2) {
    return 0;
  }
  switch (a[0]) {
    case "LOCAL":
      return 4;
    case "JP":
      return 3;
    case "US":
      return 2;
    case "CN":
      return 1;
    case "RU":
      return 0;
  }
  return 5;
}

function getScoreColor(s) {
  if(s.indexOf("repair") != -1  ){
    return "#1f78b4";
  } else if (s.indexOf("info") != -1 ) {
    return "#a6cee3";
  } else if (s.indexOf("warn") != -1 ) {
    return "#dfdf22";
  } else if (s.indexOf("low") != -1){
    return "#fb9a99";
  } else if (s.indexOf("unkown") != -1){
    return "#aaa"
  }
  return "#e31a1c";
}

function showFlowChart() {
  const opt = {
    series:[
      {data:[],
        links:[]
      },
    ]
  };
  const nodes = {};
  flowsTable.rows({search:"applied"}).every( function ( rowIdx, tableLoop, rowLoop ) {
    if (opt.series[0].links.length > 1000) {
      return;
    }
    const d = this.data();
    const c = `${d[2]}(${d[1]})`;
    const s = `${d[5]}(${d[4]})`;
    if( !nodes[s]) {
      nodes[s] = {
        name: s,
        category: getLocCategory(d[6]),
        draggable:true,
        value: d[6]
      }
    }
    if( !nodes[c]) {
      nodes[c] = {
        name: c,
        category: getLocCategory(d[3]),
        draggable:true,
        value: d[3]
      }
    }
    opt.series[0].links.push({
      source: c,
      target: s,
      value: d[7] + ":"+ getScore(d[0]) ,
      lineStyle: {
          color: getScoreColor(d[0])
      }
    });
  });
  for(let k in nodes) {
    opt.series[0].data.push(nodes[k]);
  }
  flowChart.setOption(opt);
  flowChart.resize();
}
