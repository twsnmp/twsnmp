'use strict';

let currentPage = "";
let polling;
let template;
let pane;

let pollingList = {};
let templateList = {};
let nodes = {};

function showPage(mode) {
  const pages = ["polling", "template"];
  pages.forEach(p => {
    if (mode == p) {
      $("#" + p + "_page").removeClass("hidden");
      $("div." + p + "_btns").removeClass("hidden");
      $("#" + p).addClass("active");
    } else {
      $("#" + p + "_page").addClass("hidden");
      $("div." + p + "_btns").addClass("hidden");
      $("#" + p).removeClass("active");
    }
  });
  currentPage = mode;
}

function setupPolling() {
  astilectron.sendMessage({ name: "getPollings", payload: "" }, message => {
    if (!message.payload.Pollings) {
      dialog.showErrorBox("ポーリングリスト", "ポーリングを取得できません。");
      return;
    }
    nodes = message.payload.Nodes
    polling.clear();
    pollingList = {};
    for (let i = message.payload.Pollings.length - 1; i >= 0; i--) {
      const p = message.payload.Pollings[i];
      let nodeName = "unkown";
      if (p.NodeID && nodes[p.NodeID] ) {
        nodeName = nodes[p.NodeID].Name;
      }
      const lt = moment(p.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
      const level = getStateHtml(p.Level);
      const state = getStateHtml(p.State);
      const logMode = getLogModeHtml(p.LogMode);
      polling.row.add([state,nodeName, p.Name, level, p.Type,logMode, p.Polling,lt,p.LastVal,p.LastResult, p.ID]);
      pollingList[p.ID] = p;
    }
    polling.draw();
    setPollingBtns(false,false);
  });
}


function setPollingBtns(show,bAn){
  const btns = ["edit","delete","poll","template"];
  btns.forEach( b =>{
    if(!show) {
      $('.polling_btns button.'+ b).addClass("hidden");
    } else {
      $('.polling_btns button.'+ b).removeClass("hidden");
    }
  });
  if (bAn) {
    $('.polling_btns button.show').removeClass("hidden");
  } else {
    $('.polling_btns button.show').addClass("hidden");
  }
}

function setTemplateBtns(show){
  const btns = ["select","edit","delete"];
  btns.forEach( b =>{
    if(!show) {
      $('.template_btns button.'+ b).addClass("hidden");
    } else {
      $('.template_btns button.'+ b).removeClass("hidden");
    }
  });
}


function makePollingTable() {
  polling = $('#polling_table').DataTable({
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
    "order": [[0, "desc"]],
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal": "",
      "emptyTable": "表示するポーリングがありません。",
      "info": "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty": "",
      "infoFiltered": "(全 _MAX_ 件)",
      "infoPostFix": "",
      "thousands": ",",
      "lengthMenu": "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing": "処理中...",
      "search": "フィルター:",
      "zeroRecords": "一致するポーリングがありません。",
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
  });
  $('#polling_table tbody').on('dblclick', 'tr', function () {
    const data = polling.row( this ).data();
    if(data && data.length > 1){
      const id = data[data.length-1];
      if(pollingList[id] && pollingList[id].LogMode ){
        showPolling(id);
      }
    }
  });
  $('#polling_table tbody').on('click', 'tr', function () {
    if ($(this).hasClass('selected')) {
      $(this).removeClass('selected');
      setPollingBtns(false,false);
    } else {
      polling.$('tr.selected').removeClass('selected');
      $(this).addClass('selected');
      const r = polling.row('.selected');
      let bAn = false;
      if( r ){
        const d = r.data();
        if (d){
          const id = d[d.length-1];
          if(pollingList[id] && pollingList[id].LogMode ){
            bAn = true;
          }
        }
      }
      setPollingBtns(true,bAn);
    }
  });
  $('.polling_btns button.delete').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    if (!confirmDialog("ポーリング削除",`${pollingList[id].Name}を削除しますか?`)) {
      return;
    }
    astilectron.sendMessage({ name: "deletePolling", payload: id }, message => {
      if (message.payload != "ok" ) {
        dialog.showErrorBox("ポーリング削除", "削除できません。");
        return;
      }
      const r = polling.row('.selected');
      if (r) {
        r.remove().draw(false);
      }
    });
  });
  $('.polling_btns button.poll').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    astilectron.sendMessage({ name: "pollNow", payload: id }, message => {
      if (message.payload != "ok" ) {
        dialog.showErrorBox("ポーリング確認", "ポーリングの再実行に失敗しました。");
        return;
      }
      setupPolling();
    });
  });
  $('.polling_btns button.edit').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    createEditPollingPane(id,"");
  });
  $('.polling_btns button.add').click(function () {
    createEditPollingPane("","");
  });
  $('.polling_btns button.show').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    showPolling(id);
  });
  $('.polling_btns button.template').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    createEditTemplatePane("",id);
  });
}

function showPolling(id) {
  astilectron.sendMessage({ name: "showPolling", payload: id }, message => {
    if (message.payload != "ok" ) {
      dialog.showErrorBox("ポーリングリスト", "ポーリング分析画面を表示できません。");
      return;
    }
  });
}

function getSelectedPollingID() {
  const r = polling.row('.selected');
  if (!r) {
    return undefined;
  }
  const d = r.data();
  if (!d) {
    return undefined;
  }
  const id = d[d.length-1];
  if (!pollingList[id]) {
    return undefined;
  }
  return id
}

function makeTemplateTable() {
  template = $('#template_table').DataTable({
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
    "order": [[0, "desc"]],
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal": "",
      "emptyTable": "表示するテンプレートがありません。",
      "info": "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty": "",
      "infoFiltered": "(全 _MAX_ 件)",
      "infoPostFix": "",
      "thousands": ",",
      "lengthMenu": "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing": "処理中...",
      "search": "フィルター:",
      "zeroRecords": "一致するテンプレートがありません。",
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
  });
  $('#template_table tbody').on('dblclick', 'tr', function () {
    const data = polling.row( this ).data();
    if(data && data.length > 1){
      const id = data[data.length-1];
      if(id && templateList[id]){
        createEditPollingPane("",id);
      }
    }
  });
  $('#template_table tbody').on('click', 'tr', function () {
    if ($(this).hasClass('selected')) {
      $(this).removeClass('selected');
      setTemplateBtns(false);
    } else {
      template.$('tr.selected').removeClass('selected');
      $(this).addClass('selected');
      setTemplateBtns(true);
    }
  });
  $('.template_btns button.delete').click(function () {
    const id = getSelectedTemplateID()
    if(!id){
      return;
    }
    if (!confirmDialog("テンプレート削除",`${templateList[id].Name}を削除しますか?`)) {
      return;
    }
    astilectron.sendMessage({ name: "deleteTemplate", payload: id }, message => {
      if (message.payload != "ok" ) {
        dialog.showErrorBox("テンプレート削除", "削除できません。");
        return;
      }
      const r = template.row('.selected');
      if (r) {
        r.remove().draw(false);
      }
    });
  });
  $('.template_btns button.edit').click(function () {
    const id = getSelectedTemplateID()
    if(!id){
      return;
    }
    createEditTemplatePane(id,"");
  });
  $('.template_btns button.select').click(function () {
    const id = getSelectedTemplateID()
    if(id && templateList[id]){
      createEditPollingPane("",id);
    }
  });
  $('.template_btns button.add').click(function () {
    createEditTemplatePane("","");
  });
  $('.template_btns button.import').click(function () {
    importTemplate();
  });
  $('.template_btns button.export').click(function () {
    exportTemplate();
  });
}

function importTemplate() {
  dialog.showOpenDialog({ 
    title: "TWSNMPポーリング定義",
    message: "TWSNMPポーリング定義ファイルを選択してください。",
    properties: ['openFile'],
    filters: [
      { name: 'TWSNMPポーリング定義', extensions: ['json'] },
    ]
   }).then(r => {
    if(r.canceled){
      return;
    }
    const paths = r.filePaths;
    if(paths && paths.length > 0) {
      astilectron.sendMessage({ name: "importTemplate", payload: paths[0] }, message => {
        if(message.payload !== "ok") {
          dialog.showErrorBox("インポート", "テンプレートをインポートできません。");
          return;
        }
        showPage("template");
        setupTemplate();
      });
    }
  });
}

function exportTemplate() {
  dialog.showSaveDialog({
    title: "TWSNMPポーリング定義",
    message: "テンプレートファイルを選択してください。",
    defaultPath: "twsnmpPolling",
    showsTagField: false,
    properties: ["createDirectory"],
    filters: [
      { name: 'TWSNMPポーリング定義', extensions: ['json'] },
    ]          
  }).then(r => {
    if(r.canceled || !r.filePath || r.filePath.length < 1 ){
      return;
    }
    astilectron.sendMessage({ name: "exportTemplate", payload: r.filePath }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("エクスポート", "テンプレートをエクスポートできません。");
        return;
      }
    });
  });
}

function getSelectedTemplateID() {
  const r = template.row('.selected');
  if (!r) {
    return undefined;
  }
  const d = r.data();
  if (!d) {
    return undefined;
  }
  const id = d[d.length-1];
  if (!templateList[id]) {
    return undefined;
  }
  return id
}

function setupTemplate() {
  astilectron.sendMessage({ name: "getTemplates", payload: "" }, message => {
    if (!message.payload) {
      dialog.showErrorBox("ポーリングリスト", "テンプレートを取得できません。");
      return;
    }
    template.clear();
    templateList =  message.payload;
    for(let id in templateList){
      const t = templateList[id];
      const level = getStateHtml(t.Level);
      template.row.add([level,t.Type,t.Name,t.Polling,t.NodeType,t.Descr,t.ID]);
    }
    template.draw();
    setTemplateBtns(false);
  });
}

document.addEventListener('astilectron-ready', function () {
  makePollingTable();
  makeTemplateTable();
  showPage("polling");
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "show":
        setupPolling()
        setupTemplate()
        return { name: "show", payload: "ok" };
      case "error":
        setTimeout(() => {
          dialog.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
  $("#polling").click(() => {
    showPage("polling");
  });
  $("#template").click(() => {
    showPage("template");
  });
  $('.toolbar-actions button.close').click(() => {
    astilectron.sendMessage({ name: "close", payload: "" }, message => {
    });
  });
});

function createEditPollingPane(id,tid){
  if(pane){
    return;
  }
  let p;
  if( pollingList[id] ){
    p = pollingList[id];
  } else {
    p = {
      ID: "",
      Name: "",
      NodeID: "",
      Type: "ping",
      Polling: "",
      Level: "low",
      PollInt:   60,
      Timeout: 1,
      Retry: 1,
      NextTime:0,
      DispMode:"",
      LogMode: 0,
      LastTime: 0,
      LastResult: "",
      State: "unkown",
    };
    if(tid && templateList[tid]) {
      p.Type = templateList[tid].Type;
      p.Level = templateList[tid].Level;
      p.Name = templateList[tid].Name;
      p.Polling = templateList[tid].Polling;
    }
  }
  pane = new Tweakpane({
    title: id === "" ? "新規ポーリング" : "ポーリング編集",
  });
  if(id == "") {
    const opts = {};
    for(let k in nodes){
      const e = nodes[k];
      if(e){
        opts[e.Name] = e.NodeID;
      }
    }
    pane.addInput(p, 'NodeID', { 
      label: "ノード",
      options: opts
    });
  }
  pane.addInput(p, 'Name', { label: "名前" });
  pane.addInput(p, 'Type', { 
    label: "種別",
    options: pollingTypeList,
  });
  pane.addInput(p, 'Level', { 
    label: "レベル",
    options: levelList,
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
    options: logModeList,
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
      dialog.showErrorBox("ポーリング編集", "名前を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "savePolling", payload: p }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("ポーリング編集", "保存に失敗しました。");
        return;
      }
      showPage("polling");
      setupPolling();
    });
    pane.dispose();
    pane = undefined;
  });
}

function createEditTemplatePane(id,pid){
  if(pane){
    return;
  }
  let pt;
  if( templateList[id] ){
    pt = templateList[id];
  } else {
    pt = {
      ID: "",
      Name: "",
      Level:"low",
      Type: "ping",
      Polling: "",
      NodeType: "",
      Descr: "",
    };
    if(pid && pollingList[pid]) {
      pt.Name = pollingList[pid].Name;
      pt.Type = pollingList[pid].Type;
      pt.Level = pollingList[pid].Level;
      pt.Polling = pollingList[pid].Polling;
    } 
  }
  pane = new Tweakpane({
    title: id === "" ? "新規テンプレート" : "テンプレート編集",
  });
  pane.addInput(pt, 'Name', { label: "名前" });
  pane.addInput(pt, 'Level', { 
    label: "レベル",
    options: levelList,
  });
  pane.addInput(pt, 'Type', { 
    label: "種別",
    options: pollingTypeList,
  });
  pane.addInput(pt, 'Polling', { label: "定義" });
  pane.addInput(pt, 'NodeType', { label: "ノード種別" });
  pane.addInput(pt, 'Descr', { label: "説明" });
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
    if( pt.Name == "" ){
      dialog.showErrorBox("テンプレート編集", "名前を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "saveTemplate", payload: pt }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("テンプレート編集", "保存に失敗しました。");
        return;
      }
      showPage("template");
      setupTemplate();
    });
    pane.dispose();
    pane = undefined;
  });
}


