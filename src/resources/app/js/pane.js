'use strict';

let pane = undefined;

function createMapConfPane() {
  if(pane) {
    return;
  }
  const mapConfTmp = mapConf
  pane = new Tweakpane();

  const f1 = pane.addFolder({
    title: 'マップ設定',
  });
  f1.addInput(mapConfTmp, 'MapName', { label: "名前" });
  f1.addInput(mapConfTmp, 'BackImg', { label: "背景画像" });
  f1.addButton({
    title: '背景画像ファイル選択',
  }).on('click', (value) => {
    dialog.showOpenDialog({ properties: ['openFile'], title: "背景画像ファイル" })
    .then(r => {
      if(r.canceled){
        return;
      }
      const paths = r.filePaths;
      if(paths && paths.length > 0) {
        mapConfTmp.BackImg = paths[0];
      }
      pane.refresh();
    });
  }); 
 
  const f2 = pane.addFolder({
    title: 'ポーリング',
  });
  f2.addInput(mapConfTmp, 'PollInt', { 
    label: "間隔",
    min: 60,
    max: 600,
    step: 10,
  });
  f2.addInput(mapConfTmp, 'Timeout', { 
    label: "Timeout",
    min: 1,
    max: 5,
    step: 1,
  });
  f2.addInput(mapConfTmp, 'Retry', { 
    label: "Retry",
    min: 0,
    max: 5,
    step: 1,
  });
  f2.addInput(mapConfTmp, 'Community', { label: "Community" });
  f2.addInput(mapConfTmp, 'AILevel', { 
    label: "AIレベル",
    options: {
      "重度": "high",
      "軽度": "low",
      "注意": "warn",
      "情報": "info",
    },
  });
  f2.addInput(mapConfTmp, 'AIThreshold', { 
    label: "AI閾値",
    options: {
      "0.01%以下": 88,
      "0.1%以下":  81,
      "1%以下": 74,
    },
  });
  f2.addInput(mapConfTmp, 'ArpWatchLevel', { 
    label: "ARPレベル",
    options: {
      "重度": "high",
      "軽度": "low",
      "注意": "warn",
      "情報": "info",
    },
  });
  const f3 = pane.addFolder({
    title: '受信',
  });
  f3.addInput(mapConfTmp, 'EnableSyslogd', { 
    label: "Syslog",
    options: {
      "Enable": true,
      "Disable": false,
    },
  });
  f3.addInput(mapConfTmp, 'EnableTrapd', { 
    label: "SNMP Trap",
    options: {
      "Enable": true,
      "Disable": false,
    },
  });
  f3.addInput(mapConfTmp, 'EnableNetflowd', { 
    label: "Netflow",
    options: {
      "Enable": true,
      "Disable": false,
    },
  });
  const f4 = pane.addFolder({
    title: 'ログ',
  });
  f4.addInput(mapConfTmp, 'LogDispSize', { 
    label: "表示件数",
    min: 100,
    max: 2000,
    step:100,
 });
  f4.addInput(mapConfTmp, 'LogDays', { 
    label: "保存日数",
    min:1,
    max:365,
    step:1,
  });
  const f5 = pane.addFolder({
    title: 'レポート設定',
  });
  f5.addInput(mapConfTmp, 'GeoIPPath', { label: "GeoIP DB" });
  f5.addButton({
    title: 'ファイル選択',
  }).on('click', (value) => {
    dialog.showOpenDialog({ properties: ['openFile'], title: "GeoIP DB" })
    .then(r => {
      if(r.canceled){
        return;
      }
      const paths = r.filePaths;
      if(paths && paths.length > 0) {
        mapConfTmp.GeoIPPath = paths[0];
      }
      pane.refresh();
    });
  }); 
  f5.addInput(mapConfTmp, 'GrokPath', { label: "抽出設定" });
  f5.addButton({
    title: 'ファイル選択',
  }).on('click', (value) => {
    dialog.showOpenDialog({ properties: ['openFile'], title: "抽出設定ファイル" })
    .then(r => {
      if(r.canceled){
        return;
      }
      const paths = r.filePaths;
      if(paths && paths.length > 0) {
        mapConfTmp.GrokPath = paths[0];
      }
      pane.refresh();
    });
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
    if( mapConfTmp.MapName == "" ){
      dialog.showErrorBox("マップ設定", "マップ名を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "mapConf", payload: mapConfTmp }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("マップ設定", "保存に失敗しました。");
        return;
      }
      mapConf = mapConfTmp;
      setWindowTitle();
      if(mapConf.BackImg ){
        loadImage("./images/backimg",img => {
          backimg =  img;
          redraw();
        });
      } else {
        backimg = undefined;
        redraw();
      }
    });
    pane.dispose();
    pane = undefined;
  });
  return;
}

function createNotifyConfPane() {
  if (pane) {
    return;
  }
  const notifyConfTmp = notifyConf
  pane = new Tweakpane({
    title: "通知設定"
  });
  pane.addInput(notifyConfTmp, 'MailServer', { label: "サーバー" });
  pane.addInput(notifyConfTmp, 'User', { label: "ユーザー" });
  pane.addInput(notifyConfTmp, 'Password', { label: "パスワード" });
  pane.addInput(notifyConfTmp, 'InsecureSkipVerify', { 
    label: "証明書検証",
    options: {
      "しない": true,
      "する": false,
    },
  });
  pane.addInput(notifyConfTmp, 'MailFrom', { label: "送信元" });
  pane.addInput(notifyConfTmp, 'MailTo', { label: "宛先" });
  pane.addInput(notifyConfTmp, 'Subject', { label: "件名" });
  pane.addInput(notifyConfTmp, 'Interval', { 
    label: "間隔(分)",
    min: 5,
    max: 1440,
    step: 5,
  });
  pane.addInput(notifyConfTmp, 'Level', { 
    label: "レベル",
    options: {
      "通知しない": "none",
      "注意以上": "warn",
      "軽度以上": "low",
      "重度": "high",
    },
  });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: 'Test',
  }).on('click', (value) => {
    astilectron.sendMessage({ name: "notifyTest", payload: notifyConfTmp }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("試験通知", "送信に失敗しました。");
      } else {
        dialog.showErrorBox("通信通知", "送信しました。");
      }
      return
    });
  });
  pane.addButton({
    title: 'Save',
  }).on('click', (value) => {
    if( notifyConfTmp.Subject == "" ){
      dialog.showErrorBox("通知設定", "件名を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "notifyConf", payload: notifyConfTmp }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("通知設定", "保存に失敗しました。");
        return;
      }
      notifyConf = notifyConfTmp;
    });
    pane.dispose();
    pane = undefined;
  });
  return;
}

function createStartDiscoverPane(x,y) {
  if(pane){
    return;
  }
  astilectron.sendMessage({ name: "getDiscover", payload: "" }, message => {
    if(!message.payload.Conf) {
      dialog.showErrorBox("自動発見", "設定を取得できません。");
      return;
    }
    const discoverConf = message.payload.Conf;
    const discoverStat = message.payload.Stat;
    if (discoverStat.Running ){
      createDiscoverStatPane(discoverStat);
      return;
    }
    discoverConf.X= Math.round(x);
    discoverConf.Y= Math.round(y);
    pane = new Tweakpane({
      title: "自動発見"
    });
    pane.addInput(discoverConf, 'StartIP', { label: "開始IP" });
    pane.addInput(discoverConf, 'EndIP', { label: "終了IP" });
    pane.addInput(discoverConf, 'Timeout', { 
      label: "Timeout",
      min: 1,
      max: 5,
      step: 1,
    });
    pane.addInput(discoverConf, 'Retry', { 
      label: "Retry",
      min: 0,
      max: 5,
      step: 1,
    });
    pane.addInput(discoverConf, 'Community', { label: "Community" });
    pane.addButton({
      title: 'Cancel',
    }).on('click', (value) => {
      pane.dispose();
      pane = undefined;
    });
    pane.addButton({
      title: 'Start',
    }).on('click', (value) => {
      // Check Values
      if (discoverConf.StartIP === "" || discoverConf.EndIP === ""  ) {
        dialog.showErrorBox("範囲指定エラー", "開始、終了IPアドレスが正しくありません。")
        return;
      }
      astilectron.sendMessage({ name: "startDiscover", payload: discoverConf }, message => {
        if(message.payload !== "ok") {
          dialog.showErrorBox("自動発見", "開始できません。");
          return;
        }
      });
      pane.dispose();
      pane = undefined;
    });  
  });
}

function createDiscoverStatPane(ds){
  if(pane){
    return;
  }
  let dt = new Date();
  let st = new Date(ds.StartTime/(1000*1000));
  let stats = ds;
  stats.Time = dt.toLocaleTimeString();
  stats.Start = st.toLocaleTimeString();
  stats.End = "";
  // 表示を文字列にする
  stats.Sent = ds.Sent + "";
  stats.Found = ds.Found + "";
  stats.Snmp = ds.Snmp + "";
  stats.Total = ds.Total + "";
  pane = new Tweakpane({
    title: '自動発見の状況',
  });
  pane.addMonitor(stats, 'Start',{
    label: "開始時刻",
    interval: 1000,
  });
  pane.addMonitor(stats, 'Time',{
    label: "現在時刻",
    interval: 1000,
  });
  pane.addMonitor(stats, 'End',{
    label: "終了時刻",
    interval: 1000,
  });
  pane.addMonitor(stats, 'Total',{
    label: "検索総数",
    interval: 1000,
  });
  pane.addMonitor(stats, 'Sent',{
    label: "送信済み",
    interval: 1000,
  });
  pane.addMonitor(stats, 'Progress',{
    label: "完了率",
    interval: 1000,
  });
  pane.addMonitor(stats, 'Found',{
    label: "発見数",
    interval: 1000,
  });
  pane.addMonitor(stats, 'Snmp',{
    label: "SNMP",
    interval: 1000,
  });
  pane.addMonitor(stats, 'Progress', {
    label: "完了率",
    interval: 1000,
    type: 'graph',
    min: 0,
    max: 100,
  });
  pane.addButton({
    title: 'Close',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: 'Stop',
  }).on('click', (value) => {
    astilectron.sendMessage({ name: "stopDiscover", payload: "" }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("自動発見", "停止できません。");
        return;
      }
    });
    pane.dispose();
    pane = undefined;
  });  
  function updateStat() {
    astilectron.sendMessage({ name: "getDiscover", payload: "" }, message => {
      dt = new Date();
      stats.Time = dt.toLocaleTimeString();
      if(message.payload.Stat) {
        const s = message.payload.Stat;
        stats.Sent = s.Sent + "";
        stats.Found = s.Found + "";
        stats.Snmp = s.Snmp + "";
        stats.Progress = s.Progress;
        if (s.EndTime){
          const et = new Date(s.EndTime/(1000*1000));
          stats.End = et.toLocaleTimeString();
        }
        if (!s.Running) {
          setTimeout(()=>{
            dialog.showMessageBox({message: "自動発見完了しました。", title: "自動発見完了"});
            pane.dispose();
            pane = undefined;
          },1500);
          return;
        }
      }
      setTimeout(updateStat,5000);
    });
  }
  updateStat();
}


function createEditNodePane(x,y,nodeID) {
  if(pane){
    return;
  }
  let node;
  if(nodeID != "") {
    node = nodes[nodeID];
  } else {
    node = {
      ID: "",
      Name: "",
      Icon: "desktop",
      Descr: "",
      X: Math.round(x),
      Y: Math.round(y),
      IP: "",
      Community: "",
      URL: "",
    };
  }
  pane = new Tweakpane({
    title: nodeID === "" ? "新規ノード" : "ノード編集"
  });
  pane.addInput(node, 'Name', { label: "名前" });
  pane.addInput(node, 'IP', { label: "IPアドレス" });
  pane.addInput(node, 'Icon', { 
    label: "アイコン",
    options: {
      "デスクトップPC": "desktop",
      "ノートPC": "laptop",
      "タブレット" : "tablet",
      "モバイル": "mobile-alt",
      "サーバー": "server",
      "ルーター": "sync",
      "ネットワーク機器": "hdd",
      "プリンター": "print",
      "有線LAN": "network-wired",
      "無線LAN": "wifi",
      "クラウド": "cloud",
      "TV": "tv",
      "データベース": "database",
      "NTPサーバー": "clock",
      "電話": "phone",
      "ビデオカメラ": "video",
      "地球": "globe",
    },
  });
  pane.addInput(node, 'Community', { label: "Community" });
  pane.addInput(node, 'URL', { label: "URL" });
  pane.addInput(node, 'Descr', { label: "説明" });
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
    if( node.Name == "" ){
      dialog.showErrorBox("ノード編集", "名前を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "saveNode", payload: node }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("ノード編集", "保存に失敗しました。");
        return;
      }
    });
    pane.dispose();
    pane = undefined;
  });
}

function createEditLinePane(nodeID1,nodeID2) {
  if(pane){
    return;
  }
  astilectron.sendMessage({ name: "getLine", payload: {NodeID1:nodeID1,NodeID2:nodeID2} }, message => {
    if(!message.payload) {
      dialog.showErrorBox("ライン編集", "ライン情報を取得できません。");
      return;
    }
    const lineDlg  = message.payload;
    const line = lineDlg.Line;
    pane = new Tweakpane({
      title: lineDlg.Line.ID  === "" ? "新規ライン" : "ライン編集"
    });
    const n1 = pane.addFolder({
      title: lineDlg.NodeName1,
    });
    const opt1 = {};
    lineDlg.Pollings1.forEach( p => {
      opt1[p.Name] = p.ID;
    });
    n1.addInput(lineDlg.Line, 'PollingID1', { 
      label: "ポーリング",
      options: opt1,
    });
    const n2 = pane.addFolder({
      title: lineDlg.NodeName2,
    });
    const opt2 = {};
    lineDlg.Pollings2.forEach( p => {
      opt2[p.Name] = p.ID;
    });
    n2.addInput(line, 'PollingID2', { 
      label: "ポーリング",
      options: opt2,
    });
    pane.addButton({
      title: 'Cancel',
    }).on('click', (value) => {
      pane.dispose();
      pane = undefined;
    });
    if( line.ID != "" ){
      pane.addButton({
        title: 'Delete',
      }).on('click', (value) => {
        astilectron.sendMessage({ name: "deleteLine", payload: line }, message => {
          if(message.payload !== "ok") {
            dialog.showErrorBox("ライン編集", "削除に失敗しました。");
            return;
          }
        });
        pane.dispose();
        pane = undefined;
      });
    }
    pane.addButton({
      title: 'Save',
    }).on('click', (value) => {
      // Check Values
      if( line.PollingID1 === "" || line.PollingID1 === ""  ){
        dialog.showErrorBox("ライン編集", "ポーリングを指定してください。");
        return;
      }
      astilectron.sendMessage({ name: "saveLine", payload: line }, message => {
        if(message.payload !== "ok") {
          dialog.showErrorBox("ライン編集", "保存に失敗しました。");
          return;
        }
      });
      pane.dispose();
      pane = undefined;
    });
  });
}

function createMIBDBPane() {
  if(pane){
    return;
  }
  astilectron.sendMessage({ name: "getMIBModuleList", payload: "" }, message => {
    if(!message.payload) {
      dialog.showErrorBox("MIBデータベース", "リストを取得できません。");
      return;
    }
    const MIBModuleList = {} 
    message.payload.forEach(e => MIBModuleList[e] = e);
    const tmpParams = {
      MIBModule: ""
    };
    pane = new Tweakpane({
      title: "MIBデータベース"
    });
    pane.addButton({
      title: 'MIB追加',
    }).on('click', (value) => {
      dialog.showOpenDialog({ properties: ['openFile'], title: "MIB追加" })
      .then(r => {
        if(r.canceled){
          return;
        }
        const paths = r.filePaths;
        if(paths && paths[0]){
          astilectron.sendMessage({ name: "addMIBFile", payload: paths[0] }, message => {
            if(message.payload !== "ok") {
              dialog.showErrorBox("MIBファイル追加",message.payload);
              return
            }
            pane.dispose();
            pane = undefined;
            setTimeout(createMIBDBPane,100);
            return
          });
        }
      });
    });
    pane.addInput(tmpParams, 'MIBModule', { 
      label: "MIB",
      options: MIBModuleList
    });
    pane.addButton({
      title: 'MIB削除',
    }).on('click', (value) => {
      if(tmpParams.MIBModule){
        astilectron.sendMessage({ name: "delMIBModule", payload: tmpParams.MIBModule }, message => {
          if(message.payload !== "ok") {
            dialog.showErrorBox("MIB削除",message.payload);
            return
          }
          pane.dispose();
          pane = undefined;
          setTimeout(createMIBDBPane,100);
          return
        });
      }
    });
    pane.addButton({
      title: 'Close',
    }).on('click', (value) => {
      pane.dispose();
      pane = undefined;
    });  
  }); 
}

function createActionPane() {
  pane = new Tweakpane({
    title: "操作"
  });
  pane.addButton({
    title: 'ARPリセット...',
  }).on('click', (value) => {
    if (!confirmDialog("ARPリセット","ARP監視をリセットしますか？")){
      return
    }
    astilectron.sendMessage({ name: "resetArpTable", payload: "" }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("ARPリセット", "ARP監視をリセットできません。");
        return;
      }
      pane.dispose();
      pane = undefined;
    });
  });
  pane.addButton({
    title: 'AIモデル削除...',
  }).on('click', (value) => {
    if (!confirmDialog("AIモデルリセット","全てのAIモデルをクリアしますか？")){
      return
    }
    astilectron.sendMessage({ name: "clearAllAIMoldes", payload: "" }, message => {
      pane.dispose();
      pane = undefined;
      return
    });
  });
  pane.addButton({
    title: 'レポートクリア...',
  }).on('click', (value) => {
    if (!confirmDialog("レポートクリア","全てのレポートをクリアしますか？")){
      return
    }
    astilectron.sendMessage({ name: "clearAllReport", payload: "" }, message => {
      pane.dispose();
      pane = undefined;
      return
    });
  });
  pane.addButton({
    title: 'Close',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  return;
}

function createDBStatsPane(){
  if(pane || !dbStats){
    return;
  }
  pane = new Tweakpane({
    title: 'DB統計',
  });
  pane.addMonitor(dbStats, 'Time',{
    label: "更新時刻",
    interval: 30000,
  });
  pane.addMonitor(dbStats, 'Size',{
    label: "サイズ",
    interval: 30000,
  });
  pane.addMonitor(dbStats, 'StartTime',{
    label: "起動時期",
    interval: 30000,
  });
  pane.addMonitor(dbStats, 'TotalWrite',{
    label: "総件数",
    interval: 30000,
  });
  pane.addMonitor(dbStats, 'AvgWrite',{
    label: "平均件数",
    interval: 30000,
  });
  pane.addMonitor(dbStats, 'LastWrite',{
    label: "件数",
    interval: 30000,
  });
  pane.addMonitor(dbStats, 'PeakWrite',{
    label: "最大件数",
    interval: 30000,
  });
  pane.addMonitor(dbStats, 'Rate', {
    label: "傾向",
    type: 'graph',
    min: 0,
    max: 100,
    interval:30000,
  });
  pane.addMonitor(dbStats, 'Speed',{
    label: "速度",
    interval: 30000,
  });
  pane.addMonitor(dbStats, 'Peak',{
    label: "最大速度",
    interval: 30000,
  });
  pane.addButton({
    title: 'Close',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
}
