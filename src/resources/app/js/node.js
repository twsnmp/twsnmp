'use strict';

let nodeID = "";
let currentPage = "";
let basic;
let polling;
let log;
let pane;

let pollingList = {};

function setupBasicPage() {
  astilectron.sendMessage({ name: "getNodeBasicInfo", payload: nodeID }, message => {
    if (!message.payload) {
      astilectron.showErrorBox("ノード情報", "ノード情報を取得できません。");
      return;
    }
    basic.rows().remove();
    const node = message.payload;
    basic.row.add(["名前", node.Name]);
    basic.row.add(["IPアドレス", node.IP]);
    basic.row.add(["状態", getStateHtml(node.State)]);
    basic.row.add(["説明", node.Descr]);
    basic.row.add(["Community", node.Community]);
    basic.draw();
    setWindowTitle(node.Name);
  });
}

function setupPollingPage() {
  astilectron.sendMessage({ name: "getNodePollings", payload: nodeID }, message => {
    if (!message.payload[0].Name) {
      astilectron.showErrorBox("ノード情報", "ポーリングを取得できません。");
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
}

function setupLogPage() {
  astilectron.sendMessage({ name: "getNodeLog", payload: nodeID }, message => {
    if (!message.payload[0].Time) {
      astilectron.showErrorBox("ノード情報", "ログを取得できません。");
      return;
    }
    log.rows().remove();
    for (let i = message.payload.length - 1; i >= 0; i--) {
      const l = message.payload[i]
      const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
      const lvl = getStateHtml(l.Level)
      log.row.add([lvl, ts, l.Type, l.Event]);
    }
    log.draw();
  });
}

function showPage(mode) {
  const pages = ["basic", "polling", "log"];
  pages.forEach(p => {
    if (mode == p) {
      $("#" + p + "_page").removeClass("hidden");
      $(".toolbar-footer ." + p + "_btns").removeClass("hidden");
      $("#" + p).addClass("active");
    } else {
      $("#" + p + "_page").addClass("hidden");
      $(".toolbar-footer ." + p + "_btns").addClass("hidden");
      $("#" + p).removeClass("active");
    }
  });
  currentPage = mode;
}

function makeBasicTable() {
  basic = $('#basic_table').DataTable({
    "paging": false,
    "info": false,
    "ordering": false,
    "searching": false,
    "autoWidth": true,
  });
}

function setPollingBtns(show){
  const btns = ["edit","delete","poll","show"];
  btns.forEach( b =>{
    if(!show) {
      $('.polling_btns button.'+ b).addClass("hidden");
    } else {
      $('.polling_btns button.'+ b).removeClass("hidden");
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
      "decimal":        "",
      "emptyTable":     "ポーリングがありません。",
      "thousands":      "",
      "loadingRecords": "読み込み中...",
      "processing":     "処理中...",
      "search":         "検索:",
      "zeroRecords":    "一致するポーリングがありません。",
      "aria": {
          "sortAscending":  ": 昇順でソート",
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
    if( !r ){
      return;
    }
    const d = r.data();
    if (!d || d.length < 7){
      return;
    }
    const id = d[6];
    if(!pollingList[id]){
      return;
    }
    if (!confirm(`${pollingList[id].Name}を削除しますか?`)) {
      return;
    }
    astilectron.sendMessage({ name: "deletePolling", payload: id }, message => {
      if (message.payload != "ok" ) {
        astilectron.showErrorBox("ポーリング削除", "削除できません。");
        return;
      }
      r.remove().draw(false);     
    });
  });
  $('.polling_btns button.poll').click(function () {
    const r = polling.row('.selected');
    if( !r ){
      return;
    }
    const d = r.data();
    if (!d || d.length < 7){
      return;
    }
    const id = d[6];
    if(!pollingList[id]){
      return;
    }
    astilectron.sendMessage({ name: "pollNow", payload: id }, message => {
      if (message.payload != "ok" ) {
        astilectron.showErrorBox("ポーリング確認", "ポーリングの再実行に失敗しました。");
        return;
      }
      setupPollingPage();
      showPage("polling");
    });
  });
  $('.polling_btns button.edit').click(function () {
    const r = polling.row('.selected');
    if( !r ){
      return;
    }
    const d = r.data();
    if (!d || d.length < 7){
      return;
    }
    const id = d[6];
    if(!pollingList[id]){
      return;
    }
    createEditPollingPane(id);
  });
  $('.polling_btns button.add').click(function () {
    createEditPollingPane("");
  });
  $('.polling_btns button.show').click(function () {
    const r = polling.row('.selected');
    if( !r ){
      return;
    }
    const d = r.data();
    if (!d || d.length < 7){
      return;
    }
    const id = d[6];
    if(!pollingList[id]){
      return;
    }
    astilectron.sendMessage({ name: "showPolling", payload: id }, message => {
      if (message.payload != "ok" ) {
        astilectron.showErrorBox("ノード情報", "ポーリング分析画面を表示できません。");
        return;
      }
    });
  });
}

function makeLogTable() {
  log = $('#log_table').DataTable({
    "paging": true,
    "info": false,
    "order": [[1, "desc"]],
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal":        "",
      "emptyTable":     "表示するログがありません。",
      "info":           "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty":      "",
      "infoFiltered":   "(全 _MAX_ 件)",
      "infoPostFix":    "",
      "thousands":      ",",
      "lengthMenu":     "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing":     "処理中...",
      "search":         "検索:",
      "zeroRecords":    "一致するログがありません。",
      "paginate": {
          "first":      "最初",
          "last":       "最後",
          "next":       "次へ",
          "previous":   "前へ"
      },
      "aria": {
          "sortAscending":  ": 昇順でソート",
          "sortDescending": ": 降順でソート"
      }
    },
  });
}

document.addEventListener('astilectron-ready', function () {
  makeBasicTable();
  makePollingTable();
  makeLogTable();
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "setNodeID":
        if (message.payload) {
          nodeID = message.payload;
          setupBasicPage();
          showPage("basic");
        }
        return { name: "setNodeID", payload: "ok" };
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
  $('#basic').click(() => {
    if(pane){
      return true;
    }
    setupBasicPage();
    showPage("basic");
  });
  $('#polling').click(() => {
    if(pane){
      return true;
    }
    setupPollingPage();
    showPage("polling");
  });
  $('#log').click(() => {
    if(pane){
      return true;
    }
    setupLogPage();
    showPage("log");
  });
});

function createEditPollingPane(id){
  if(pane){
    pane.dispose();
    pane = undefined;
  }
  let p;
  if( pollingList[id] ){
    p = pollingList[id];
  } else {
    p = {
      ID: "",
      Name: "",
      NodeID: nodeID,
      Type: "ping",
      Polling: "",
      Level: "low",
      PollInt:   60,
      Timeout: 1,
      Retry: 1,
      LogMode: 0,
      LastTime: 0,
      LastResult: "",
      State: "unkown",
    };
  }
  pane = new Tweakpane({
    title: id === "" ? "新規ポーリング" : "ポーリング編集",
  });
  pane.addInput(p, 'Name', { label: "名前" });
  pane.addInput(p, 'Type', { 
    label: "種別",
    options: {
      "PING": "ping",
      "SNMP": "snmp",
      "SYSLOG": "syslog",
      "TRAP":   "trap",
      "NetFlow": "netflow",
      "IPFIX":  "ipfix",
    },
  });
  pane.addInput(p, 'Level', { 
    label: "レベル",
    options: {
      "重度": "high",
      "軽度": "low",
      "注意": "warn",
      "情報": "info",
    },
  });
  pane.addInput(p, 'Polling', { label: "定義" });
  pane.addInput(p, 'PollInt', { 
    label: "間隔",
    min: 60,
    max: 600,
    step: 10,
  });
  pane.addInput(p, 'Timeout', { 
    label: "Timeout",
    min: 1,
    max: 5,
    step: 1,
  });
  pane.addInput(p, 'Retry', { 
    label: "Retry",
    min: 0,
    max: 5,
    step: 1,
  });
  pane.addInput(p, 'LogMode', { 
    label: "ログモード",
    options: {
      "記録しない": 0,
      "常に記録": 1,
      "状態変化時のみ記録": 2,
    },
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
    if( p.Name == "" ){
      astilectron.showErrorBox("ポーリング編集", "名前を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "savePolling", payload: p }, message => {
      if(message.payload !== "ok") {
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

function setWindowTitle(n){
  const t = "ノード情報 - " + n;
  $("title").html(t);
  $("h1.title").html(t);
}
