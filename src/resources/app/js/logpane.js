'use strict';

let pane = undefined;

function createLogSearchPane() {
  switch(currentPage){
    case "log":
      createLogPane();
      break;
    case "syslog":
      createSyslogPane();
      break;
    case "trap":
      createTrapPane();
      break;
    case "netflow":
      createNetflowPane();
      break;
    case "ipfix":
      createIpfixPane();
      break;
    case "arp":
      createArpLogPane();
      break;
    default:
      dialog.showErrorBox("ログ表示", "内部エラー検索画面の不整合");
  }
}

function doSearchLog(f) {
  $(".log_btns input[name=filter]").val("`"+f+"`");
  $('.log_btns button.search').click();
}

function getFilterStr(s){
  return s.trim().replace(/[\\^$.*+?()[\]{}|]/g, '\\$&');
}

function addFilterStr(f,a) {
  if(f != "" ){
    f += ".*";
  }
  return f + a
}

/*
{"Time":1586032536720050000,"Type":"system","Level":"info","NodeName":"","NodeID":"","Event":"TWSNMP終了"},{"Time":1586032563185831000,"Type":"system","Level":"info","NodeName":"","NodeID":"","Event":"TWSNMP起動 データベース='/Users/yamai/Desktop/test.twdb'"}
*/
const logFilter = {
  type: "",
  level: "",
  node: "",
  event:""
}

function createLogPane() {
  if (pane) {
    return;
  }
  pane = new Tweakpane({
    title: "イベントログ検索条件"
  });
  pane.addInput(logFilter, 'type', { 
    label: "種別",
    options: {
      "指定しない": "",
      "システム": "system",
      "ユーザー操作": "user",
      "ポーリング": "polling",
      "AI": "ai"
    },
  });
  pane.addInput(logFilter, 'level', { 
    label: "レベル",
    options: {
      "指定しない": "",
      "注意以上": "warn",
      "軽度以上": "low",
      "重度": "high"
    },
  });
  pane.addInput(logFilter, 'node', { label: "関連ノード" });
  pane.addInput(logFilter, 'event', { label: "イベント" });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: '検索',
  }).on('click', (value) => {
    let f = "";
    //"Type":"system","Level":"info","NodeName":"","NodeID":"","Event"
    if (logFilter.type != "") {
      f += `"Type":"${logFilter.type}",`;
    }
    switch(logFilter.level) {
      case "warn":
        f += `"Level":"(warn|low|high)",`;
        break;
      case "low":
        f += `"Level":"(low|high)",`;
        break;
      case "high":
        f += `"Level":"high",`;
        break;
    }
    let s = getFilterStr(logFilter.node);
    if(s != ""){
      f = addFilterStr(f,`"NodeName":"${s}.*"`);
    }
    s = getFilterStr(logFilter.event);
    if(s != ""){
      f = addFilterStr(f,`"Event":".*${s}.*"`);
    }
    pane.dispose();
    pane = undefined;
    doSearchLog(f);
  });
  return;
}

/*
{"Time":1586032993272943000,"Type":"syslog",
"Log":"{\"client\":\"192.168.1.201:52141\",\"content\":\"Connection from UDP: [192.168.1.5]:52940-\\u003e[192.168.1.201]:161\",\"facility\":3,\"hostname\":\"rpi\",\"priority\":30,\"severity\":6,\"tag\":\"snmpd\",\"timestamp\":\"2020-04-05T05:43:13Z\",\"tls_peer\":\"\"}"},
*/
const syslogFilter = {
  severity: "",
  facility: "",
  hostname: "",
  tag:      "",
  content:  ""
}

function createSyslogPane() {
  if (pane) {
    return;
  }
  pane = new Tweakpane({
    title: "syslog検索条件"
  });
  pane.addInput(syslogFilter, 'severity', { 
    label: "Severity",
    options: {
      "指定しない": "",
      "err以上": "err",
      "warn以上": "warn",
      "info以上": "info",
    },
  });
  pane.addInput(syslogFilter, 'facility', { 
    label: "Facility",
    options: {
      "指定しない": "",
      "kern": "0",
      "user": "1",
      "mail": "2",
      "daemon": "3",
      "auth": "4",
      "syslog": "5",
      "lpr": "6",
      "news": "7",
      "uucp": "8",
      "cron": "9",
      "authpriv": "10",
      "ftp": "11",
      "ntp": "12",
      "logaudit":"13",
      "logalert":"14",
      "clock":"15",
      "local0":"16",
      "local1":"17",
      "local2":"18",
      "local3":"19",
      "local4":"20",
      "local5":"21",
      "local6":"22",
      "local7":"23"
    },
  });
  pane.addInput(syslogFilter, 'hostname', { label: "送信元" });
  pane.addInput(syslogFilter, 'tag', { label: "TAG" });
  pane.addInput(syslogFilter, 'content', { label: "ログ" });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: '検索',
  }).on('click', (value) => {
    let f = "";
    // "{\"content\":\"Connection from UDP: [192.168.1.5]:52940-\\u003e[192.168.1.201]:161\",\"facility\":3,\"hostname\":\"rpi\",\"priority\":30,\"severity\":6,\"tag\":\"snmpd\",\"timestamp\":\"2020-04-05T05:43:13Z\",\"tls_peer\":\"\"}"},
    let s = getFilterStr(syslogFilter.content);
    if ( s != "") {
      f += `\\\\"content\\\\":\\\\".*${s}.*\\\\",`;
    }
    if(syslogFilter.facility != "") {
      f = addFilterStr(f,`\\\\"facility\\\\":${syslogFilter.facility},`);
    }
    s = getFilterStr(syslogFilter.hostname);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"hostname\\\\":\\\\"${s}.*\\\\",`);
    }    
    switch(syslogFilter.severity) {
      case "warn":
        f = addFilterStr(f,`\\\\"severity\\\\":(0|1|2|3|4),`);
        break;
      case "err":
        f = addFilterStr(f,`\\\\"severity\\\\":(0|1|2|3),`);
        break;
      case "info":
        f = addFilterStr(f,`\\\\"severity\\\\":(0|1|2|3|4|5|6),`);
        break;
    }
    s = getFilterStr(syslogFilter.tag);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"tag\\\\":\\\\"${s}.*\\\\",`);
    }    
    pane.dispose();
    pane = undefined;
    doSearchLog(f);
  });
  return;
}

/*
{"Time":1586033421887408000,"Type":"trap",
"Log":"{\"Enterprise\":\"\",\"FromAddress\":\"192.168.1.202:47740\",\"GenericTrap\":0,\"SpecificTrap\":0,\"Timestamp\":0,\"Variables\":\"sysUpTimeInstance=11\\nsnmpTrapOID.0=linkDown\\nifIndex.6=6\\nifAdminStatus.6=2\\nifOperStatus.6=2\\nsnmpTrapEnterprise.0=netSnmpAgentOIDs.10\\n\"}"}]}}
*/
const trapFilter = {
  fromAddress: "",
  genericTrap: "",
  snmpTrapOID:"",
  variables: "",
}

function createTrapPane() {
  if (pane) {
    return;
  }
  pane = new Tweakpane({
    title: "SNMP TRAP検索条件"
  });
  pane.addInput(trapFilter, 'fromAddress', { label: "送信元" });
  pane.addInput(trapFilter, 'genericTrap', { 
    label: "GenericTrap",
    options: {
      "指定しない": "",
      "coldStart": "0",
      "warmStart": "1",
      "linkDown": "2",
      "linkUp": "3",
      "authenticationFailure": "4",
      "egpNeighborLoss": "5",
      "enterpriseSpecific": "6",
    },
  });
  pane.addInput(trapFilter, 'snmpTrapOID', { label: "SNMP TRAP OID" });
  pane.addInput(trapFilter, 'variables', { label: "Variables" });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: '検索',
  }).on('click', (value) => {
    let f = "";
    //\"FromAddress\":\"192.168.1.202:47740\",\"GenericTrap\":0,\"SpecificTrap\":0,\"Timestamp\":0,\"Variables\":\"sysUpTimeInstance=11\\nsnmpTrapOID.0=linkDown\\nifIndex.6=6\\nifAdminStatus.6=2\\nifOperStatus.6=2\\nsnmpTrapEnterprise.0=netSnmpAgentOIDs.10\\n\"}"}]}}
    let s = getFilterStr(trapFilter.fromAddress);
    if ( s != "") {
      f += `\\\\"FromAddress\\\\":\\\\"${s}:.*\\\\",`;
    }
    if(trapFilter.genericTrap != "") {
      f = addFilterStr(f,`\\\\"GenericTrap\\\\":${trapFilter.genericTrap},`);
    }
    s = getFilterStr(trapFilter.snmpTrapOID);
    if ( s != "") {
      f = addFilterStr(f,`snmpTrapOID\.0=${s}.*\s+`);
    }    
    s = getFilterStr(trapFilter.variables);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"Variables\\\\":\\\\"${s}.*\\\\",`);
    }    
    pane.dispose();
    pane = undefined;
    doSearchLog(f);
  });
  return;
}

/*
{"Time":1586033701005322000,"Type":"netflow",
"Log":"{\"bytes\":832,\"dstAddr\":\"192.168.1.21\",\"dstAs\":0,\"dstMask\":0,\"dstPort\":64397,\"first\":1791239131,\"last\":1791304127,\"nextHop\":\"0.0.0.0\",\"packets\":9,\"protocol\":6,\"protocolStr\":\"tcp\",\"srcAddr\":\"192.168.1.203\",\"srcAs\":0,\"srcMask\":0,\"srcPort\":80,\"tcpflags\":27,\"tcpflagsStr\":\"[FS.PA...]\",\"tos\":0}"}]}}
*/
const netflowFilter = {
  srcAddr: "",
  srcPort: "",
  dstAddr: "",
  dstPort: "",
  protocol:  ""
}

function createNetflowPane() {
  if (pane) {
    return;
  }
  pane = new Tweakpane({
    title: "Netflow検索条件"
  });
  pane.addInput(netflowFilter, 'srcAddr', { label: "送信元IP" });
  pane.addInput(netflowFilter, 'srcPort', { label: "送信元Port" });
  pane.addInput(netflowFilter, 'dstAddr', { label: "宛先IP" });
  pane.addInput(netflowFilter, 'dstPort', { label: "宛先Port" });
  pane.addInput(netflowFilter, 'protocol', { 
    label: "プロトコル",
    options: {
      "指定しない": "",
      "tcp": "6",
      "udp": "17",
      "icmp": "1",
    },
  });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: '検索',
  }).on('click', (value) => {
    let f = "";
    // \"dstAddr\":\"192.168.1.21\",\"dstPort\":64397,\"protocol\":6,\"srcAddr\":\"192.168.1.203\",\"srcPort\":80
    let s = getFilterStr(netflowFilter.dstAddr);
    if ( s != "") {
      f += `\\\\"dstAddr\\\\":\\\\"${s}\\\\",`;
    }
    s = getFilterStr(netflowFilter.dstPort);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"dstPort\\\\":${s},`);
    }    
    s = getFilterStr(netflowFilter.protocol);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"protocol\\\\":${s},`);
    }    
    s = getFilterStr(netflowFilter.srcAddr);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"srcAddr\\\\":\\\\"${s}\\\\",`);
    }
    s = getFilterStr(netflowFilter.srcPort);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"srcPort\\\\":${s},`);
    }   
    pane.dispose();
    pane = undefined;
    doSearchLog(f);
  });
  return;
}


/*
{"Time":1586033941009158000,"Type":"ipfix",
"Log":"{\"destinationIPv4Address\":\"192.168.1.21\",\"destinationTransportPort\":53639,\"egressInterface\":0,\"flowEndSysUpTime\":1634748334,\"flowStartSysUpTime\":1634748333,\"icmpTypeCodeIPv4\":0,\"ingressInterface\":0,\"ipClassOfService\":0,\"ipVersion\":4,\"octetDeltaCount\":75,\"packetDeltaCount\":1,\"protocolIdentifier\":17,\"sourceIPv4Address\":\"192.168.1.203\",\"sourceTransportPort\":161,\"tcpControlBits\":0,\"vlanId\":0}"}]}}
*/
const ipfixFilter = {
  srcAddr: "",
  srcPort: "",
  dstAddr: "",
  dstPort: "",
  protocol:  ""
}

function createIpfixPane() {
  if (pane) {
    return;
  }
  pane = new Tweakpane({
    title: "IPFIX検索条件"
  });
  pane.addInput(ipfixFilter, 'srcAddr', { label: "送信元IP" });
  pane.addInput(ipfixFilter, 'srcPort', { label: "送信元Port" });
  pane.addInput(ipfixFilter, 'dstAddr', { label: "宛先IP" });
  pane.addInput(ipfixFilter, 'dstPort', { label: "宛先Port" });
  pane.addInput(ipfixFilter, 'protocol', { 
    label: "プロトコル",
    options: {
      "指定しない": "",
      "tcp": "6",
      "udp": "17",
      "icmp": "1",
    },
  });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: '検索',
  }).on('click', (value) => {
    let f = "";
    // \"destinationIPv4Address\":\"192.168.1.21\",\"destinationTransportPort\":53639,\"protocolIdentifier\":17,\"sourceIPv4Address\":\"192.168.1.203\",\"sourceTransportPort\":161
    let s = getFilterStr(ipfixFilter.dstAddr);
    if ( s != "") {
      f += `\\\\"destinationIPv4Address\\\\":\\\\"${s}\\\\",`;
    }
    s = getFilterStr(ipfixFilter.dstPort);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"destinationTransportPort\\\\":${s},`);
    }    
    s = getFilterStr(ipfixFilter.protocol);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"protocolIdentifier\\\\":${s},`);
    }    
    s = getFilterStr(ipfixFilter.srcAddr);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"sourceIPv4Address\\\\":\\\\"${s}\\\\",`);
    }
    s = getFilterStr(ipfixFilter.srcPort);
    if ( s != "") {
      f = addFilterStr(f,`\\\\"sourceTransportPort\\\\":${s},`);
    }   
    pane.dispose();
    pane = undefined;
    doSearchLog(f);
  });
  return;
}

/*
{"Time":1585999852969241000,"Type":"arplog",
"Log":"Change,192.168.1.8,08:E6:89:25:31:CB,84:AF:EC:F1:88:D0"}
*/
const arpLogFilter = {
  type: "",
  ip: "",
  oldMac: "",
  mac: ""
}

function createArpLogPane() {
  if (pane) {
    return;
  }
  pane = new Tweakpane({
    title: "Arpログ検索条件"
  });
  pane.addInput(arpLogFilter, 'type', { 
    label: "種別",
    options: {
      "指定しない": "",
      "新規": "New",
      "変化": "Change",
    },
  });
  pane.addInput(arpLogFilter, 'ip', { label: "IPアドレス"});
  pane.addInput(arpLogFilter, 'mac', { label: "MACアドレス" });
  pane.addInput(arpLogFilter, 'oldMac', { label: "前のMACアドレス" });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: '検索',
  }).on('click', (value) => {
    let f = "";
    // "Log":"Change,192.168.1.8,08:E6:89:25:31:CB,84:AF:EC:F1:88:D0"}
    let s = getFilterStr(arpLogFilter.type);
    if ( s != "") {
      f += `"Log":"${s},`;
    } else {
      f += `"Log":".+,`;
    }
    s = getFilterStr(arpLogFilter.ip);
    if ( s != "") {
      f += `${s},`;
    } else {
      f += `.*,`;
    }
    s = getFilterStr(arpLogFilter.oldMac);
    if ( s != "") {
      f += `${s},`;
    } else {
      f += `.*,`;
    }
    s = getFilterStr(arpLogFilter.mac);
    if ( s != "") {
      f += `${s}`;
    }    
    pane.dispose();
    pane = undefined;
    doSearchLog(f);
  });
  return;
}

