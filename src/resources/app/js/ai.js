'use strict';

document.addEventListener('astilectron-ready', function () {
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "start":
        return { name: "start", payload: "ok" };
      case "error":
        setTimeout(() => {
          astilectron.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
});