'use strict';

let devicesTable;
let usersTable;
let flowsTable;
let serversTable;
let allowRecomendTable;
let allowTable;
let dennyRecomendTable;
let dennyTable;
let currentPage;
let pane;

function showDevices() {
  devicesTable.clear();
  $('#wait').removeClass("hidden");
  astilectron.sendMessage({ name: "getDevices", payload: "" }, message => {
    $('#wait').addClass("hidden");
    let devices =message.payload;
    if (devices == "ng") {
      astilectron.showErrorBox("レポート", "レポートを取得できません。");
      // 表示をクリアするため
      devices = [];
    } else if (devices.length < 1 ) {
      astilectron.showErrorBox("レポート", "該当するデータがありません。");
    }  
    for (let i = 0 ;i < devices.length;i++) {
      const d = devices[i]
      const ft = moment(d.FirstTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const lt = moment(d.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss");
      const score = getScoreHtml(d.Score)
      devicesTable.row.add([score, d.ID, d.Name,d.IP, d.Info, ft,lt,d.ID]);
    }
    devicesTable.draw();
  });
}

function showUsers(){
  usersTable.clear();
  $('#wait').removeClass("hidden");
  astilectron.sendMessage({ name: "getUsers", payload: "" }, message => {
    $('#wait').addClass("hidden");
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
      serversTable.row.add([score, s.Server,s.ServerName,s.Service,s.Loc, ft,lt,s.ID]);
    }
    serversTable.draw();
    $('#wait').addClass("hidden");
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
      flowsTable.row.add([score, f.Client,f.ClientName,f.ClientLoc,f.Server,f.ServerName,f.Service,f.ServerLoc, ft,lt,f.ID]);
    }
    flowsTable.draw();
    $('#wait').addClass("hidden");
  });
}

function showAllow() {
  $('#wait').removeClass("hidden");
  allowTable.clear();
  allowRecomendTable.clear();
  astilectron.sendMessage({ name: "getAllow", payload: "" }, message => {
    let res = message.payload;
    if ( res == "ng" || !res.Rules ) {
      res.Rules = [];
      res.Recomends = [];
    }
    for (let i = 0 ;i < res.Rules.length;i++) {
      const r = res.Rules[i]
      allowTable.row.add([r.Server,r.ServerName,r.Service,r.ID]);
    }
    allowTable.draw();
    $('#wait').addClass("hidden");
  });
}

function showDenny() {
  $('#wait').removeClass("hidden");
  dennyTable.clear();
  dennyRecomendTable.clear();
  astilectron.sendMessage({ name: "getDenny", payload: "" }, message => {
    let res = message.payload;
    if ( res == "ng" || !res.Rules ) {
      res.Rules = [];
      res.Recomends = [];
    }
    for (let i = 0 ;i < res.Rules.length;i++) {
      const r = res.Rules[i]
      dennyTable.row.add([r.Server,r.ServerName,r.Loc,r.Service,r.ID]);
    }
    dennyTable.draw();
    $('#wait').addClass("hidden");
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

function showPage(mode) {
  if(pane) {
    return;
  }
  const pages = ["devices", "users", "servers", "flows","allow","denny"];
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
  if( mode == "allow" || mode == "denny"){
    $("div.report_btns").addClass("hidden");
    $("div.conf_btns").removeClass("hidden");
  } else {
    $("div.conf_btns").addClass("hidden");
    $("div.report_btns").removeClass("hidden");
  }
  setReportBtns(false);
  setRuleAddBtns(false);
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
    case "allow":
      showAllow();
      break;
    case "denny":
      showDenny();
      break;
  }
}

function makeTables() {
  const opt = {
    "paging": true,
    "info": false,
    "pageLength": 25,
    "order": [[0, "desc"]],
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
  
  opt["pageLength"] = 10;
  allowRecomendTable = makeTable('#allow_recomend_table',opt,"addconf");
  allowTable = makeTable('#allow_table',opt,"delconf");
  dennyRecomendTable = makeTable('#denny_recomend_table',opt,"addconf");
  dennyTable = makeTable('#denny_table',opt,"delconf");
}

function makeTable(id,opt,mode){
  const t = $(id).DataTable(opt);
  $(id +' tbody').on('click', 'tr', function () {
    if ($(this).hasClass('selected')) {
      $(this).removeClass('selected');
      if(mode == "report"){
        setReportBtns(false);
      } else if (mode == "addconf") {
        setRuleAddBtns(false);
      } else if (mode == "delconf") {
        setRuleDeleteBtns(false);
      }
    } else {
      t.$('tr.selected').removeClass('selected');
      $(this).addClass('selected');
      if(mode == "report"){
        setReportBtns(true);
      } else if (mode == "addconf") {
        setRuleAddBtns(true);
      } else if (mode == "delconf") {
        setRuleDeleteBtns(true);
      }
    }
  });
  return t
}

document.addEventListener('astilectron-ready', function () {
  makeTables();
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
  $('#allow').click(() => {
    showPage("allow");
  });
  $('#denny').click(() => {
    showPage("denny");
  });

  $('.report_btns button.reset').click(() => {
    resetReportEnt();
  });

  $('.report_btns button.delete').click(() => {
    deleteReportEnt();
  });

  $('.report_btns button.add').click(() => {
    addRuleFromReportEnt();
  });

  $('.conf_btns button.delete').click(() => {
    deleteRule();
  });

  $('.conf_btns button.add').click(() => {
    addRule();
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
  const btns = ["reset","delete","add"];
  btns.forEach( b =>{
    if(!show || (b=="add" && (currentPage=="devices" || currentPage =="users") )) {
      $('.report_btns button.'+ b).addClass("hidden");
    } else {
      $('.report_btns button.'+ b).removeClass("hidden");
    }
  });
}

function setRuleAddBtns(show){
  if(!show) {
    $('.conf_btns button.add').addClass("hidden");
  } else {
    $('.conf_btns button.add').removeClass("hidden");
  }
}

function setRuleDeleteBtns(show){
  if(!show){
    $('.conf_btns button.delete').addClass("hidden");
  } else {
    $('.conf_btns button.delete').removeClass("hidden");
  }
}

function resetReportEnt() {
  let id;
  let cmd;
  let t;
  switch(currentPage) {
    case "devices":
      id  = getSelectedID(devicesTable);
      cmd = "resetDevice";
      t = devicesTable;
      break;
    case "users":
      id  = getSelectedID(usersTable);
      cmd = "resetUser";
      t = usersTable;
      break;
    case "servers":
      id  = getSelectedID(serversTable);
      cmd = "resetServer";
      t = serversTable;
      break;
    case "flows":
      id  = getSelectedID(flowsTable);
      cmd = "resetFlow";
      t = flowsTable;
      break;
  }
  if (!id) {
    return;
  }
  if (!confirm(`レポート${id}をリセットしますか?`)) {
    return;
  }
  astilectron.sendMessage({ name: cmd, payload: id }, message => {
    if (message.payload != "ok" ) {
      astilectron.showErrorBox("レポート", "リセットできません。");
      return;
    }
    const r = t.row('.selected');
    if (r) {
      const d =  r.data();
      d[0] = getScoreHtml(0);
      r.data(d);
    }
  });
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
    Service: d[3],
    Loc: d[4],
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
    Service: d[6],
    Loc: d[7],
  });
}

function addRulePane(e) {
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
  pane.addMonitor(e, 'Service', { label: "サービス",interval:60000 });
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
  let id;
  let cmd;
  let t;
  switch(currentPage) {
    case "allow":
      id  = getSelectedID(allowTable);
      cmd = "deleteAllow";
      t = allowTable;
      break;
    case "denny":
      id  = getSelectedID(dennyTable);
      cmd = "deleteDenny";
      t = dennyTable;
      break;
  }
  if (!id) {
    return;
  }
  if (!confirm(`ルール${id}を削除しますか?`)) {
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

function addRule() {
  let id;
  let cmd;
  let t;
  switch(currentPage) {
    case "allow":
      id  = getSelectedID(allowRecomendTable);
      break;
    case "denny":
      id  = getSelectedID(dennyRecomendTable);
      break;
  }
  if (!id) {
    return;
  }
  astilectron.sendMessage({ name: "addRuleByID", payload: id }, message => {
    if (message.payload != "ok" ) {
      astilectron.showErrorBox("ルール", "追加できません。");
      return;
    }
    setTimeout(()=>{
      showPage(currentPage);
    },100);
  });
}
