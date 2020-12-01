'use strict';
const { dialog } = require('electron').remote
// ICONS
const iconArray = [
  ["desktop", 0xf108],
  ["tablet", 0xf3fa],
  ["server", 0xf233],
  ["hdd", 0xf0a0],
  ["laptop", 0xf109],
  ["network-wired", 0xf6ff],
  ["wifi", 0xf1eb],
  ["cloud", 0xf0c2],
  ["print", 0xf02f],
  ["sync", 0xf021],
  ["mobile-alt", 0xf3cd],
  ["tv", 0xf26c],
  ["database", 0xf1c0],
  ["clock", 0xf017],
  ["phone", 0xf095],
  ["video", 0xf03d],
  ["globe", 0xf0ac],
];
const iconMap = new Map(iconArray);

// State Colors
const stateColorArray = [
  ["high", "#e31a1c"],
  ["low", "#fb9a99"],
  ["warn", "#dfdf22"],
  ["normal", "#33a02c"],
  ["info", "#1f78b4"],
  ["repair", "#1f78b4"]
];
const stateColorMap = new Map(stateColorArray);

// State Html
const stateHtmlArray = [
  ["high", '<i class="fas fa-exclamation-circle state state_high"></i>重度'],
  ["low", '<i class="fas fa-exclamation-circle state state_low"></i>軽度'],
  ["warn", '<i class="fas fa-exclamation-triangle state state_warn"></i>注意'],
  ["normal", '<i class="fas fa-check-circle state state_normal"></i>正常'],
  ["info", '<i class="fas fa-info-circle state state_info"></i>情報'],
  ["repair", '<i class="fas fa-check-circle state state_repair"></i>復帰']
];

const stateHtmlMap = new Map(stateHtmlArray);

// Service Name Map
const serviceNameArray = [
  ["submission/tcp", "SMTP"],
  ["http/tcp", "HTTP"],
  ["https/tcp", "HTTPS"],
  ["ldap/tcp", "LDAP"],
  ["ldaps/tcp", "LDAPS"],
  ["domain/tcp", "DNS"],
  ["domain/udp", "DNS"],
  ["snmp/udp", "SNMP"],
  ["ntp/udp", "NTP"],
  ["smtp/tcp", "SMTP"],
  ["pop3/tcp", "POP3"],
  ["pop3s/tcp", "POP3S"],
  ["imap/tcp", "IMAP"],
  ["imaps/tcp", "IMAPS"],
  ["ssh/tcp", "SSH"],
  ["telnet/tcp", "TELNET"],
  ["ftp/tcp", "FTP"],
  ["bootps/udp", "DHCP"],
  ["syslog/udp", "SYSLOG"],
  ["microsoft-ds/tcp", "CIFS"],
  ["rfb/tcp", "RFB"],
  ["netbios-ns/udp", "NETBIOS"],
  ["netbios-dgm/udp", "NETBIOS"],
  ["icmp", "ICMP"],
  ["igmp", "IGMP"]
];

const serviceNameMap = new Map(serviceNameArray);


function getIcon(icon) {
  const ret = iconMap.get(icon);
  return ret ? char(ret) : char(0xf059);
}

function getStateColor(state) {
  const ret = stateColorMap.get(state);
  return ret ? color(ret) : color("#999");
}

function getStateHtml(state) {
  const ret = stateHtmlMap.get(state);
  return ret ? ret : '<i class="fas fa-check-circle state state_unknown"></i>不明';
}

function getServiceName(s) {
  const ret = serviceNameMap.get(s);
  return ret ? ret : 'Other';
}

// severity Html
const severityHtml = [
  '<i class="fas fa-exclamation-circle state state_high"></i>emerg',
  '<i class="fas fa-exclamation-circle state state_high"></i>alert',
  '<i class="fas fa-exclamation-triangle state state_high"></i>crit',
  '<i class="fas fa-check-circle state state_low"></i>err',
  '<i class="fas fa-info-circle state state_warn"></i>warning',
  '<i class="fas fa-check-circle state state_repair"></i>notice',
  '<i class="fas fa-check-circle state state_info"></i>info',
  '<i class="fas fa-check-circle state state_unknown"></i>debug',
];


function getSeverityHtml(s) {
  if (s >= 0 && s < severityHtml.length) {
    return severityHtml[s];
  }
  return severityHtml[6];
}

// Facility Name List
const facilityNames = [
  "kern",
  "user",
  "mail",
  "daemon",
  "auth",
  "syslog",
  "lpr",
  "news",
  "uucp",
  "cron",
  "authpriv",
  "ftp",
  "ntp",
  "logaudit",
  "logalert",
  "clock",
  "local0",
  "local1",
  "local2",
  "local3",
  "local4",
  "local5",
  "local6",
  "local7"
];

function getFacilityName(f) {
  if (f >= 0 && f < facilityNames.length) {
    return facilityNames[f];
  }
  return "unknown";
}

const trapGenericNames = [
  "coldStart",
  "warmStart",
  "linkDown",
  "linkUp",
  "authenticationFailure",
  "egpNeighborLoss",
  "enterpriseSpecific"
];

function getTrapGenericName(g) {
  if (g >= 0 && g < trapGenericNames.length) {
    return trapGenericNames[g];
  }
  return `unknown(${g})`;
}

const logModeHtml = [
  '<i class="fas fa-stop-circle state state_unknown"></i>しない',
  '<i class="fas fa-video state state_info"></i>常時',
  '<i class="fas fa-ellipsis-h state state_info"></i>変化時',
  '<i class="fas fa-brain state state_high"></i>AI分析',
];

function getLogModeHtml(m) {
  if (m >= 0 && m < logModeHtml.length) {
    return logModeHtml[m];
  }
  return logModeHtml[0];
}

const logModeList = {
  "記録しない": 0,
  "常に記録": 1,
  "状態変化時のみ記録": 2,
  "AI分析": 3,
};

const pollingTypeList = {
  "PING": "ping",
  "SNMP": "snmp",
  "TCP": "tcp",
  "HTTP": "http",
  "HTTPS": "https",
  "TLS": "tls",
  "DNS": "dns",
  "NTP": "ntp",
  "SYSLOG": "syslog",
  "SYSLOG PRI": "syslogpri",
  "SYSLOG Device": "syslogdevice",
  "SYSLOG User": "sysloguser",
  "SYSLOG Flow": "syslogflow",
  "TRAP": "trap",
  "NetFlow": "netflow",
  "IPFIX": "ipfix",
  "Command": "cmd",
  "SSH": "ssh",
  "TWSNMP": "twsnmp",
  "VMware": "vmware",
};

const chartDispInfo = {
  "rtt": {
    mul: 1.0 / (1000 * 1000 * 1000),
    axis: "応答時間(秒)"
  },
  "rtt_cv": {
    mul: 1.0,
    axis: "応答時間変動係数"
  },
  "successRate": {
    mul: 100.0,
    axis: "成功率(%)"
  },
  "speed": {
    mul: 1.0,
    axis: "回線速度(Mbps)"
  },
  "speed_cv": {
    mul: 1.0,
    axis: "回線速度変動係数"
  },
  "feels_like": {
    mul: 1.0,
    axis: "体感温度(℃）"
  },
  "humidity": {
    mul: 1.0,
    axis: "湿度(%)"
  },
  "pressure": {
    mul: 1.0,
    axis: "気圧(hPa)"
  },
  "temp": {
    mul: 1.0,
    axis: "温度(℃）"
  },
  "temp_max": {
    mul: 1.0,
    axis: "最高温度(℃）"
  },
  "temp_min": {
    mul: 1.0,
    axis: "最低温度(℃）"
  },
  "wind": {
    mul: 1.0,
    axis: "風速(m/sec)"
  },
  "offset": {
    mul: 1.0 / (1000 * 1000 * 1000),
    axis: "時刻差(秒)"
  },
  "stratum": {
    mul: 1,
    axis: "階層"
  },
  "fail": {
    mul: 1.0,
    axis: "失敗回数"
  }
}

const levelList = {
  "重度": "high",
  "軽度": "low",
  "注意": "warn",
  "情報": "info",
};


function confirmDialog(title, msg) {
  return dialog.showMessageBoxSync(
    { type: "question", title: title, cancelId: 1, message: msg, buttons: ["OK", "Cancel"] }
  ) == 0;
}

function setPasswordInput(pos) {
  $('input.tp-txtiv_i').eq(pos).attr('type', 'password');
}

function setInputError(pos, msg) {
  $('input.tp-txtiv_i').eq(pos).addClass('error');
  $('input.tp-txtiv_i').eq(pos).after(`<p class="error">${msg}</p>`);
}

function clearInputError() {
  $('div.tp-dfwv p.error').remove();
  $('input.tp-txtiv_i').removeClass('error');
}

function setupPanePosAndSize() {
  $('.tp-dfwv').css({
    "position": "absolute",
    "top": "35px",
    "right": "15px",
    "width": "320px",
  });
  $('.tp-lblv_v').css({
    "width": "180px",
  })
}
