'use strict';

// ICONS
const iconArray =[
  ["desktop",0xf108],
  ["tablet",0xf3fa],
  ["server",0xf233],
  ["hdd",0xf0a0],
  ["laptop",0xf109],
  ["network-wired",0xf6ff],
  ["wifi",0xf1eb],
  ["cloud",0xf0c2],
  ["print",0xf02f],
  ["sync",0xf021],
  ["mobile-alt",0xf3cd],
];
const iconMap = new Map(iconArray);

// State Colors
const stateColorArray = [
    ["high","#e31a1c"],
    ["low","#fb9a99"],
    ["warn","#dfdf22"],
    ["normal","#33a02c"],
    ["info","#1f78b4"],
    ["repair","#1f78b4"]
];
const  stateColorMap = new Map(stateColorArray);

// State Html
const stateHtmlArray = [
  ["high",'<i class="fas fa-exclamation-circle state state_high"></i>重度'],
  ["low",'<i class="fas fa-exclamation-circle state state_low"></i>軽度'],
  ["warn",'<i class="fas fa-exclamation-triangle state state_warn"></i>注意'],
  ["normal",'<i class="fas fa-check-circle state state_normal"></i>正常'],
  ["info",'<i class="fas fa-info-circle state state_info"></i>情報'],
  ["repair",'<i class="fas fa-check-circle state state_repair"></i>復帰']
];

const  stateHtmlMap = new Map(stateHtmlArray);

function getIcon(icon) {
  const ret = iconMap.get(icon);
  return  ret  ? char(ret) : char(0xf059);
 }

function getStateColor(state) {
  const ret = stateColorMap.get(state);
  return  ret ? color(ret) : color("#999");
}

function getStateHtml(state) {
  const ret = stateHtmlMap.get(state);
  return  ret ? ret : '<i class="fas fa-check-circle state state_unkown"></i>不明';
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
  if( s >= 0 && s < severityHtml.length ){
    return severityHtml[s];
  }
  return  severityHtml[6];
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
  if( f >=0 && f < facilityNames.length){
    return facilityNames[f];
  }
  return  "unkown";
}

const logModeHtml = [
  '<i class="fas fa-stop-circle state state_unknown"></i>しない',
  '<i class="fas fa-video state state_info"></i>常時',
  '<i class="fas fa-ellipsis-h state state_info"></i>変化時',
  '<i class="fas fa-brain state state_high"></i>AI分析',
];

function getLogModeHtml(m) {
  if( m >=0 && m < logModeHtml.length){
    return logModeHtml[m];
  }
  return  logModeHtml[0];
}