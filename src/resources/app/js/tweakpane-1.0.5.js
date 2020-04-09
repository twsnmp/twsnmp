(function webpackUniversalModuleDefinition(root, factory) {
	if(typeof exports === 'object' && typeof module === 'object')
		module.exports = factory();
	else if(typeof define === 'function' && define.amd)
		define([], factory);
	else if(typeof exports === 'object')
		exports["tweakpane"] = factory();
	else
		root["Tweakpane"] = factory();
})(typeof self !== 'undefined' ? self : this, function() {
return /******/ (function(modules) { // webpackBootstrap
/******/ 	// The module cache
/******/ 	var installedModules = {};
/******/
/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {
/******/
/******/ 		// Check if module is in cache
/******/ 		if(installedModules[moduleId]) {
/******/ 			return installedModules[moduleId].exports;
/******/ 		}
/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = installedModules[moduleId] = {
/******/ 			i: moduleId,
/******/ 			l: false,
/******/ 			exports: {}
/******/ 		};
/******/
/******/ 		// Execute the module function
/******/ 		modules[moduleId].call(module.exports, module, module.exports, __webpack_require__);
/******/
/******/ 		// Flag the module as loaded
/******/ 		module.l = true;
/******/
/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}
/******/
/******/
/******/ 	// expose the modules object (__webpack_modules__)
/******/ 	__webpack_require__.m = modules;
/******/
/******/ 	// expose the module cache
/******/ 	__webpack_require__.c = installedModules;
/******/
/******/ 	// define getter function for harmony exports
/******/ 	__webpack_require__.d = function(exports, name, getter) {
/******/ 		if(!__webpack_require__.o(exports, name)) {
/******/ 			Object.defineProperty(exports, name, { enumerable: true, get: getter });
/******/ 		}
/******/ 	};
/******/
/******/ 	// define __esModule on exports
/******/ 	__webpack_require__.r = function(exports) {
/******/ 		if(typeof Symbol !== 'undefined' && Symbol.toStringTag) {
/******/ 			Object.defineProperty(exports, Symbol.toStringTag, { value: 'Module' });
/******/ 		}
/******/ 		Object.defineProperty(exports, '__esModule', { value: true });
/******/ 	};
/******/
/******/ 	// create a fake namespace object
/******/ 	// mode & 1: value is a module id, require it
/******/ 	// mode & 2: merge all properties of value into the ns
/******/ 	// mode & 4: return value when already ns object
/******/ 	// mode & 8|1: behave like require
/******/ 	__webpack_require__.t = function(value, mode) {
/******/ 		if(mode & 1) value = __webpack_require__(value);
/******/ 		if(mode & 8) return value;
/******/ 		if((mode & 4) && typeof value === 'object' && value && value.__esModule) return value;
/******/ 		var ns = Object.create(null);
/******/ 		__webpack_require__.r(ns);
/******/ 		Object.defineProperty(ns, 'default', { enumerable: true, value: value });
/******/ 		if(mode & 2 && typeof value != 'string') for(var key in value) __webpack_require__.d(ns, key, function(key) { return value[key]; }.bind(null, key));
/******/ 		return ns;
/******/ 	};
/******/
/******/ 	// getDefaultExport function for compatibility with non-harmony modules
/******/ 	__webpack_require__.n = function(module) {
/******/ 		var getter = module && module.__esModule ?
/******/ 			function getDefault() { return module['default']; } :
/******/ 			function getModuleExports() { return module; };
/******/ 		__webpack_require__.d(getter, 'a', getter);
/******/ 		return getter;
/******/ 	};
/******/
/******/ 	// Object.prototype.hasOwnProperty.call
/******/ 	__webpack_require__.o = function(object, property) { return Object.prototype.hasOwnProperty.call(object, property); };
/******/
/******/ 	// __webpack_public_path__
/******/ 	__webpack_require__.p = "";
/******/
/******/
/******/ 	// Load entry module and return exports
/******/ 	return __webpack_require__(__webpack_require__.s = "./src/main/js/index.ts");
/******/ })
/************************************************************************/
/******/ ({

/***/ "./node_modules/css-loader/lib/css-base.js":
/*!*************************************************!*\
  !*** ./node_modules/css-loader/lib/css-base.js ***!
  \*************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

/*
	MIT License http://www.opensource.org/licenses/mit-license.php
	Author Tobias Koppers @sokra
*/
// css base code, injected by the css-loader
module.exports = function(useSourceMap) {
	var list = [];

	// return the list of modules as css string
	list.toString = function toString() {
		return this.map(function (item) {
			var content = cssWithMappingToString(item, useSourceMap);
			if(item[2]) {
				return "@media " + item[2] + "{" + content + "}";
			} else {
				return content;
			}
		}).join("");
	};

	// import a list of modules into the list
	list.i = function(modules, mediaQuery) {
		if(typeof modules === "string")
			modules = [[null, modules, ""]];
		var alreadyImportedModules = {};
		for(var i = 0; i < this.length; i++) {
			var id = this[i][0];
			if(typeof id === "number")
				alreadyImportedModules[id] = true;
		}
		for(i = 0; i < modules.length; i++) {
			var item = modules[i];
			// skip already imported module
			// this implementation is not 100% perfect for weird media query combinations
			//  when a module is imported multiple times with different media queries.
			//  I hope this will never occur (Hey this way we have smaller bundles)
			if(typeof item[0] !== "number" || !alreadyImportedModules[item[0]]) {
				if(mediaQuery && !item[2]) {
					item[2] = mediaQuery;
				} else if(mediaQuery) {
					item[2] = "(" + item[2] + ") and (" + mediaQuery + ")";
				}
				list.push(item);
			}
		}
	};
	return list;
};

function cssWithMappingToString(item, useSourceMap) {
	var content = item[1] || '';
	var cssMapping = item[3];
	if (!cssMapping) {
		return content;
	}

	if (useSourceMap && typeof btoa === 'function') {
		var sourceMapping = toComment(cssMapping);
		var sourceURLs = cssMapping.sources.map(function (source) {
			return '/*# sourceURL=' + cssMapping.sourceRoot + source + ' */'
		});

		return [content].concat(sourceURLs).concat([sourceMapping]).join('\n');
	}

	return [content].join('\n');
}

// Adapted from convert-source-map (MIT)
function toComment(sourceMap) {
	// eslint-disable-next-line no-undef
	var base64 = btoa(unescape(encodeURIComponent(JSON.stringify(sourceMap))));
	var data = 'sourceMappingURL=data:application/json;charset=utf-8;base64,' + base64;

	return '/*# ' + data + ' */';
}


/***/ }),

/***/ "./node_modules/process/browser.js":
/*!*****************************************!*\
  !*** ./node_modules/process/browser.js ***!
  \*****************************************/
/*! no static exports found */
/***/ (function(module, exports) {

// shim for using process in browser
var process = module.exports = {};

// cached from whatever global is present so that test runners that stub it
// don't break things.  But we need to wrap it in a try catch in case it is
// wrapped in strict mode code which doesn't define any globals.  It's inside a
// function because try/catches deoptimize in certain engines.

var cachedSetTimeout;
var cachedClearTimeout;

function defaultSetTimout() {
    throw new Error('setTimeout has not been defined');
}
function defaultClearTimeout () {
    throw new Error('clearTimeout has not been defined');
}
(function () {
    try {
        if (typeof setTimeout === 'function') {
            cachedSetTimeout = setTimeout;
        } else {
            cachedSetTimeout = defaultSetTimout;
        }
    } catch (e) {
        cachedSetTimeout = defaultSetTimout;
    }
    try {
        if (typeof clearTimeout === 'function') {
            cachedClearTimeout = clearTimeout;
        } else {
            cachedClearTimeout = defaultClearTimeout;
        }
    } catch (e) {
        cachedClearTimeout = defaultClearTimeout;
    }
} ())
function runTimeout(fun) {
    if (cachedSetTimeout === setTimeout) {
        //normal enviroments in sane situations
        return setTimeout(fun, 0);
    }
    // if setTimeout wasn't available but was latter defined
    if ((cachedSetTimeout === defaultSetTimout || !cachedSetTimeout) && setTimeout) {
        cachedSetTimeout = setTimeout;
        return setTimeout(fun, 0);
    }
    try {
        // when when somebody has screwed with setTimeout but no I.E. maddness
        return cachedSetTimeout(fun, 0);
    } catch(e){
        try {
            // When we are in I.E. but the script has been evaled so I.E. doesn't trust the global object when called normally
            return cachedSetTimeout.call(null, fun, 0);
        } catch(e){
            // same as above but when it's a version of I.E. that must have the global object for 'this', hopfully our context correct otherwise it will throw a global error
            return cachedSetTimeout.call(this, fun, 0);
        }
    }


}
function runClearTimeout(marker) {
    if (cachedClearTimeout === clearTimeout) {
        //normal enviroments in sane situations
        return clearTimeout(marker);
    }
    // if clearTimeout wasn't available but was latter defined
    if ((cachedClearTimeout === defaultClearTimeout || !cachedClearTimeout) && clearTimeout) {
        cachedClearTimeout = clearTimeout;
        return clearTimeout(marker);
    }
    try {
        // when when somebody has screwed with setTimeout but no I.E. maddness
        return cachedClearTimeout(marker);
    } catch (e){
        try {
            // When we are in I.E. but the script has been evaled so I.E. doesn't  trust the global object when called normally
            return cachedClearTimeout.call(null, marker);
        } catch (e){
            // same as above but when it's a version of I.E. that must have the global object for 'this', hopfully our context correct otherwise it will throw a global error.
            // Some versions of I.E. have different rules for clearTimeout vs setTimeout
            return cachedClearTimeout.call(this, marker);
        }
    }



}
var queue = [];
var draining = false;
var currentQueue;
var queueIndex = -1;

function cleanUpNextTick() {
    if (!draining || !currentQueue) {
        return;
    }
    draining = false;
    if (currentQueue.length) {
        queue = currentQueue.concat(queue);
    } else {
        queueIndex = -1;
    }
    if (queue.length) {
        drainQueue();
    }
}

function drainQueue() {
    if (draining) {
        return;
    }
    var timeout = runTimeout(cleanUpNextTick);
    draining = true;

    var len = queue.length;
    while(len) {
        currentQueue = queue;
        queue = [];
        while (++queueIndex < len) {
            if (currentQueue) {
                currentQueue[queueIndex].run();
            }
        }
        queueIndex = -1;
        len = queue.length;
    }
    currentQueue = null;
    draining = false;
    runClearTimeout(timeout);
}

process.nextTick = function (fun) {
    var args = new Array(arguments.length - 1);
    if (arguments.length > 1) {
        for (var i = 1; i < arguments.length; i++) {
            args[i - 1] = arguments[i];
        }
    }
    queue.push(new Item(fun, args));
    if (queue.length === 1 && !draining) {
        runTimeout(drainQueue);
    }
};

// v8 likes predictible objects
function Item(fun, array) {
    this.fun = fun;
    this.array = array;
}
Item.prototype.run = function () {
    this.fun.apply(null, this.array);
};
process.title = 'browser';
process.browser = true;
process.env = {};
process.argv = [];
process.version = ''; // empty string to avoid regexp issues
process.versions = {};

function noop() {}

process.on = noop;
process.addListener = noop;
process.once = noop;
process.off = noop;
process.removeListener = noop;
process.removeAllListeners = noop;
process.emit = noop;
process.prependListener = noop;
process.prependOnceListener = noop;

process.listeners = function (name) { return [] }

process.binding = function (name) {
    throw new Error('process.binding is not supported');
};

process.cwd = function () { return '/' };
process.chdir = function (dir) {
    throw new Error('process.chdir is not supported');
};
process.umask = function() { return 0; };


/***/ }),

/***/ "./src/main/js/api/button.ts":
/*!***********************************!*\
  !*** ./src/main/js/api/button.ts ***!
  \***********************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var ButtonApi = /** @class */ (function () {
    /**
     * @hidden
     */
    function ButtonApi(buttonController) {
        this.controller = buttonController;
    }
    ButtonApi.prototype.dispose = function () {
        this.controller.dispose();
    };
    ButtonApi.prototype.on = function (eventName, handler) {
        var emitter = this.controller.button.emitter;
        emitter.on(eventName, handler);
    };
    return ButtonApi;
}());
exports.default = ButtonApi;


/***/ }),

/***/ "./src/main/js/api/folder.ts":
/*!***********************************!*\
  !*** ./src/main/js/api/folder.ts ***!
  \***********************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var InputBindingControllerCreators = __webpack_require__(/*! ../controller/binding-creators/input */ "./src/main/js/controller/binding-creators/input.ts");
var MonitorBindingControllerCreators = __webpack_require__(/*! ../controller/binding-creators/monitor */ "./src/main/js/controller/binding-creators/monitor.ts");
var button_1 = __webpack_require__(/*! ../controller/button */ "./src/main/js/controller/button.ts");
var separator_1 = __webpack_require__(/*! ../controller/separator */ "./src/main/js/controller/separator.ts");
var target_1 = __webpack_require__(/*! ../model/target */ "./src/main/js/model/target.ts");
var button_2 = __webpack_require__(/*! ./button */ "./src/main/js/api/button.ts");
var input_binding_1 = __webpack_require__(/*! ./input-binding */ "./src/main/js/api/input-binding.ts");
var monitor_binding_1 = __webpack_require__(/*! ./monitor-binding */ "./src/main/js/api/monitor-binding.ts");
var TO_INTERNAL_EVENT_NAME_MAP = {
    change: 'inputchange',
    fold: 'fold',
    update: 'monitorupdate',
};
var FolderApi = /** @class */ (function () {
    /**
     * @hidden
     */
    function FolderApi(folderController) {
        this.controller = folderController;
    }
    Object.defineProperty(FolderApi.prototype, "expanded", {
        get: function () {
            return this.controller.folder.expanded;
        },
        set: function (expanded) {
            this.controller.folder.expanded = expanded;
        },
        enumerable: true,
        configurable: true
    });
    FolderApi.prototype.dispose = function () {
        this.controller.dispose();
    };
    FolderApi.prototype.addInput = function (object, key, opt_params) {
        var params = opt_params || {};
        var uc = InputBindingControllerCreators.create(this.controller.document, new target_1.default(object, key, params.presetKey), params);
        this.controller.uiControllerList.append(uc);
        return new input_binding_1.default(uc);
    };
    FolderApi.prototype.addMonitor = function (object, key, opt_params) {
        var params = opt_params || {};
        var uc = MonitorBindingControllerCreators.create(this.controller.document, new target_1.default(object, key), params);
        this.controller.uiControllerList.append(uc);
        return new monitor_binding_1.default(uc);
    };
    FolderApi.prototype.addButton = function (params) {
        var uc = new button_1.default(this.controller.document, params);
        this.controller.uiControllerList.append(uc);
        return new button_2.default(uc);
    };
    FolderApi.prototype.addSeparator = function () {
        var uc = new separator_1.default(this.controller.document);
        this.controller.uiControllerList.append(uc);
    };
    FolderApi.prototype.on = function (eventName, handler) {
        var internalEventName = TO_INTERNAL_EVENT_NAME_MAP[eventName];
        if (internalEventName) {
            var emitter = this.controller.emitter;
            emitter.on(internalEventName, handler);
        }
        return this;
    };
    return FolderApi;
}());
exports.default = FolderApi;


/***/ }),

/***/ "./src/main/js/api/input-binding.ts":
/*!******************************************!*\
  !*** ./src/main/js/api/input-binding.ts ***!
  \******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * The API for the input binding between the parameter and the pane.
 * @param In The type inner Tweakpane.
 * @param Out The type outer Tweakpane (= parameter object).
 */
var InputBindingApi = /** @class */ (function () {
    /**
     * @hidden
     */
    function InputBindingApi(bindingController) {
        this.controller = bindingController;
    }
    InputBindingApi.prototype.dispose = function () {
        this.controller.dispose();
    };
    InputBindingApi.prototype.on = function (eventName, handler) {
        var emitter = this.controller.binding.value.emitter;
        emitter.on(eventName, handler);
        return this;
    };
    InputBindingApi.prototype.refresh = function () {
        this.controller.binding.read();
    };
    return InputBindingApi;
}());
exports.default = InputBindingApi;


/***/ }),

/***/ "./src/main/js/api/monitor-binding.ts":
/*!********************************************!*\
  !*** ./src/main/js/api/monitor-binding.ts ***!
  \********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * The API for the monitor binding between the parameter and the pane.
 */
var MonitorBindingApi = /** @class */ (function () {
    /**
     * @hidden
     */
    function MonitorBindingApi(bindingController) {
        this.controller = bindingController;
    }
    MonitorBindingApi.prototype.dispose = function () {
        this.controller.dispose();
    };
    MonitorBindingApi.prototype.on = function (eventName, handler) {
        var emitter = this.controller.binding.value.emitter;
        emitter.on(eventName, handler);
        return this;
    };
    MonitorBindingApi.prototype.refresh = function () {
        this.controller.binding.read();
    };
    return MonitorBindingApi;
}());
exports.default = MonitorBindingApi;


/***/ }),

/***/ "./src/main/js/api/preset.ts":
/*!***********************************!*\
  !*** ./src/main/js/api/preset.ts ***!
  \***********************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
function exportJson(targets) {
    return targets.reduce(function (result, target) {
        var _a;
        return Object.assign(result, (_a = {},
            _a[target.presetKey] = target.read(),
            _a));
    }, {});
}
exports.exportJson = exportJson;
/**
 * @hidden
 */
function importJson(targets, preset) {
    targets.forEach(function (target) {
        var value = preset[target.presetKey];
        if (value !== undefined) {
            target.write(value);
        }
    });
}
exports.importJson = importJson;


/***/ }),

/***/ "./src/main/js/api/root.ts":
/*!*********************************!*\
  !*** ./src/main/js/api/root.ts ***!
  \*********************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var InputBindingControllerCreators = __webpack_require__(/*! ../controller/binding-creators/input */ "./src/main/js/controller/binding-creators/input.ts");
var MonitorBindingControllerCreators = __webpack_require__(/*! ../controller/binding-creators/monitor */ "./src/main/js/controller/binding-creators/monitor.ts");
var button_1 = __webpack_require__(/*! ../controller/button */ "./src/main/js/controller/button.ts");
var folder_1 = __webpack_require__(/*! ../controller/folder */ "./src/main/js/controller/folder.ts");
var input_binding_1 = __webpack_require__(/*! ../controller/input-binding */ "./src/main/js/controller/input-binding.ts");
var monitor_binding_1 = __webpack_require__(/*! ../controller/monitor-binding */ "./src/main/js/controller/monitor-binding.ts");
var separator_1 = __webpack_require__(/*! ../controller/separator */ "./src/main/js/controller/separator.ts");
var UiUtil = __webpack_require__(/*! ../controller/ui-util */ "./src/main/js/controller/ui-util.ts");
var target_1 = __webpack_require__(/*! ../model/target */ "./src/main/js/model/target.ts");
var button_2 = __webpack_require__(/*! ./button */ "./src/main/js/api/button.ts");
var folder_2 = __webpack_require__(/*! ./folder */ "./src/main/js/api/folder.ts");
var input_binding_2 = __webpack_require__(/*! ./input-binding */ "./src/main/js/api/input-binding.ts");
var monitor_binding_2 = __webpack_require__(/*! ./monitor-binding */ "./src/main/js/api/monitor-binding.ts");
var Preset = __webpack_require__(/*! ./preset */ "./src/main/js/api/preset.ts");
var TO_INTERNAL_EVENT_NAME_MAP = {
    change: 'inputchange',
    fold: 'fold',
    update: 'monitorupdate',
};
/**
 * The Tweakpane interface.
 *
 * ```
 * new Tweakpane(options: TweakpaneConfig): RootApi
 * ```
 *
 * See [[TweakpaneConfig]] interface for available options.
 */
var RootApi = /** @class */ (function () {
    /**
     * @hidden
     */
    function RootApi(rootController) {
        this.controller = rootController;
    }
    Object.defineProperty(RootApi.prototype, "element", {
        get: function () {
            return this.controller.view.element;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(RootApi.prototype, "expanded", {
        get: function () {
            var folder = this.controller.folder;
            return folder ? folder.expanded : true;
        },
        set: function (expanded) {
            var folder = this.controller.folder;
            if (folder) {
                folder.expanded = expanded;
            }
        },
        enumerable: true,
        configurable: true
    });
    RootApi.prototype.dispose = function () {
        this.controller.dispose();
    };
    RootApi.prototype.addInput = function (object, key, opt_params) {
        var params = opt_params || {};
        var uc = InputBindingControllerCreators.create(this.controller.document, new target_1.default(object, key, params.presetKey), params);
        this.controller.uiControllerList.append(uc);
        return new input_binding_2.default(uc);
    };
    RootApi.prototype.addMonitor = function (object, key, opt_params) {
        var params = opt_params || {};
        var uc = MonitorBindingControllerCreators.create(this.controller.document, new target_1.default(object, key), params);
        this.controller.uiControllerList.append(uc);
        return new monitor_binding_2.default(uc);
    };
    RootApi.prototype.addButton = function (params) {
        var uc = new button_1.default(this.controller.document, params);
        this.controller.uiControllerList.append(uc);
        return new button_2.default(uc);
    };
    RootApi.prototype.addFolder = function (params) {
        var uc = new folder_1.default(this.controller.document, params);
        this.controller.uiControllerList.append(uc);
        return new folder_2.default(uc);
    };
    RootApi.prototype.addSeparator = function () {
        var uc = new separator_1.default(this.controller.document);
        this.controller.uiControllerList.append(uc);
    };
    /**
     * Import a preset of all inputs.
     * @param preset The preset object to import.
     */
    RootApi.prototype.importPreset = function (preset) {
        var targets = UiUtil.findControllers(this.controller.uiControllerList.items, input_binding_1.default).map(function (ibc) {
            return ibc.binding.target;
        });
        Preset.importJson(targets, preset);
        this.refresh();
    };
    /**
     * Export a preset of all inputs.
     * @return The exported preset object.
     */
    RootApi.prototype.exportPreset = function () {
        var targets = UiUtil.findControllers(this.controller.uiControllerList.items, input_binding_1.default).map(function (ibc) {
            return ibc.binding.target;
        });
        return Preset.exportJson(targets);
    };
    /**
     * Adds a global event listener. It handles all events of child inputs/monitors.
     * @param eventName The event name to listen.
     * @return The API object itself.
     */
    RootApi.prototype.on = function (eventName, handler) {
        var internalEventName = TO_INTERNAL_EVENT_NAME_MAP[eventName];
        if (internalEventName) {
            var emitter = this.controller.emitter;
            emitter.on(internalEventName, handler);
        }
        return this;
    };
    /**
     * Refreshes all bindings of the pane.
     */
    RootApi.prototype.refresh = function () {
        // Force-read all input bindings
        UiUtil.findControllers(this.controller.uiControllerList.items, input_binding_1.default).forEach(function (ibc) {
            ibc.binding.read();
        });
        // Force-read all monitor bindings
        UiUtil.findControllers(this.controller.uiControllerList.items, monitor_binding_1.default).forEach(function (mbc) {
            mbc.binding.read();
        });
    };
    return RootApi;
}());
exports.default = RootApi;


/***/ }),

/***/ "./src/main/js/binding/input.ts":
/*!**************************************!*\
  !*** ./src/main/js/binding/input.ts ***!
  \**************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var InputBinding = /** @class */ (function () {
    function InputBinding(config) {
        this.onValueChange_ = this.onValueChange_.bind(this);
        this.reader_ = config.reader;
        this.writer_ = config.writer;
        this.value = config.value;
        this.value.emitter.on('change', this.onValueChange_);
        this.target = config.target;
        this.read();
    }
    InputBinding.prototype.read = function () {
        var targetValue = this.target.read();
        if (targetValue !== undefined) {
            this.value.rawValue = this.reader_(targetValue);
        }
    };
    InputBinding.prototype.write_ = function (rawValue) {
        var value = this.writer_(rawValue);
        this.target.write(value);
    };
    InputBinding.prototype.onValueChange_ = function (rawValue) {
        this.write_(rawValue);
    };
    return InputBinding;
}());
exports.default = InputBinding;


/***/ }),

/***/ "./src/main/js/binding/monitor.ts":
/*!****************************************!*\
  !*** ./src/main/js/binding/monitor.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var MonitorBinding = /** @class */ (function () {
    function MonitorBinding(config) {
        this.onTick_ = this.onTick_.bind(this);
        this.reader_ = config.reader;
        this.target = config.target;
        this.value = config.value;
        this.ticker = config.ticker;
        this.ticker.emitter.on('tick', this.onTick_);
        this.read();
    }
    MonitorBinding.prototype.dispose = function () {
        this.ticker.dispose();
    };
    MonitorBinding.prototype.read = function () {
        var targetValue = this.target.read();
        if (targetValue !== undefined) {
            this.value.append(this.reader_(targetValue));
        }
    };
    MonitorBinding.prototype.onTick_ = function () {
        this.read();
    };
    return MonitorBinding;
}());
exports.default = MonitorBinding;


/***/ }),

/***/ "./src/main/js/constraint/composite.ts":
/*!*********************************************!*\
  !*** ./src/main/js/constraint/composite.ts ***!
  \*********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var CompositeConstraint = /** @class */ (function () {
    function CompositeConstraint(config) {
        this.constraints_ = config.constraints;
    }
    Object.defineProperty(CompositeConstraint.prototype, "constraints", {
        get: function () {
            return this.constraints_;
        },
        enumerable: true,
        configurable: true
    });
    CompositeConstraint.prototype.constrain = function (value) {
        return this.constraints_.reduce(function (result, c) {
            return c.constrain(result);
        }, value);
    };
    return CompositeConstraint;
}());
exports.default = CompositeConstraint;


/***/ }),

/***/ "./src/main/js/constraint/list.ts":
/*!****************************************!*\
  !*** ./src/main/js/constraint/list.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var ListConstraint = /** @class */ (function () {
    function ListConstraint(config) {
        this.opts_ = config.options;
    }
    Object.defineProperty(ListConstraint.prototype, "options", {
        get: function () {
            return this.opts_;
        },
        enumerable: true,
        configurable: true
    });
    ListConstraint.prototype.constrain = function (value) {
        var opts = this.opts_;
        if (opts.length === 0) {
            return value;
        }
        var matched = opts.filter(function (item) {
            return item.value === value;
        }).length > 0;
        return matched ? value : opts[0].value;
    };
    return ListConstraint;
}());
exports.default = ListConstraint;


/***/ }),

/***/ "./src/main/js/constraint/range.ts":
/*!*****************************************!*\
  !*** ./src/main/js/constraint/range.ts ***!
  \*****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var RangeConstraint = /** @class */ (function () {
    function RangeConstraint(config) {
        this.max_ = config.max;
        this.min_ = config.min;
    }
    Object.defineProperty(RangeConstraint.prototype, "minValue", {
        get: function () {
            return this.min_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(RangeConstraint.prototype, "maxValue", {
        get: function () {
            return this.max_;
        },
        enumerable: true,
        configurable: true
    });
    RangeConstraint.prototype.constrain = function (value) {
        var result = value;
        if (this.min_ !== null && this.min_ !== undefined) {
            result = Math.max(result, this.min_);
        }
        if (this.max_ !== null && this.max_ !== undefined) {
            result = Math.min(result, this.max_);
        }
        return result;
    };
    return RangeConstraint;
}());
exports.default = RangeConstraint;


/***/ }),

/***/ "./src/main/js/constraint/step.ts":
/*!****************************************!*\
  !*** ./src/main/js/constraint/step.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var StepConstraint = /** @class */ (function () {
    function StepConstraint(config) {
        this.step = config.step;
    }
    StepConstraint.prototype.constrain = function (value) {
        var r = value < 0
            ? -Math.round(-value / this.step)
            : Math.round(value / this.step);
        return r * this.step;
    };
    return StepConstraint;
}());
exports.default = StepConstraint;


/***/ }),

/***/ "./src/main/js/constraint/util.ts":
/*!****************************************!*\
  !*** ./src/main/js/constraint/util.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var composite_1 = __webpack_require__(/*! ./composite */ "./src/main/js/constraint/composite.ts");
/**
 * @hidden
 */
var ConstraintUtil = {
    findConstraint: function (c, constraintClass) {
        if (c instanceof constraintClass) {
            return c;
        }
        if (c instanceof composite_1.default) {
            var result = c.constraints.reduce(function (tmpResult, sc) {
                if (tmpResult) {
                    return tmpResult;
                }
                return sc instanceof constraintClass ? sc : null;
            }, null);
            if (result) {
                return result;
            }
        }
        return null;
    },
};
exports.default = ConstraintUtil;


/***/ }),

/***/ "./src/main/js/controller/binding-creators/boolean-input.ts":
/*!******************************************************************!*\
  !*** ./src/main/js/controller/binding-creators/boolean-input.ts ***!
  \******************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var input_1 = __webpack_require__(/*! ../../binding/input */ "./src/main/js/binding/input.ts");
var composite_1 = __webpack_require__(/*! ../../constraint/composite */ "./src/main/js/constraint/composite.ts");
var list_1 = __webpack_require__(/*! ../../constraint/list */ "./src/main/js/constraint/list.ts");
var util_1 = __webpack_require__(/*! ../../constraint/util */ "./src/main/js/constraint/util.ts");
var BooleanConverter = __webpack_require__(/*! ../../converter/boolean */ "./src/main/js/converter/boolean.ts");
var input_value_1 = __webpack_require__(/*! ../../model/input-value */ "./src/main/js/model/input-value.ts");
var input_binding_1 = __webpack_require__(/*! ../input-binding */ "./src/main/js/controller/input-binding.ts");
var checkbox_1 = __webpack_require__(/*! ../input/checkbox */ "./src/main/js/controller/input/checkbox.ts");
var list_2 = __webpack_require__(/*! ../input/list */ "./src/main/js/controller/input/list.ts");
function createConstraint(params) {
    var constraints = [];
    if (params.options) {
        constraints.push(new list_1.default({
            options: params.options,
        }));
    }
    return new composite_1.default({
        constraints: constraints,
    });
}
function createController(document, value) {
    var c = value.constraint;
    if (c && util_1.default.findConstraint(c, list_1.default)) {
        return new list_2.default(document, {
            stringifyValue: BooleanConverter.toString,
            value: value,
        });
    }
    return new checkbox_1.default(document, {
        value: value,
    });
}
/**
 * @hidden
 */
function create(document, target, params) {
    var value = new input_value_1.default(false, createConstraint(params));
    var binding = new input_1.default({
        reader: BooleanConverter.fromMixed,
        target: target,
        value: value,
        writer: function (v) { return v; },
    });
    return new input_binding_1.default(document, {
        binding: binding,
        controller: createController(document, value),
        label: params.label || target.key,
    });
}
exports.create = create;


/***/ }),

/***/ "./src/main/js/controller/binding-creators/boolean-monitor.ts":
/*!********************************************************************!*\
  !*** ./src/main/js/controller/binding-creators/boolean-monitor.ts ***!
  \********************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var monitor_1 = __webpack_require__(/*! ../../binding/monitor */ "./src/main/js/binding/monitor.ts");
var BooleanConverter = __webpack_require__(/*! ../../converter/boolean */ "./src/main/js/converter/boolean.ts");
var boolean_1 = __webpack_require__(/*! ../../formatter/boolean */ "./src/main/js/formatter/boolean.ts");
var interval_1 = __webpack_require__(/*! ../../misc/ticker/interval */ "./src/main/js/misc/ticker/interval.ts");
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var monitor_value_1 = __webpack_require__(/*! ../../model/monitor-value */ "./src/main/js/model/monitor-value.ts");
var monitor_binding_1 = __webpack_require__(/*! ../monitor-binding */ "./src/main/js/controller/monitor-binding.ts");
var multi_log_1 = __webpack_require__(/*! ../monitor/multi-log */ "./src/main/js/controller/monitor/multi-log.ts");
var single_log_1 = __webpack_require__(/*! ../monitor/single-log */ "./src/main/js/controller/monitor/single-log.ts");
/**
 * @hidden
 */
function createTextMonitor(document, target, params) {
    var value = new monitor_value_1.default(type_util_1.default.getOrDefault(params.count, 1));
    var controller = value.totalCount === 1
        ? new single_log_1.default(document, {
            formatter: new boolean_1.default(),
            value: value,
        })
        : new multi_log_1.default(document, {
            formatter: new boolean_1.default(),
            value: value,
        });
    var ticker = new interval_1.default(document, type_util_1.default.getOrDefault(params.interval, 200));
    return new monitor_binding_1.default(document, {
        binding: new monitor_1.default({
            reader: BooleanConverter.fromMixed,
            target: target,
            ticker: ticker,
            value: value,
        }),
        controller: controller,
        label: params.label || target.key,
    });
}
exports.createTextMonitor = createTextMonitor;


/***/ }),

/***/ "./src/main/js/controller/binding-creators/color-input.ts":
/*!****************************************************************!*\
  !*** ./src/main/js/controller/binding-creators/color-input.ts ***!
  \****************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var input_1 = __webpack_require__(/*! ../../binding/input */ "./src/main/js/binding/input.ts");
var ColorConverter = __webpack_require__(/*! ../../converter/color */ "./src/main/js/converter/color.ts");
var color_1 = __webpack_require__(/*! ../../formatter/color */ "./src/main/js/formatter/color.ts");
var input_value_1 = __webpack_require__(/*! ../../model/input-value */ "./src/main/js/model/input-value.ts");
var color_2 = __webpack_require__(/*! ../../parser/color */ "./src/main/js/parser/color.ts");
var input_binding_1 = __webpack_require__(/*! ../input-binding */ "./src/main/js/controller/input-binding.ts");
var color_swatch_text_1 = __webpack_require__(/*! ../input/color-swatch-text */ "./src/main/js/controller/input/color-swatch-text.ts");
/**
 * @hidden
 */
function create(document, target, initialValue, params) {
    var value = new input_value_1.default(initialValue);
    return new input_binding_1.default(document, {
        binding: new input_1.default({
            reader: ColorConverter.fromMixed,
            target: target,
            value: value,
            writer: ColorConverter.toString,
        }),
        controller: new color_swatch_text_1.default(document, {
            formatter: new color_1.default(),
            parser: color_2.default,
            value: value,
        }),
        label: params.label || target.key,
    });
}
exports.create = create;


/***/ }),

/***/ "./src/main/js/controller/binding-creators/input.ts":
/*!**********************************************************!*\
  !*** ./src/main/js/controller/binding-creators/input.ts ***!
  \**********************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var BooleanConverter = __webpack_require__(/*! ../../converter/boolean */ "./src/main/js/converter/boolean.ts");
var NumberConverter = __webpack_require__(/*! ../../converter/number */ "./src/main/js/converter/number.ts");
var StringConverter = __webpack_require__(/*! ../../converter/string */ "./src/main/js/converter/string.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var color_1 = __webpack_require__(/*! ../../parser/color */ "./src/main/js/parser/color.ts");
var BooleanInputBindingControllerCreators = __webpack_require__(/*! ./boolean-input */ "./src/main/js/controller/binding-creators/boolean-input.ts");
var ColorInputBindingControllerCreators = __webpack_require__(/*! ./color-input */ "./src/main/js/controller/binding-creators/color-input.ts");
var NumberInputBindingControllerCreators = __webpack_require__(/*! ./number-input */ "./src/main/js/controller/binding-creators/number-input.ts");
var StringInputBindingControllerCreators = __webpack_require__(/*! ./string-input */ "./src/main/js/controller/binding-creators/string-input.ts");
function normalizeParams(p1, convert) {
    var p2 = {
        label: p1.label,
        max: p1.max,
        min: p1.min,
        step: p1.step,
    };
    if (p1.options) {
        if (Array.isArray(p1.options)) {
            p2.options = p1.options.map(function (item) {
                return {
                    text: item.text,
                    value: convert(item.value),
                };
            });
        }
        else {
            var textToValueMap_1 = p1.options;
            var texts = Object.keys(textToValueMap_1);
            p2.options = texts.reduce(function (options, text) {
                return options.concat({
                    text: text,
                    value: convert(textToValueMap_1[text]),
                });
            }, []);
        }
    }
    return p2;
}
/**
 * @hidden
 */
function create(document, target, params) {
    var initialValue = target.read();
    if (initialValue === null || initialValue === undefined) {
        throw new pane_error_1.default({
            context: {
                key: target.key,
            },
            type: 'emptyvalue',
        });
    }
    if (typeof initialValue === 'boolean') {
        return BooleanInputBindingControllerCreators.create(document, target, normalizeParams(params, BooleanConverter.fromMixed));
    }
    if (typeof initialValue === 'number') {
        return NumberInputBindingControllerCreators.create(document, target, normalizeParams(params, NumberConverter.fromMixed));
    }
    if (typeof initialValue === 'string') {
        if(params.input === 'color'){
            var color = color_1.default(initialValue);
            if (color) {
                return ColorInputBindingControllerCreators.create(document, target, color, params);
            }
        }
        return StringInputBindingControllerCreators.create(document, target, normalizeParams(params, StringConverter.fromMixed));
    }
    throw new pane_error_1.default({
        context: {
            key: target.key,
        },
        type: 'nomatchingcontroller',
    });
}
exports.create = create;


/***/ }),

/***/ "./src/main/js/controller/binding-creators/monitor.ts":
/*!************************************************************!*\
  !*** ./src/main/js/controller/binding-creators/monitor.ts ***!
  \************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var BooleanMonitorBindingControllerCreators = __webpack_require__(/*! ./boolean-monitor */ "./src/main/js/controller/binding-creators/boolean-monitor.ts");
var NumberMonitorBindingControllerCreators = __webpack_require__(/*! ./number-monitor */ "./src/main/js/controller/binding-creators/number-monitor.ts");
var StringMonitorBindingControllerCreators = __webpack_require__(/*! ./string-monitor */ "./src/main/js/controller/binding-creators/string-monitor.ts");
/**
 * @hidden
 */
function create(document, target, params) {
    var initialValue = target.read();
    if (initialValue === null || initialValue === undefined) {
        throw new pane_error_1.default({
            context: {
                key: target.key,
            },
            type: 'emptyvalue',
        });
    }
    if (typeof initialValue === 'number') {
        if (params.type === 'graph') {
            return NumberMonitorBindingControllerCreators.createGraphMonitor(document, target, params);
        }
        return NumberMonitorBindingControllerCreators.createTextMonitor(document, target, params);
    }
    if (typeof initialValue === 'string') {
        return StringMonitorBindingControllerCreators.createTextMonitor(document, target, params);
    }
    if (typeof initialValue === 'boolean') {
        return BooleanMonitorBindingControllerCreators.createTextMonitor(document, target, params);
    }
    throw new pane_error_1.default({
        context: {
            key: target.key,
        },
        type: 'nomatchingcontroller',
    });
}
exports.create = create;


/***/ }),

/***/ "./src/main/js/controller/binding-creators/number-input.ts":
/*!*****************************************************************!*\
  !*** ./src/main/js/controller/binding-creators/number-input.ts ***!
  \*****************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var input_1 = __webpack_require__(/*! ../../binding/input */ "./src/main/js/binding/input.ts");
var composite_1 = __webpack_require__(/*! ../../constraint/composite */ "./src/main/js/constraint/composite.ts");
var list_1 = __webpack_require__(/*! ../../constraint/list */ "./src/main/js/constraint/list.ts");
var range_1 = __webpack_require__(/*! ../../constraint/range */ "./src/main/js/constraint/range.ts");
var step_1 = __webpack_require__(/*! ../../constraint/step */ "./src/main/js/constraint/step.ts");
var util_1 = __webpack_require__(/*! ../../constraint/util */ "./src/main/js/constraint/util.ts");
var NumberConverter = __webpack_require__(/*! ../../converter/number */ "./src/main/js/converter/number.ts");
var number_1 = __webpack_require__(/*! ../../formatter/number */ "./src/main/js/formatter/number.ts");
var number_util_1 = __webpack_require__(/*! ../../misc/number-util */ "./src/main/js/misc/number-util.ts");
var input_value_1 = __webpack_require__(/*! ../../model/input-value */ "./src/main/js/model/input-value.ts");
var number_2 = __webpack_require__(/*! ../../parser/number */ "./src/main/js/parser/number.ts");
var input_binding_1 = __webpack_require__(/*! ../input-binding */ "./src/main/js/controller/input-binding.ts");
var list_2 = __webpack_require__(/*! ../input/list */ "./src/main/js/controller/input/list.ts");
var number_text_1 = __webpack_require__(/*! ../input/number-text */ "./src/main/js/controller/input/number-text.ts");
var slider_text_1 = __webpack_require__(/*! ../input/slider-text */ "./src/main/js/controller/input/slider-text.ts");
function createConstraint(params) {
    var constraints = [];
    if (params.step !== null && params.step !== undefined) {
        constraints.push(new step_1.default({
            step: params.step,
        }));
    }
    if ((params.max !== null && params.max !== undefined) ||
        (params.min !== null && params.min !== undefined)) {
        constraints.push(new range_1.default({
            max: params.max,
            min: params.min,
        }));
    }
    if (params.options) {
        constraints.push(new list_1.default({
            options: params.options,
        }));
    }
    return new composite_1.default({
        constraints: constraints,
    });
}
function getSuitableDecimalDigits(value) {
    var c = value.constraint;
    var sc = c && util_1.default.findConstraint(c, step_1.default);
    if (sc) {
        return number_util_1.default.getDecimalDigits(sc.step);
    }
    return Math.max(number_util_1.default.getDecimalDigits(value.rawValue), 2);
}
function createController(document, value) {
    var c = value.constraint;
    if (c && util_1.default.findConstraint(c, list_1.default)) {
        return new list_2.default(document, {
            stringifyValue: NumberConverter.toString,
            value: value,
        });
    }
    if (c && util_1.default.findConstraint(c, range_1.default)) {
        return new slider_text_1.default(document, {
            formatter: new number_1.default(getSuitableDecimalDigits(value)),
            parser: number_2.default,
            value: value,
        });
    }
    return new number_text_1.default(document, {
        formatter: new number_1.default(getSuitableDecimalDigits(value)),
        parser: number_2.default,
        value: value,
    });
}
function create(document, target, params) {
    var value = new input_value_1.default(0, createConstraint(params));
    var binding = new input_1.default({
        reader: NumberConverter.fromMixed,
        target: target,
        value: value,
        writer: function (v) { return v; },
    });
    return new input_binding_1.default(document, {
        binding: binding,
        controller: createController(document, value),
        label: params.label || target.key,
    });
}
exports.create = create;


/***/ }),

/***/ "./src/main/js/controller/binding-creators/number-monitor.ts":
/*!*******************************************************************!*\
  !*** ./src/main/js/controller/binding-creators/number-monitor.ts ***!
  \*******************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var monitor_1 = __webpack_require__(/*! ../../binding/monitor */ "./src/main/js/binding/monitor.ts");
var NumberConverter = __webpack_require__(/*! ../../converter/number */ "./src/main/js/converter/number.ts");
var number_1 = __webpack_require__(/*! ../../formatter/number */ "./src/main/js/formatter/number.ts");
var interval_1 = __webpack_require__(/*! ../../misc/ticker/interval */ "./src/main/js/misc/ticker/interval.ts");
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var monitor_value_1 = __webpack_require__(/*! ../../model/monitor-value */ "./src/main/js/model/monitor-value.ts");
var monitor_binding_1 = __webpack_require__(/*! ../monitor-binding */ "./src/main/js/controller/monitor-binding.ts");
var graph_1 = __webpack_require__(/*! ../monitor/graph */ "./src/main/js/controller/monitor/graph.ts");
var multi_log_1 = __webpack_require__(/*! ../monitor/multi-log */ "./src/main/js/controller/monitor/multi-log.ts");
var single_log_1 = __webpack_require__(/*! ../monitor/single-log */ "./src/main/js/controller/monitor/single-log.ts");
function createFormatter() {
    // TODO: formatter precision
    return new number_1.default(2);
}
/**
 * @hidden
 */
function createTextMonitor(document, target, params) {
    var value = new monitor_value_1.default(type_util_1.default.getOrDefault(params.count, 1));
    var controller = value.totalCount === 1
        ? new single_log_1.default(document, {
            formatter: createFormatter(),
            value: value,
        })
        : new multi_log_1.default(document, {
            formatter: createFormatter(),
            value: value,
        });
    var ticker = new interval_1.default(document, type_util_1.default.getOrDefault(params.interval, 200));
    return new monitor_binding_1.default(document, {
        binding: new monitor_1.default({
            reader: NumberConverter.fromMixed,
            target: target,
            ticker: ticker,
            value: value,
        }),
        controller: controller,
        label: params.label || target.key,
    });
}
exports.createTextMonitor = createTextMonitor;
/**
 * @hidden
 */
function createGraphMonitor(document, target, params) {
    var value = new monitor_value_1.default(type_util_1.default.getOrDefault(params.count, 64));
    var ticker = new interval_1.default(document, type_util_1.default.getOrDefault(params.interval, 200));
    return new monitor_binding_1.default(document, {
        binding: new monitor_1.default({
            reader: NumberConverter.fromMixed,
            target: target,
            ticker: ticker,
            value: value,
        }),
        controller: new graph_1.default(document, {
            formatter: createFormatter(),
            maxValue: type_util_1.default.getOrDefault(params.max, 100),
            minValue: type_util_1.default.getOrDefault(params.min, 0),
            value: value,
        }),
        label: params.label || target.key,
    });
}
exports.createGraphMonitor = createGraphMonitor;


/***/ }),

/***/ "./src/main/js/controller/binding-creators/string-input.ts":
/*!*****************************************************************!*\
  !*** ./src/main/js/controller/binding-creators/string-input.ts ***!
  \*****************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var input_1 = __webpack_require__(/*! ../../binding/input */ "./src/main/js/binding/input.ts");
var composite_1 = __webpack_require__(/*! ../../constraint/composite */ "./src/main/js/constraint/composite.ts");
var list_1 = __webpack_require__(/*! ../../constraint/list */ "./src/main/js/constraint/list.ts");
var util_1 = __webpack_require__(/*! ../../constraint/util */ "./src/main/js/constraint/util.ts");
var StringConverter = __webpack_require__(/*! ../../converter/string */ "./src/main/js/converter/string.ts");
var string_1 = __webpack_require__(/*! ../../formatter/string */ "./src/main/js/formatter/string.ts");
var input_value_1 = __webpack_require__(/*! ../../model/input-value */ "./src/main/js/model/input-value.ts");
var input_binding_1 = __webpack_require__(/*! ../input-binding */ "./src/main/js/controller/input-binding.ts");
var list_2 = __webpack_require__(/*! ../input/list */ "./src/main/js/controller/input/list.ts");
var text_1 = __webpack_require__(/*! ../input/text */ "./src/main/js/controller/input/text.ts");
function createConstraint(params) {
    var constraints = [];
    if (params.options) {
        constraints.push(new list_1.default({
            options: params.options,
        }));
    }
    return new composite_1.default({
        constraints: constraints,
    });
}
function createController(document, value) {
    var c = value.constraint;
    if (c && util_1.default.findConstraint(c, list_1.default)) {
        return new list_2.default(document, {
            stringifyValue: StringConverter.toString,
            value: value,
        });
    }
    return new text_1.default(document, {
        formatter: new string_1.default(),
        parser: StringConverter.toString,
        value: value,
    });
}
/**
 * @hidden
 */
function create(document, target, params) {
    var value = new input_value_1.default('', createConstraint(params));
    var binding = new input_1.default({
        reader: StringConverter.fromMixed,
        target: target,
        value: value,
        writer: function (v) { return v; },
    });
    return new input_binding_1.default(document, {
        binding: binding,
        controller: createController(document, value),
        label: params.label || target.key,
    });
}
exports.create = create;


/***/ }),

/***/ "./src/main/js/controller/binding-creators/string-monitor.ts":
/*!*******************************************************************!*\
  !*** ./src/main/js/controller/binding-creators/string-monitor.ts ***!
  \*******************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var monitor_1 = __webpack_require__(/*! ../../binding/monitor */ "./src/main/js/binding/monitor.ts");
var StringConverter = __webpack_require__(/*! ../../converter/string */ "./src/main/js/converter/string.ts");
var string_1 = __webpack_require__(/*! ../../formatter/string */ "./src/main/js/formatter/string.ts");
var interval_1 = __webpack_require__(/*! ../../misc/ticker/interval */ "./src/main/js/misc/ticker/interval.ts");
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var monitor_value_1 = __webpack_require__(/*! ../../model/monitor-value */ "./src/main/js/model/monitor-value.ts");
var monitor_binding_1 = __webpack_require__(/*! ../monitor-binding */ "./src/main/js/controller/monitor-binding.ts");
var multi_log_1 = __webpack_require__(/*! ../monitor/multi-log */ "./src/main/js/controller/monitor/multi-log.ts");
var single_log_1 = __webpack_require__(/*! ../monitor/single-log */ "./src/main/js/controller/monitor/single-log.ts");
/**
 * @hidden
 */
function createTextMonitor(document, target, params) {
    var value = new monitor_value_1.default(type_util_1.default.getOrDefault(params.count, 1));
    var multiline = value.totalCount > 1 || params.multiline;
    var controller = multiline
        ? new multi_log_1.default(document, {
            formatter: new string_1.default(),
            value: value,
        })
        : new single_log_1.default(document, {
            formatter: new string_1.default(),
            value: value,
        });
    var ticker = new interval_1.default(document, type_util_1.default.getOrDefault(params.interval, 200));
    return new monitor_binding_1.default(document, {
        binding: new monitor_1.default({
            reader: StringConverter.fromMixed,
            target: target,
            ticker: ticker,
            value: value,
        }),
        controller: controller,
        label: params.label || target.key,
    });
}
exports.createTextMonitor = createTextMonitor;


/***/ }),

/***/ "./src/main/js/controller/button.ts":
/*!******************************************!*\
  !*** ./src/main/js/controller/button.ts ***!
  \******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var button_1 = __webpack_require__(/*! ../model/button */ "./src/main/js/model/button.ts");
var button_2 = __webpack_require__(/*! ../view/button */ "./src/main/js/view/button.ts");
/**
 * @hidden
 */
var ButtonController = /** @class */ (function () {
    function ButtonController(document, config) {
        this.onButtonClick_ = this.onButtonClick_.bind(this);
        this.button = new button_1.default(config.title);
        this.view = new button_2.default(document, {
            button: this.button,
        });
        this.view.buttonElement.addEventListener('click', this.onButtonClick_);
    }
    ButtonController.prototype.dispose = function () {
        this.view.dispose();
    };
    ButtonController.prototype.onButtonClick_ = function () {
        this.button.click();
    };
    return ButtonController;
}());
exports.default = ButtonController;


/***/ }),

/***/ "./src/main/js/controller/folder.ts":
/*!******************************************!*\
  !*** ./src/main/js/controller/folder.ts ***!
  \******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var DomUtil = __webpack_require__(/*! ../misc/dom-util */ "./src/main/js/misc/dom-util.ts");
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
var type_util_1 = __webpack_require__(/*! ../misc/type-util */ "./src/main/js/misc/type-util.ts");
var folder_1 = __webpack_require__(/*! ../model/folder */ "./src/main/js/model/folder.ts");
var list_1 = __webpack_require__(/*! ../model/list */ "./src/main/js/model/list.ts");
var folder_2 = __webpack_require__(/*! ../view/folder */ "./src/main/js/view/folder.ts");
var input_binding_1 = __webpack_require__(/*! ./input-binding */ "./src/main/js/controller/input-binding.ts");
var monitor_binding_1 = __webpack_require__(/*! ./monitor-binding */ "./src/main/js/controller/monitor-binding.ts");
/**
 * @hidden
 */
var FolderController = /** @class */ (function () {
    function FolderController(document, config) {
        this.onFolderChange_ = this.onFolderChange_.bind(this);
        this.onInputChange_ = this.onInputChange_.bind(this);
        this.onMonitorUpdate_ = this.onMonitorUpdate_.bind(this);
        this.onTitleClick_ = this.onTitleClick_.bind(this);
        this.onUiControllerListAppend_ = this.onUiControllerListAppend_.bind(this);
        this.emitter = new emitter_1.default();
        this.folder = new folder_1.default(config.title, type_util_1.default.getOrDefault(config.expanded, true));
        this.folder.emitter.on('change', this.onFolderChange_);
        this.ucList_ = new list_1.default();
        this.ucList_.emitter.on('append', this.onUiControllerListAppend_);
        this.doc_ = document;
        this.view = new folder_2.default(this.doc_, {
            folder: this.folder,
        });
        this.view.titleElement.addEventListener('click', this.onTitleClick_);
    }
    Object.defineProperty(FolderController.prototype, "document", {
        get: function () {
            return this.doc_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(FolderController.prototype, "uiControllerList", {
        get: function () {
            return this.ucList_;
        },
        enumerable: true,
        configurable: true
    });
    FolderController.prototype.dispose = function () {
        this.view.dispose();
    };
    FolderController.prototype.computeExpandedHeight_ = function () {
        var _this = this;
        var elem = this.view.containerElement;
        var height = 0;
        DomUtil.disableTransitionTemporarily(elem, function () {
            // Expand folder
            var expanded = _this.folder.expanded;
            _this.folder.expandedHeight = null;
            _this.folder.expanded = true;
            DomUtil.forceReflow(elem);
            // Compute height
            height = elem.getBoundingClientRect().height;
            // Restore expanded
            _this.folder.expanded = expanded;
            DomUtil.forceReflow(elem);
        });
        return height;
    };
    FolderController.prototype.onTitleClick_ = function () {
        this.folder.expanded = !this.folder.expanded;
    };
    FolderController.prototype.onUiControllerListAppend_ = function (uc) {
        if (uc instanceof input_binding_1.default) {
            var emitter = uc.binding.value.emitter;
            emitter.on('change', this.onInputChange_);
        }
        else if (uc instanceof monitor_binding_1.default) {
            var emitter = uc.binding.value.emitter;
            emitter.on('update', this.onMonitorUpdate_);
        }
        this.view.containerElement.appendChild(uc.view.element);
        this.folder.expandedHeight = this.computeExpandedHeight_();
    };
    FolderController.prototype.onInputChange_ = function (value) {
        this.emitter.emit('inputchange', [value]);
    };
    FolderController.prototype.onMonitorUpdate_ = function (value) {
        this.emitter.emit('monitorupdate', [value]);
    };
    FolderController.prototype.onFolderChange_ = function () {
        this.emitter.emit('fold');
    };
    return FolderController;
}());
exports.default = FolderController;


/***/ }),

/***/ "./src/main/js/controller/input-binding.ts":
/*!*************************************************!*\
  !*** ./src/main/js/controller/input-binding.ts ***!
  \*************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var labeled_1 = __webpack_require__(/*! ../view/labeled */ "./src/main/js/view/labeled.ts");
/**
 * @hidden
 */
var InputBindingController = /** @class */ (function () {
    function InputBindingController(document, config) {
        this.binding = config.binding;
        this.controller = config.controller;
        this.view = new labeled_1.default(document, {
            label: config.label,
            view: this.controller.view,
        });
    }
    InputBindingController.prototype.dispose = function () {
        this.controller.dispose();
        this.view.dispose();
    };
    return InputBindingController;
}());
exports.default = InputBindingController;


/***/ }),

/***/ "./src/main/js/controller/input/checkbox.ts":
/*!**************************************************!*\
  !*** ./src/main/js/controller/input/checkbox.ts ***!
  \**************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var checkbox_1 = __webpack_require__(/*! ../../view/input/checkbox */ "./src/main/js/view/input/checkbox.ts");
/**
 * @hidden
 */
var CheckboxInputController = /** @class */ (function () {
    function CheckboxInputController(document, config) {
        this.onInputChange_ = this.onInputChange_.bind(this);
        this.value = config.value;
        this.view = new checkbox_1.default(document, {
            value: this.value,
        });
        this.view.inputElement.addEventListener('change', this.onInputChange_);
    }
    CheckboxInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    CheckboxInputController.prototype.onInputChange_ = function (e) {
        var inputElem = type_util_1.default.forceCast(e.currentTarget);
        this.value.rawValue = inputElem.checked;
        this.view.update();
    };
    return CheckboxInputController;
}());
exports.default = CheckboxInputController;


/***/ }),

/***/ "./src/main/js/controller/input/color-picker.ts":
/*!******************************************************!*\
  !*** ./src/main/js/controller/input/color-picker.ts ***!
  \******************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var composite_1 = __webpack_require__(/*! ../../constraint/composite */ "./src/main/js/constraint/composite.ts");
var range_1 = __webpack_require__(/*! ../../constraint/range */ "./src/main/js/constraint/range.ts");
var step_1 = __webpack_require__(/*! ../../constraint/step */ "./src/main/js/constraint/step.ts");
var number_1 = __webpack_require__(/*! ../../formatter/number */ "./src/main/js/formatter/number.ts");
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var color_1 = __webpack_require__(/*! ../../model/color */ "./src/main/js/model/color.ts");
var foldable_1 = __webpack_require__(/*! ../../model/foldable */ "./src/main/js/model/foldable.ts");
var input_value_1 = __webpack_require__(/*! ../../model/input-value */ "./src/main/js/model/input-value.ts");
var number_2 = __webpack_require__(/*! ../../parser/number */ "./src/main/js/parser/number.ts");
var color_picker_1 = __webpack_require__(/*! ../../view/input/color-picker */ "./src/main/js/view/input/color-picker.ts");
var h_palette_1 = __webpack_require__(/*! ./h-palette */ "./src/main/js/controller/input/h-palette.ts");
var number_text_1 = __webpack_require__(/*! ./number-text */ "./src/main/js/controller/input/number-text.ts");
var sv_palette_1 = __webpack_require__(/*! ./sv-palette */ "./src/main/js/controller/input/sv-palette.ts");
var COMPONENT_CONSTRAINT = new composite_1.default({
    constraints: [
        new range_1.default({
            max: 255,
            min: 0,
        }),
        new step_1.default({
            step: 1,
        }),
    ],
});
/**
 * @hidden
 */
var ColorPickerInputController = /** @class */ (function () {
    function ColorPickerInputController(document, config) {
        var _this = this;
        this.onInputBlur_ = this.onInputBlur_.bind(this);
        this.onValueChange_ = this.onValueChange_.bind(this);
        this.value = config.value;
        this.value.emitter.on('change', this.onValueChange_);
        this.foldable = new foldable_1.default();
        this.hPaletteIc_ = new h_palette_1.default(document, {
            value: this.value,
        });
        this.svPaletteIc_ = new sv_palette_1.default(document, {
            value: this.value,
        });
        var initialComps = this.value.rawValue.getComponents();
        var rgbValues = [0, 1, 2].map(function (index) {
            return new input_value_1.default(initialComps[index], COMPONENT_CONSTRAINT);
        });
        rgbValues.forEach(function (compValue, index) {
            compValue.emitter.on('change', function (rawValue) {
                var comps = _this.value.rawValue.getComponents();
                if (index === 0 || index === 1 || index === 2) {
                    comps[index] = rawValue;
                }
                _this.value.rawValue = new (color_1.default.bind.apply(color_1.default, [void 0].concat(comps)))();
            });
        });
        this.rgbIcs_ = rgbValues.map(function (compValue) {
            return new number_text_1.default(document, {
                formatter: new number_1.default(0),
                parser: number_2.default,
                value: compValue,
            });
        });
        this.view = new color_picker_1.default(document, {
            foldable: this.foldable,
            hPaletteInputView: this.hPaletteIc_.view,
            rgbInputViews: this.rgbIcs_.map(function (ic) {
                return ic.view;
            }),
            svPaletteInputView: this.svPaletteIc_.view,
            value: this.value,
        });
        this.view.allFocusableElements.forEach(function (elem) {
            elem.addEventListener('blur', _this.onInputBlur_);
        });
    }
    ColorPickerInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    ColorPickerInputController.prototype.onInputBlur_ = function (e) {
        var elem = this.view.element;
        var nextTarget = type_util_1.default.forceCast(e.relatedTarget);
        if (!nextTarget || !elem.contains(nextTarget)) {
            this.foldable.expanded = false;
        }
    };
    ColorPickerInputController.prototype.onValueChange_ = function () {
        var comps = this.value.rawValue.getComponents();
        this.rgbIcs_.forEach(function (ic, index) {
            ic.value.rawValue = comps[index];
        });
    };
    return ColorPickerInputController;
}());
exports.default = ColorPickerInputController;


/***/ }),

/***/ "./src/main/js/controller/input/color-swatch-text.ts":
/*!***********************************************************!*\
  !*** ./src/main/js/controller/input/color-swatch-text.ts ***!
  \***********************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var color_swatch_text_1 = __webpack_require__(/*! ../../view/input/color-swatch-text */ "./src/main/js/view/input/color-swatch-text.ts");
var color_swatch_1 = __webpack_require__(/*! ../input/color-swatch */ "./src/main/js/controller/input/color-swatch.ts");
var text_1 = __webpack_require__(/*! ./text */ "./src/main/js/controller/input/text.ts");
/**
 * @hidden
 */
var ColorSwatchTextInputController = /** @class */ (function () {
    function ColorSwatchTextInputController(document, config) {
        this.value = config.value;
        this.swatchIc_ = new color_swatch_1.default(document, {
            value: this.value,
        });
        this.textIc_ = new text_1.default(document, {
            formatter: config.formatter,
            parser: config.parser,
            value: this.value,
        });
        this.view = new color_swatch_text_1.default(document, {
            swatchInputView: this.swatchIc_.view,
            textInputView: this.textIc_.view,
        });
    }
    ColorSwatchTextInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    return ColorSwatchTextInputController;
}());
exports.default = ColorSwatchTextInputController;


/***/ }),

/***/ "./src/main/js/controller/input/color-swatch.ts":
/*!******************************************************!*\
  !*** ./src/main/js/controller/input/color-swatch.ts ***!
  \******************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var color_swatch_1 = __webpack_require__(/*! ../../view/input/color-swatch */ "./src/main/js/view/input/color-swatch.ts");
var color_picker_1 = __webpack_require__(/*! ./color-picker */ "./src/main/js/controller/input/color-picker.ts");
/**
 * @hidden
 */
var ColorSwatchInputController = /** @class */ (function () {
    function ColorSwatchInputController(document, config) {
        this.onButtonBlur_ = this.onButtonBlur_.bind(this);
        this.onButtonClick_ = this.onButtonClick_.bind(this);
        this.value = config.value;
        this.pickerIc_ = new color_picker_1.default(document, {
            value: this.value,
        });
        this.view = new color_swatch_1.default(document, {
            pickerInputView: this.pickerIc_.view,
            value: this.value,
        });
        this.view.buttonElement.addEventListener('blur', this.onButtonBlur_);
        this.view.buttonElement.addEventListener('click', this.onButtonClick_);
    }
    ColorSwatchInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    ColorSwatchInputController.prototype.onButtonBlur_ = function (e) {
        var elem = this.view.element;
        var nextTarget = type_util_1.default.forceCast(e.relatedTarget);
        if (!nextTarget || !elem.contains(nextTarget)) {
            this.pickerIc_.foldable.expanded = false;
        }
    };
    ColorSwatchInputController.prototype.onButtonClick_ = function () {
        this.pickerIc_.foldable.expanded = !this.pickerIc_.foldable.expanded;
    };
    return ColorSwatchInputController;
}());
exports.default = ColorSwatchInputController;


/***/ }),

/***/ "./src/main/js/controller/input/h-palette.ts":
/*!***************************************************!*\
  !*** ./src/main/js/controller/input/h-palette.ts ***!
  \***************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var ColorModel = __webpack_require__(/*! ../../misc/color-model */ "./src/main/js/misc/color-model.ts");
var number_util_1 = __webpack_require__(/*! ../../misc/number-util */ "./src/main/js/misc/number-util.ts");
var pointer_handler_1 = __webpack_require__(/*! ../../misc/pointer-handler */ "./src/main/js/misc/pointer-handler.ts");
var color_1 = __webpack_require__(/*! ../../model/color */ "./src/main/js/model/color.ts");
var h_palette_1 = __webpack_require__(/*! ../../view/input/h-palette */ "./src/main/js/view/input/h-palette.ts");
/**
 * @hidden
 */
var HPaletteInputController = /** @class */ (function () {
    function HPaletteInputController(document, config) {
        this.onPointerDown_ = this.onPointerDown_.bind(this);
        this.onPointerMove_ = this.onPointerMove_.bind(this);
        this.onPointerUp_ = this.onPointerUp_.bind(this);
        this.value = config.value;
        this.view = new h_palette_1.default(document, {
            value: this.value,
        });
        this.ptHandler_ = new pointer_handler_1.default(document, this.view.canvasElement);
        this.ptHandler_.emitter.on('down', this.onPointerDown_);
        this.ptHandler_.emitter.on('move', this.onPointerMove_);
        this.ptHandler_.emitter.on('up', this.onPointerUp_);
    }
    HPaletteInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    HPaletteInputController.prototype.handlePointerEvent_ = function (d) {
        var hue = number_util_1.default.map(d.py, 0, 1, 0, 360);
        var c = this.value.rawValue;
        var _a = ColorModel.rgbToHsv.apply(ColorModel, c.getComponents()), s = _a[1], v = _a[2];
        this.value.rawValue = new (color_1.default.bind.apply(color_1.default, [void 0].concat(ColorModel.hsvToRgb(hue, s, v))))();
        this.view.update();
    };
    HPaletteInputController.prototype.onPointerDown_ = function (d) {
        this.handlePointerEvent_(d);
    };
    HPaletteInputController.prototype.onPointerMove_ = function (d) {
        this.handlePointerEvent_(d);
    };
    HPaletteInputController.prototype.onPointerUp_ = function (d) {
        this.handlePointerEvent_(d);
    };
    return HPaletteInputController;
}());
exports.default = HPaletteInputController;


/***/ }),

/***/ "./src/main/js/controller/input/list.ts":
/*!**********************************************!*\
  !*** ./src/main/js/controller/input/list.ts ***!
  \**********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var list_1 = __webpack_require__(/*! ../../constraint/list */ "./src/main/js/constraint/list.ts");
var util_1 = __webpack_require__(/*! ../../constraint/util */ "./src/main/js/constraint/util.ts");
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var list_2 = __webpack_require__(/*! ../../view/input/list */ "./src/main/js/view/input/list.ts");
function findListItems(value) {
    var c = value.constraint
        ? util_1.default.findConstraint(value.constraint, list_1.default)
        : null;
    if (!c) {
        return null;
    }
    return c.options;
}
/**
 * @hidden
 */
var ListInputController = /** @class */ (function () {
    function ListInputController(document, config) {
        this.onSelectChange_ = this.onSelectChange_.bind(this);
        this.value_ = config.value;
        this.listItems_ = findListItems(this.value_) || [];
        this.view_ = new list_2.default(document, {
            options: this.listItems_,
            stringifyValue: config.stringifyValue,
            value: this.value_,
        });
        this.view_.selectElement.addEventListener('change', this.onSelectChange_);
    }
    Object.defineProperty(ListInputController.prototype, "value", {
        get: function () {
            return this.value_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(ListInputController.prototype, "view", {
        get: function () {
            return this.view_;
        },
        enumerable: true,
        configurable: true
    });
    ListInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    ListInputController.prototype.onSelectChange_ = function (e) {
        var selectElem = type_util_1.default.forceCast(e.currentTarget);
        var optElem = selectElem.selectedOptions.item(0);
        if (!optElem) {
            return;
        }
        var itemIndex = Number(optElem.dataset.index);
        this.value_.rawValue = this.listItems_[itemIndex].value;
        this.view_.update();
    };
    return ListInputController;
}());
exports.default = ListInputController;


/***/ }),

/***/ "./src/main/js/controller/input/number-text.ts":
/*!*****************************************************!*\
  !*** ./src/main/js/controller/input/number-text.ts ***!
  \*****************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var step_1 = __webpack_require__(/*! ../../constraint/step */ "./src/main/js/constraint/step.ts");
var util_1 = __webpack_require__(/*! ../../constraint/util */ "./src/main/js/constraint/util.ts");
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var text_1 = __webpack_require__(/*! ./text */ "./src/main/js/controller/input/text.ts");
function findStep(value) {
    var c = value.constraint
        ? util_1.default.findConstraint(value.constraint, step_1.default)
        : null;
    if (!c) {
        return null;
    }
    return c.step;
}
function estimateSuitableStep(value) {
    var step = findStep(value);
    return type_util_1.default.getOrDefault(step, 1);
}
/**
 * @hidden
 */
var NumberTextInputController = /** @class */ (function (_super) {
    __extends(NumberTextInputController, _super);
    function NumberTextInputController(document, config) {
        var _this = _super.call(this, document, config) || this;
        _this.onInputKeyDown_ = _this.onInputKeyDown_.bind(_this);
        _this.step_ = estimateSuitableStep(_this.value);
        _this.view.inputElement.addEventListener('keydown', _this.onInputKeyDown_);
        return _this;
    }
    NumberTextInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    NumberTextInputController.prototype.onInputKeyDown_ = function (e) {
        var step = this.step_ * (e.altKey ? 0.1 : 1) * (e.shiftKey ? 10 : 1);
        if (e.keyCode === 38) {
            this.value.rawValue += step;
            this.view.update();
        }
        else if (e.keyCode === 40) {
            this.value.rawValue -= step;
            this.view.update();
        }
    };
    return NumberTextInputController;
}(text_1.default));
exports.default = NumberTextInputController;


/***/ }),

/***/ "./src/main/js/controller/input/slider-text.ts":
/*!*****************************************************!*\
  !*** ./src/main/js/controller/input/slider-text.ts ***!
  \*****************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var slider_text_1 = __webpack_require__(/*! ../../view/input/slider-text */ "./src/main/js/view/input/slider-text.ts");
var number_text_1 = __webpack_require__(/*! ./number-text */ "./src/main/js/controller/input/number-text.ts");
var slider_1 = __webpack_require__(/*! ./slider */ "./src/main/js/controller/input/slider.ts");
/**
 * @hidden
 */
var SliderTextInputController = /** @class */ (function () {
    function SliderTextInputController(document, config) {
        this.value_ = config.value;
        this.sliderIc_ = new slider_1.default(document, {
            value: config.value,
        });
        this.textIc_ = new number_text_1.default(document, {
            formatter: config.formatter,
            parser: config.parser,
            value: config.value,
        });
        this.view_ = new slider_text_1.default(document, {
            sliderInputView: this.sliderIc_.view,
            textInputView: this.textIc_.view,
        });
    }
    Object.defineProperty(SliderTextInputController.prototype, "value", {
        get: function () {
            return this.value_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SliderTextInputController.prototype, "view", {
        get: function () {
            return this.view_;
        },
        enumerable: true,
        configurable: true
    });
    SliderTextInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    return SliderTextInputController;
}());
exports.default = SliderTextInputController;


/***/ }),

/***/ "./src/main/js/controller/input/slider.ts":
/*!************************************************!*\
  !*** ./src/main/js/controller/input/slider.ts ***!
  \************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var range_1 = __webpack_require__(/*! ../../constraint/range */ "./src/main/js/constraint/range.ts");
var util_1 = __webpack_require__(/*! ../../constraint/util */ "./src/main/js/constraint/util.ts");
var number_util_1 = __webpack_require__(/*! ../../misc/number-util */ "./src/main/js/misc/number-util.ts");
var pointer_handler_1 = __webpack_require__(/*! ../../misc/pointer-handler */ "./src/main/js/misc/pointer-handler.ts");
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var slider_1 = __webpack_require__(/*! ../../view/input/slider */ "./src/main/js/view/input/slider.ts");
function findRange(value) {
    var c = value.constraint
        ? util_1.default.findConstraint(value.constraint, range_1.default)
        : null;
    if (!c) {
        return [undefined, undefined];
    }
    return [c.minValue, c.maxValue];
}
function estimateSuitableRange(value) {
    var _a = findRange(value), min = _a[0], max = _a[1];
    return [
        type_util_1.default.getOrDefault(min, 0),
        type_util_1.default.getOrDefault(max, 100),
    ];
}
/**
 * @hidden
 */
var SliderInputController = /** @class */ (function () {
    function SliderInputController(document, config) {
        this.onPointerDown_ = this.onPointerDown_.bind(this);
        this.onPointerMove_ = this.onPointerMove_.bind(this);
        this.onPointerUp_ = this.onPointerUp_.bind(this);
        this.value = config.value;
        var _a = estimateSuitableRange(this.value), min = _a[0], max = _a[1];
        this.minValue_ = min;
        this.maxValue_ = max;
        this.view = new slider_1.default(document, {
            maxValue: this.maxValue_,
            minValue: this.minValue_,
            value: this.value,
        });
        this.ptHandler_ = new pointer_handler_1.default(document, this.view.outerElement);
        this.ptHandler_.emitter.on('down', this.onPointerDown_);
        this.ptHandler_.emitter.on('move', this.onPointerMove_);
        this.ptHandler_.emitter.on('up', this.onPointerUp_);
    }
    SliderInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    SliderInputController.prototype.onPointerDown_ = function (d) {
        this.value.rawValue = number_util_1.default.map(d.px, 0, 1, this.minValue_, this.maxValue_);
        this.view.update();
    };
    SliderInputController.prototype.onPointerMove_ = function (d) {
        this.value.rawValue = number_util_1.default.map(d.px, 0, 1, this.minValue_, this.maxValue_);
        this.view.update();
    };
    SliderInputController.prototype.onPointerUp_ = function (d) {
        this.value.rawValue = number_util_1.default.map(d.px, 0, 1, this.minValue_, this.maxValue_);
        this.view.update();
    };
    return SliderInputController;
}());
exports.default = SliderInputController;


/***/ }),

/***/ "./src/main/js/controller/input/sv-palette.ts":
/*!****************************************************!*\
  !*** ./src/main/js/controller/input/sv-palette.ts ***!
  \****************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var ColorModel = __webpack_require__(/*! ../../misc/color-model */ "./src/main/js/misc/color-model.ts");
var number_util_1 = __webpack_require__(/*! ../../misc/number-util */ "./src/main/js/misc/number-util.ts");
var pointer_handler_1 = __webpack_require__(/*! ../../misc/pointer-handler */ "./src/main/js/misc/pointer-handler.ts");
var color_1 = __webpack_require__(/*! ../../model/color */ "./src/main/js/model/color.ts");
var sv_palette_1 = __webpack_require__(/*! ../../view/input/sv-palette */ "./src/main/js/view/input/sv-palette.ts");
/**
 * @hidden
 */
var SvPaletteInputController = /** @class */ (function () {
    function SvPaletteInputController(document, config) {
        this.onPointerDown_ = this.onPointerDown_.bind(this);
        this.onPointerMove_ = this.onPointerMove_.bind(this);
        this.onPointerUp_ = this.onPointerUp_.bind(this);
        this.value = config.value;
        this.view = new sv_palette_1.default(document, {
            value: this.value,
        });
        this.ptHandler_ = new pointer_handler_1.default(document, this.view.canvasElement);
        this.ptHandler_.emitter.on('down', this.onPointerDown_);
        this.ptHandler_.emitter.on('move', this.onPointerMove_);
        this.ptHandler_.emitter.on('up', this.onPointerUp_);
    }
    SvPaletteInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    SvPaletteInputController.prototype.handlePointerEvent_ = function (d) {
        var saturation = number_util_1.default.map(d.px, 0, 1, 0, 100);
        var value = number_util_1.default.map(d.py, 0, 1, 100, 0);
        var c = this.value.rawValue;
        var h = ColorModel.rgbToHsv.apply(ColorModel, c.getComponents())[0];
        this.value.rawValue = new (color_1.default.bind.apply(color_1.default, [void 0].concat(ColorModel.hsvToRgb(h, saturation, value))))();
        this.view.update();
    };
    SvPaletteInputController.prototype.onPointerDown_ = function (d) {
        this.handlePointerEvent_(d);
    };
    SvPaletteInputController.prototype.onPointerMove_ = function (d) {
        this.handlePointerEvent_(d);
    };
    SvPaletteInputController.prototype.onPointerUp_ = function (d) {
        this.handlePointerEvent_(d);
    };
    return SvPaletteInputController;
}());
exports.default = SvPaletteInputController;


/***/ }),

/***/ "./src/main/js/controller/input/text.ts":
/*!**********************************************!*\
  !*** ./src/main/js/controller/input/text.ts ***!
  \**********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var type_util_1 = __webpack_require__(/*! ../../misc/type-util */ "./src/main/js/misc/type-util.ts");
var text_1 = __webpack_require__(/*! ../../view/input/text */ "./src/main/js/view/input/text.ts");
/**
 * @hidden
 */
var TextInputController = /** @class */ (function () {
    function TextInputController(document, config) {
        this.onInputChange_ = this.onInputChange_.bind(this);
        this.parser_ = config.parser;
        this.value = config.value;
        this.view = new text_1.default(document, {
            formatter: config.formatter,
            value: this.value,
        });
        this.view.inputElement.addEventListener('change', this.onInputChange_);
    }
    TextInputController.prototype.dispose = function () {
        this.view.dispose();
    };
    TextInputController.prototype.onInputChange_ = function (e) {
        var _this = this;
        var inputElem = type_util_1.default.forceCast(e.currentTarget);
        var value = inputElem.value;
        type_util_1.default.ifNotEmpty(this.parser_(value), function (parsedValue) {
            _this.value.rawValue = parsedValue;
        });
        this.view.update();
    };
    return TextInputController;
}());
exports.default = TextInputController;


/***/ }),

/***/ "./src/main/js/controller/monitor-binding.ts":
/*!***************************************************!*\
  !*** ./src/main/js/controller/monitor-binding.ts ***!
  \***************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var labeled_1 = __webpack_require__(/*! ../view/labeled */ "./src/main/js/view/labeled.ts");
/**
 * @hidden
 */
var MonitorBindingController = /** @class */ (function () {
    function MonitorBindingController(document, config) {
        this.binding = config.binding;
        this.controller = config.controller;
        this.view = new labeled_1.default(document, {
            label: config.label,
            view: this.controller.view,
        });
    }
    MonitorBindingController.prototype.dispose = function () {
        this.binding.dispose();
        this.controller.dispose();
        this.view.dispose();
    };
    return MonitorBindingController;
}());
exports.default = MonitorBindingController;


/***/ }),

/***/ "./src/main/js/controller/monitor/graph.ts":
/*!*************************************************!*\
  !*** ./src/main/js/controller/monitor/graph.ts ***!
  \*************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var number_util_1 = __webpack_require__(/*! ../../misc/number-util */ "./src/main/js/misc/number-util.ts");
var graph_cursor_1 = __webpack_require__(/*! ../../model/graph-cursor */ "./src/main/js/model/graph-cursor.ts");
var graph_1 = __webpack_require__(/*! ../../view/monitor/graph */ "./src/main/js/view/monitor/graph.ts");
/**
 * @hidden
 */
var GraphMonitorController = /** @class */ (function () {
    function GraphMonitorController(document, config) {
        this.onGraphMouseLeave_ = this.onGraphMouseLeave_.bind(this);
        this.onGraphMouseMove_ = this.onGraphMouseMove_.bind(this);
        this.value = config.value;
        this.cursor_ = new graph_cursor_1.default();
        this.view = new graph_1.default(document, {
            cursor: this.cursor_,
            formatter: config.formatter,
            maxValue: config.maxValue,
            minValue: config.minValue,
            value: this.value,
        });
        this.view.graphElement.addEventListener('mouseleave', this.onGraphMouseLeave_);
        this.view.graphElement.addEventListener('mousemove', this.onGraphMouseMove_);
    }
    GraphMonitorController.prototype.dispose = function () {
        this.view.dispose();
    };
    GraphMonitorController.prototype.onGraphMouseLeave_ = function () {
        this.cursor_.index = -1;
    };
    GraphMonitorController.prototype.onGraphMouseMove_ = function (e) {
        var bounds = this.view.graphElement.getBoundingClientRect();
        var x = e.offsetX;
        this.cursor_.index = Math.floor(number_util_1.default.map(x, 0, bounds.width, 0, this.value.totalCount));
    };
    return GraphMonitorController;
}());
exports.default = GraphMonitorController;


/***/ }),

/***/ "./src/main/js/controller/monitor/multi-log.ts":
/*!*****************************************************!*\
  !*** ./src/main/js/controller/monitor/multi-log.ts ***!
  \*****************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var multi_log_1 = __webpack_require__(/*! ../../view/monitor/multi-log */ "./src/main/js/view/monitor/multi-log.ts");
/**
 * @hidden
 */
var MultiLogMonitorController = /** @class */ (function () {
    function MultiLogMonitorController(document, config) {
        this.value = config.value;
        this.view = new multi_log_1.default(document, {
            formatter: config.formatter,
            value: this.value,
        });
    }
    MultiLogMonitorController.prototype.dispose = function () {
        this.view.dispose();
    };
    return MultiLogMonitorController;
}());
exports.default = MultiLogMonitorController;


/***/ }),

/***/ "./src/main/js/controller/monitor/single-log.ts":
/*!******************************************************!*\
  !*** ./src/main/js/controller/monitor/single-log.ts ***!
  \******************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var single_log_1 = __webpack_require__(/*! ../../view/monitor/single-log */ "./src/main/js/view/monitor/single-log.ts");
/**
 * @hidden
 */
var SingleLogMonitorController = /** @class */ (function () {
    function SingleLogMonitorController(document, config) {
        this.value = config.value;
        this.view = new single_log_1.default(document, {
            formatter: config.formatter,
            value: this.value,
        });
    }
    SingleLogMonitorController.prototype.dispose = function () {
        this.view.dispose();
    };
    return SingleLogMonitorController;
}());
exports.default = SingleLogMonitorController;


/***/ }),

/***/ "./src/main/js/controller/root.ts":
/*!****************************************!*\
  !*** ./src/main/js/controller/root.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
var type_util_1 = __webpack_require__(/*! ../misc/type-util */ "./src/main/js/misc/type-util.ts");
var folder_1 = __webpack_require__(/*! ../model/folder */ "./src/main/js/model/folder.ts");
var list_1 = __webpack_require__(/*! ../model/list */ "./src/main/js/model/list.ts");
var root_1 = __webpack_require__(/*! ../view/root */ "./src/main/js/view/root.ts");
var folder_2 = __webpack_require__(/*! ./folder */ "./src/main/js/controller/folder.ts");
var input_binding_1 = __webpack_require__(/*! ./input-binding */ "./src/main/js/controller/input-binding.ts");
var monitor_binding_1 = __webpack_require__(/*! ./monitor-binding */ "./src/main/js/controller/monitor-binding.ts");
function createFolder(config) {
    if (!config.title) {
        return null;
    }
    return new folder_1.default(config.title, type_util_1.default.getOrDefault(config.expanded, true));
}
/**
 * @hidden
 */
var RootController = /** @class */ (function () {
    function RootController(document, config) {
        this.onFolderChange_ = this.onFolderChange_.bind(this);
        this.onRootFolderChange_ = this.onRootFolderChange_.bind(this);
        this.onTitleClick_ = this.onTitleClick_.bind(this);
        this.onUiControllerListAppend_ = this.onUiControllerListAppend_.bind(this);
        this.onInputChange_ = this.onInputChange_.bind(this);
        this.onMonitorUpdate_ = this.onMonitorUpdate_.bind(this);
        this.emitter = new emitter_1.default();
        this.folder = createFolder(config);
        this.ucList_ = new list_1.default();
        this.ucList_.emitter.on('append', this.onUiControllerListAppend_);
        this.doc_ = document;
        this.view = new root_1.default(this.doc_, {
            folder: this.folder,
        });
        if (this.view.titleElement) {
            this.view.titleElement.addEventListener('click', this.onTitleClick_);
        }
        if (this.folder) {
            this.folder.emitter.on('change', this.onRootFolderChange_);
        }
    }
    Object.defineProperty(RootController.prototype, "document", {
        get: function () {
            return this.doc_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(RootController.prototype, "uiControllerList", {
        get: function () {
            return this.ucList_;
        },
        enumerable: true,
        configurable: true
    });
    RootController.prototype.dispose = function () {
        this.view.dispose();
    };
    RootController.prototype.onUiControllerListAppend_ = function (uc) {
        if (uc instanceof input_binding_1.default) {
            var emitter = uc.binding.value.emitter;
            emitter.on('change', this.onInputChange_);
        }
        else if (uc instanceof monitor_binding_1.default) {
            var emitter = uc.binding.value.emitter;
            emitter.on('update', this.onMonitorUpdate_);
        }
        else if (uc instanceof folder_2.default) {
            var emitter = uc.emitter;
            emitter.on('fold', this.onFolderChange_);
            emitter.on('inputchange', this.onInputChange_);
            emitter.on('monitorupdate', this.onMonitorUpdate_);
        }
        this.view.containerElement.appendChild(uc.view.element);
    };
    RootController.prototype.onTitleClick_ = function () {
        if (this.folder) {
            this.folder.expanded = !this.folder.expanded;
        }
    };
    RootController.prototype.onInputChange_ = function (value) {
        this.emitter.emit('inputchange', [value]);
    };
    RootController.prototype.onMonitorUpdate_ = function (value) {
        this.emitter.emit('monitorupdate', [value]);
    };
    RootController.prototype.onFolderChange_ = function () {
        this.emitter.emit('fold');
    };
    RootController.prototype.onRootFolderChange_ = function () {
        this.emitter.emit('fold');
    };
    return RootController;
}());
exports.default = RootController;


/***/ }),

/***/ "./src/main/js/controller/separator.ts":
/*!*********************************************!*\
  !*** ./src/main/js/controller/separator.ts ***!
  \*********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var separator_1 = __webpack_require__(/*! ../view/separator */ "./src/main/js/view/separator.ts");
/**
 * @hidden
 */
var SeparatorController = /** @class */ (function () {
    function SeparatorController(document) {
        this.view = new separator_1.default(document);
    }
    SeparatorController.prototype.dispose = function () {
        this.view.dispose();
    };
    return SeparatorController;
}());
exports.default = SeparatorController;


/***/ }),

/***/ "./src/main/js/controller/ui-util.ts":
/*!*******************************************!*\
  !*** ./src/main/js/controller/ui-util.ts ***!
  \*******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var folder_1 = __webpack_require__(/*! ./folder */ "./src/main/js/controller/folder.ts");
/**
 * @hidden
 */
function findControllers(uiControllers, controllerClass) {
    return uiControllers.reduce(function (results, uc) {
        if (uc instanceof folder_1.default) {
            // eslint-disable-next-line no-use-before-define
            results.push.apply(results, findControllers(uc.uiControllerList.items, controllerClass));
        }
        if (uc instanceof controllerClass) {
            results.push(uc);
        }
        return results;
    }, []);
}
exports.findControllers = findControllers;


/***/ }),

/***/ "./src/main/js/converter/boolean.ts":
/*!******************************************!*\
  !*** ./src/main/js/converter/boolean.ts ***!
  \******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
function fromMixed(value) {
    if (value === 'false') {
        return false;
    }
    return !!value;
}
exports.fromMixed = fromMixed;
/**
 * @hidden
 */
function toString(value) {
    return String(value);
}
exports.toString = toString;


/***/ }),

/***/ "./src/main/js/converter/color.ts":
/*!****************************************!*\
  !*** ./src/main/js/converter/color.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var number_util_1 = __webpack_require__(/*! ../misc/number-util */ "./src/main/js/misc/number-util.ts");
var color_1 = __webpack_require__(/*! ../model/color */ "./src/main/js/model/color.ts");
var color_2 = __webpack_require__(/*! ../parser/color */ "./src/main/js/parser/color.ts");
/**
 * @hidden
 */
function fromMixed(value) {
    if (typeof value === 'string') {
        var cv = color_2.default(value);
        if (cv) {
            return cv;
        }
    }
    return new color_1.default(0, 0, 0);
}
exports.fromMixed = fromMixed;
/**
 * @hidden
 */
function toString(value) {
    var hexes = value
        .getComponents()
        .map(function (comp) {
        var hex = number_util_1.default.constrain(Math.floor(comp), 0, 255).toString(16);
        return hex.length === 1 ? "0" + hex : hex;
    })
        .join('');
    return "#" + hexes;
}
exports.toString = toString;


/***/ }),

/***/ "./src/main/js/converter/number.ts":
/*!*****************************************!*\
  !*** ./src/main/js/converter/number.ts ***!
  \*****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var number_1 = __webpack_require__(/*! ../parser/number */ "./src/main/js/parser/number.ts");
/**
 * @hidden
 */
function fromMixed(value) {
    if (typeof value === 'number') {
        return value;
    }
    if (typeof value === 'string') {
        var pv = number_1.default(value);
        if (pv !== null && pv !== undefined) {
            return pv;
        }
    }
    return 0;
}
exports.fromMixed = fromMixed;
/**
 * @hidden
 */
function toString(value) {
    return String(value);
}
exports.toString = toString;


/***/ }),

/***/ "./src/main/js/converter/string.ts":
/*!*****************************************!*\
  !*** ./src/main/js/converter/string.ts ***!
  \*****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
function fromMixed(value) {
    return String(value);
}
exports.fromMixed = fromMixed;
/**
 * @hidden
 */
function toString(value) {
    return value;
}
exports.toString = toString;


/***/ }),

/***/ "./src/main/js/formatter/boolean.ts":
/*!******************************************!*\
  !*** ./src/main/js/formatter/boolean.ts ***!
  \******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var BooleanConverter = __webpack_require__(/*! ../converter/boolean */ "./src/main/js/converter/boolean.ts");
/**
 * @hidden
 */
var BooleanFormatter = /** @class */ (function () {
    function BooleanFormatter() {
    }
    BooleanFormatter.prototype.format = function (value) {
        return BooleanConverter.toString(value);
    };
    return BooleanFormatter;
}());
exports.default = BooleanFormatter;


/***/ }),

/***/ "./src/main/js/formatter/color.ts":
/*!****************************************!*\
  !*** ./src/main/js/formatter/color.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var ColorConverter = __webpack_require__(/*! ../converter/color */ "./src/main/js/converter/color.ts");
var number_util_1 = __webpack_require__(/*! ../misc/number-util */ "./src/main/js/misc/number-util.ts");
/**
 * @hidden
 */
var ColorFormatter = /** @class */ (function () {
    function ColorFormatter() {
    }
    ColorFormatter.rgb = function (r, g, b) {
        var compsText = [
            number_util_1.default.constrain(Math.floor(r), 0, 255),
            number_util_1.default.constrain(Math.floor(g), 0, 255),
            number_util_1.default.constrain(Math.floor(b), 0, 255),
        ].join(', ');
        return "rgb(" + compsText + ")";
    };
    ColorFormatter.hsl = function (h, s, l) {
        var compsText = [
            ((Math.floor(h) % 360) + 360) % 360,
            number_util_1.default.constrain(Math.floor(s), 0, 100) + "%",
            number_util_1.default.constrain(Math.floor(l), 0, 100) + "%",
        ].join(', ');
        return "hsl(" + compsText + ")";
    };
    ColorFormatter.prototype.format = function (value) {
        return ColorConverter.toString(value);
    };
    return ColorFormatter;
}());
exports.default = ColorFormatter;


/***/ }),

/***/ "./src/main/js/formatter/number.ts":
/*!*****************************************!*\
  !*** ./src/main/js/formatter/number.ts ***!
  \*****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var NumberFormatter = /** @class */ (function () {
    function NumberFormatter(digits) {
        this.digits_ = digits;
    }
    Object.defineProperty(NumberFormatter.prototype, "digits", {
        get: function () {
            return this.digits_;
        },
        enumerable: true,
        configurable: true
    });
    NumberFormatter.prototype.format = function (value) {
        return value.toFixed(this.digits_);
    };
    return NumberFormatter;
}());
exports.default = NumberFormatter;


/***/ }),

/***/ "./src/main/js/formatter/string.ts":
/*!*****************************************!*\
  !*** ./src/main/js/formatter/string.ts ***!
  \*****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var StringFormatter = /** @class */ (function () {
    function StringFormatter() {
    }
    StringFormatter.prototype.format = function (value) {
        return value;
    };
    return StringFormatter;
}());
exports.default = StringFormatter;


/***/ }),

/***/ "./src/main/js/index.ts":
/*!******************************!*\
  !*** ./src/main/js/index.ts ***!
  \******************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var Style = __webpack_require__(/*! ../sass/bundle.scss */ "./src/main/sass/bundle.scss");
var tweakpane_without_style_1 = __webpack_require__(/*! ./tweakpane-without-style */ "./src/main/js/tweakpane-without-style.ts");
function embedDefaultStyleIfNeeded(document) {
    var MARKER = 'tweakpane';
    if (document.querySelector("style[data-for=" + MARKER + "]")) {
        return;
    }
    var styleElem = document.createElement('style');
    styleElem.dataset.for = MARKER;
    styleElem.textContent = Style.toString();
    if (document.head) {
        document.head.appendChild(styleElem);
    }
}
var Tweakpane = /** @class */ (function (_super) {
    __extends(Tweakpane, _super);
    function Tweakpane(opt_config) {
        var _this = _super.call(this, opt_config) || this;
        embedDefaultStyleIfNeeded(_this.document);
        return _this;
    }
    return Tweakpane;
}(tweakpane_without_style_1.default));
exports.default = Tweakpane;


/***/ }),

/***/ "./src/main/js/misc/class-name.ts":
/*!****************************************!*\
  !*** ./src/main/js/misc/class-name.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var PREFIX = 'tp';
var TYPE_TO_POSTFIX_MAP = {
    '': 'v',
    input: 'iv',
    monitor: 'mv',
};
function className(viewName, opt_viewType) {
    var viewType = opt_viewType || '';
    var postfix = TYPE_TO_POSTFIX_MAP[viewType];
    return function (opt_elementName, opt_modifier) {
        return [
            PREFIX,
            '-',
            viewName,
            postfix,
            opt_elementName ? "_" + opt_elementName : '',
            opt_modifier ? "-" + opt_modifier : '',
        ].join('');
    };
}
exports.default = className;


/***/ }),

/***/ "./src/main/js/misc/color-model.ts":
/*!*****************************************!*\
  !*** ./src/main/js/misc/color-model.ts ***!
  \*****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var number_util_1 = __webpack_require__(/*! ./number-util */ "./src/main/js/misc/number-util.ts");
function rgbToHsl(r, g, b) {
    var rp = number_util_1.default.constrain(r / 255, 0, 1);
    var gp = number_util_1.default.constrain(g / 255, 0, 1);
    var bp = number_util_1.default.constrain(b / 255, 0, 1);
    var cmax = Math.max(rp, gp, bp);
    var cmin = Math.min(rp, gp, bp);
    var c = cmax - cmin;
    var h = 0;
    var s = 0;
    var l = (cmin + cmax) / 2;
    if (c !== 0) {
        s = l > 0.5 ? c / (2 - cmin - cmax) : c / (cmax + cmin);
        if (rp === cmax) {
            h = (gp - bp) / c;
        }
        else if (gp === cmax) {
            h = 2 + (bp - rp) / c;
        }
        else {
            h = 4 + (rp - gp) / c;
        }
        h = h / 6 + (h < 0 ? 1 : 0);
    }
    return [h * 360, s * 100, l * 100];
}
exports.rgbToHsl = rgbToHsl;
function hslToRgb(h, s, l) {
    var _a, _b, _c, _d, _e, _f;
    var hp = ((h % 360) + 360) % 360;
    var sp = number_util_1.default.constrain(s / 100, 0, 1);
    var lp = number_util_1.default.constrain(l / 100, 0, 1);
    var c = (1 - Math.abs(2 * lp - 1)) * sp;
    var x = c * (1 - Math.abs(((hp / 60) % 2) - 1));
    var m = lp - c / 2;
    var rp, gp, bp;
    if (hp >= 0 && hp < 60) {
        _a = [c, x, 0], rp = _a[0], gp = _a[1], bp = _a[2];
    }
    else if (hp >= 60 && hp < 120) {
        _b = [x, c, 0], rp = _b[0], gp = _b[1], bp = _b[2];
    }
    else if (hp >= 120 && hp < 180) {
        _c = [0, c, x], rp = _c[0], gp = _c[1], bp = _c[2];
    }
    else if (hp >= 180 && hp < 240) {
        _d = [0, x, c], rp = _d[0], gp = _d[1], bp = _d[2];
    }
    else if (hp >= 240 && hp < 300) {
        _e = [x, 0, c], rp = _e[0], gp = _e[1], bp = _e[2];
    }
    else {
        _f = [c, 0, x], rp = _f[0], gp = _f[1], bp = _f[2];
    }
    return [(rp + m) * 255, (gp + m) * 255, (bp + m) * 255];
}
exports.hslToRgb = hslToRgb;
function rgbToHsv(r, g, b) {
    var rp = number_util_1.default.constrain(r / 255, 0, 1);
    var gp = number_util_1.default.constrain(g / 255, 0, 1);
    var bp = number_util_1.default.constrain(b / 255, 0, 1);
    var cmax = Math.max(rp, gp, bp);
    var cmin = Math.min(rp, gp, bp);
    var d = cmax - cmin;
    var h;
    if (d === 0) {
        h = 0;
    }
    else if (cmax === rp) {
        h = 60 * (((((gp - bp) / d) % 6) + 6) % 6);
    }
    else if (cmax === gp) {
        h = 60 * ((bp - rp) / d + 2);
    }
    else {
        h = 60 * ((rp - gp) / d + 4);
    }
    var s = cmax === 0 ? 0 : d / cmax;
    var v = cmax;
    return [h, s * 100, v * 100];
}
exports.rgbToHsv = rgbToHsv;
function hsvToRgb(h, s, v) {
    var _a, _b, _c, _d, _e, _f;
    var hp = ((h % 360) + 360) % 360;
    var sp = number_util_1.default.constrain(s / 100, 0, 1);
    var vp = number_util_1.default.constrain(v / 100, 0, 1);
    var c = vp * sp;
    var x = c * (1 - Math.abs(((hp / 60) % 2) - 1));
    var m = vp - c;
    var rp, gp, bp;
    if (hp >= 0 && hp < 60) {
        _a = [c, x, 0], rp = _a[0], gp = _a[1], bp = _a[2];
    }
    else if (hp >= 60 && hp < 120) {
        _b = [x, c, 0], rp = _b[0], gp = _b[1], bp = _b[2];
    }
    else if (hp >= 120 && hp < 180) {
        _c = [0, c, x], rp = _c[0], gp = _c[1], bp = _c[2];
    }
    else if (hp >= 180 && hp < 240) {
        _d = [0, x, c], rp = _d[0], gp = _d[1], bp = _d[2];
    }
    else if (hp >= 240 && hp < 300) {
        _e = [x, 0, c], rp = _e[0], gp = _e[1], bp = _e[2];
    }
    else {
        _f = [c, 0, x], rp = _f[0], gp = _f[1], bp = _f[2];
    }
    return [(rp + m) * 255, (gp + m) * 255, (bp + m) * 255];
}
exports.hsvToRgb = hsvToRgb;


/***/ }),

/***/ "./src/main/js/misc/disposing-util.ts":
/*!********************************************!*\
  !*** ./src/main/js/misc/disposing-util.ts ***!
  \********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
function disposeElement(elem) {
    if (elem && elem.parentElement) {
        elem.parentElement.removeChild(elem);
    }
    return null;
}
exports.disposeElement = disposeElement;


/***/ }),

/***/ "./src/main/js/misc/dom-util.ts":
/*!**************************************!*\
  !*** ./src/main/js/misc/dom-util.ts ***!
  \**************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/* WEBPACK VAR INJECTION */(function(process) {
Object.defineProperty(exports, "__esModule", { value: true });
var type_util_1 = __webpack_require__(/*! ./type-util */ "./src/main/js/misc/type-util.ts");
function forceReflow(element) {
    // tslint:disable-next-line:no-unused-expression
    element.offsetHeight;
}
exports.forceReflow = forceReflow;
function disableTransitionTemporarily(element, callback) {
    var t = element.style.transition;
    element.style.transition = 'none';
    callback();
    element.style.transition = t;
}
exports.disableTransitionTemporarily = disableTransitionTemporarily;
function supportsTouch(document) {
    return document.ontouchstart !== undefined;
}
exports.supportsTouch = supportsTouch;
function getWindowDocument() {
    // tslint:disable-next-line:function-constructor
    var globalObj = type_util_1.default.forceCast(new Function('return this')());
    return globalObj.document;
}
exports.getWindowDocument = getWindowDocument;
function isBrowser() {
    // Webpack defines process.browser = true;
    // https://github.com/webpack/node-libs-browser
    // https://github.com/defunctzombie/node-process
    return !!process.browser;
}
function getCanvasContext(canvasElement) {
    // HTMLCanvasElement.prototype.getContext is not defined on testing environment
    return isBrowser() ? canvasElement.getContext('2d') : null;
}
exports.getCanvasContext = getCanvasContext;

/* WEBPACK VAR INJECTION */}.call(this, __webpack_require__(/*! ./../../../../node_modules/process/browser.js */ "./node_modules/process/browser.js")))

/***/ }),

/***/ "./src/main/js/misc/emitter.ts":
/*!*************************************!*\
  !*** ./src/main/js/misc/emitter.ts ***!
  \*************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var Emitter = /** @class */ (function () {
    function Emitter() {
        this.observers_ = {};
    }
    Emitter.prototype.on = function (eventName, handler) {
        var observers = this.observers_[eventName];
        if (!observers) {
            observers = this.observers_[eventName] = [];
        }
        observers.push({
            handler: handler,
        });
        return this;
    };
    Emitter.prototype.off = function (eventName, handler) {
        var observers = this.observers_[eventName];
        if (observers) {
            this.observers_[eventName] = observers.filter(function (observer) {
                return observer.handler !== handler;
            });
        }
        return this;
    };
    Emitter.prototype.emit = function (eventName, opt_args) {
        var observers = this.observers_[eventName];
        if (!observers) {
            return;
        }
        observers.forEach(function (observer) {
            var handlerArgs = opt_args || [];
            observer.handler.apply(observer, handlerArgs);
        });
    };
    return Emitter;
}());
exports.default = Emitter;


/***/ }),

/***/ "./src/main/js/misc/number-util.ts":
/*!*****************************************!*\
  !*** ./src/main/js/misc/number-util.ts ***!
  \*****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var NumberUtil = {
    map: function (value, start1, end1, start2, end2) {
        var p = (value - start1) / (end1 - start1);
        return start2 + p * (end2 - start2);
    },
    getDecimalDigits: function (value) {
        var text = String(value.toFixed(10));
        var frac = text.split('.')[1];
        return frac.replace(/0+$/, '').length;
    },
    constrain: function (value, min, max) {
        return Math.min(Math.max(value, min), max);
    },
};
exports.default = NumberUtil;


/***/ }),

/***/ "./src/main/js/misc/pane-error.ts":
/*!****************************************!*\
  !*** ./src/main/js/misc/pane-error.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
function createMessage(config) {
    if (config.type === 'alreadydisposed') {
        return 'View has been already disposed';
    }
    if (config.type === 'emptyvalue') {
        return "Value is empty for " + config.context.key;
    }
    if (config.type === 'invalidparams') {
        return "Invalid parameters for " + config.context.name;
    }
    if (config.type === 'nomatchingcontroller') {
        return "No matching controller for " + config.context.key;
    }
    if (config.type === 'shouldneverhappen') {
        return 'This error should never happen';
    }
    return 'Unexpected error';
}
var PaneError = /** @class */ (function () {
    function PaneError(config) {
        this.message = createMessage(config);
        this.name = this.constructor.name;
        this.stack = new Error(this.message).stack;
        this.type = config.type;
    }
    PaneError.alreadyDisposed = function () {
        return new PaneError({ type: 'alreadydisposed' });
    };
    PaneError.shouldNeverHappen = function () {
        return new PaneError({ type: 'shouldneverhappen' });
    };
    return PaneError;
}());
exports.default = PaneError;
PaneError.prototype = Object.create(Error.prototype);
PaneError.prototype.constructor = PaneError;


/***/ }),

/***/ "./src/main/js/misc/pointer-handler.ts":
/*!*********************************************!*\
  !*** ./src/main/js/misc/pointer-handler.ts ***!
  \*********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var DomUtil = __webpack_require__(/*! ./dom-util */ "./src/main/js/misc/dom-util.ts");
var emitter_1 = __webpack_require__(/*! ./emitter */ "./src/main/js/misc/emitter.ts");
/**
 * A utility class to handle both mouse and touch events.
 * @hidden
 */
var PointerHandler = /** @class */ (function () {
    function PointerHandler(document, element) {
        this.onDocumentMouseMove_ = this.onDocumentMouseMove_.bind(this);
        this.onDocumentMouseUp_ = this.onDocumentMouseUp_.bind(this);
        this.onMouseDown_ = this.onMouseDown_.bind(this);
        this.onTouchMove_ = this.onTouchMove_.bind(this);
        this.onTouchStart_ = this.onTouchStart_.bind(this);
        this.document = document;
        this.element = element;
        this.emitter = new emitter_1.default();
        this.pressed_ = false;
        if (DomUtil.supportsTouch(this.document)) {
            element.addEventListener('touchstart', this.onTouchStart_);
            element.addEventListener('touchmove', this.onTouchMove_);
        }
        else {
            element.addEventListener('mousedown', this.onMouseDown_);
            this.document.addEventListener('mousemove', this.onDocumentMouseMove_);
            this.document.addEventListener('mouseup', this.onDocumentMouseUp_);
        }
    }
    PointerHandler.prototype.computePosition_ = function (offsetX, offsetY) {
        var rect = this.element.getBoundingClientRect();
        return {
            px: offsetX / rect.width,
            py: offsetY / rect.height,
        };
    };
    PointerHandler.prototype.onMouseDown_ = function (e) {
        // Prevent native text selection
        e.preventDefault();
        this.pressed_ = true;
        this.emitter.emit('down', [this.computePosition_(e.offsetX, e.offsetY)]);
    };
    PointerHandler.prototype.onDocumentMouseMove_ = function (e) {
        if (!this.pressed_) {
            return;
        }
        var win = this.document.defaultView;
        var rect = this.element.getBoundingClientRect();
        this.emitter.emit('move', [
            this.computePosition_(e.pageX - (((win && win.scrollX) || 0) + rect.left), e.pageY - (((win && win.scrollY) || 0) + rect.top)),
        ]);
    };
    PointerHandler.prototype.onDocumentMouseUp_ = function (e) {
        if (!this.pressed_) {
            return;
        }
        this.pressed_ = false;
        var win = this.document.defaultView;
        var rect = this.element.getBoundingClientRect();
        this.emitter.emit('up', [
            this.computePosition_(e.pageX - (((win && win.scrollX) || 0) + rect.left), e.pageY - (((win && win.scrollY) || 0) + rect.top)),
        ]);
    };
    PointerHandler.prototype.onTouchStart_ = function (e) {
        // Prevent native page scroll
        e.preventDefault();
        var touch = e.targetTouches[0];
        var rect = this.element.getBoundingClientRect();
        this.emitter.emit('down', [
            this.computePosition_(touch.clientX - rect.left, touch.clientY - rect.top),
        ]);
    };
    PointerHandler.prototype.onTouchMove_ = function (e) {
        var touch = e.targetTouches[0];
        var rect = this.element.getBoundingClientRect();
        this.emitter.emit('move', [
            this.computePosition_(touch.clientX - rect.left, touch.clientY - rect.top),
        ]);
    };
    return PointerHandler;
}());
exports.default = PointerHandler;


/***/ }),

/***/ "./src/main/js/misc/ticker/interval.ts":
/*!*********************************************!*\
  !*** ./src/main/js/misc/ticker/interval.ts ***!
  \*********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../emitter */ "./src/main/js/misc/emitter.ts");
/**
 * @hidden
 */
var IntervalTicker = /** @class */ (function () {
    function IntervalTicker(document, interval) {
        var _this = this;
        this.onTick_ = this.onTick_.bind(this);
        this.onWindowBlur_ = this.onWindowBlur_.bind(this);
        this.onWindowFocus_ = this.onWindowFocus_.bind(this);
        this.active_ = true;
        this.doc_ = document;
        this.emitter = new emitter_1.default();
        if (interval <= 0) {
            this.id_ = null;
        }
        else {
            var win = this.doc_.defaultView;
            if (win) {
                this.id_ = win.setInterval(function () {
                    if (!_this.active_) {
                        return;
                    }
                    _this.onTick_();
                }, interval);
            }
        }
        // TODO: Stop on blur?
        // const win = document.defaultView;
        // if (win) {
        //   win.addEventListener('blur', this.onWindowBlur_);
        //   win.addEventListener('focus', this.onWindowFocus_);
        // }
    }
    IntervalTicker.prototype.dispose = function () {
        if (this.id_ !== null) {
            var win = this.doc_.defaultView;
            if (win) {
                win.clearInterval(this.id_);
            }
        }
        this.id_ = null;
    };
    IntervalTicker.prototype.onTick_ = function () {
        this.emitter.emit('tick');
    };
    IntervalTicker.prototype.onWindowBlur_ = function () {
        this.active_ = false;
    };
    IntervalTicker.prototype.onWindowFocus_ = function () {
        this.active_ = true;
    };
    return IntervalTicker;
}());
exports.default = IntervalTicker;


/***/ }),

/***/ "./src/main/js/misc/type-util.ts":
/*!***************************************!*\
  !*** ./src/main/js/misc/type-util.ts ***!
  \***************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var TypeUtil = {
    forceCast: function (v) {
        return v;
    },
    getOrDefault: function (value, defaultValue) {
        return value !== null && value !== undefined ? value : defaultValue;
    },
    ifNotEmpty: function (value, thenFn, elseFn) {
        if (value !== null && value !== undefined) {
            thenFn(value);
        }
        else if (elseFn) {
            elseFn();
        }
    },
};
exports.default = TypeUtil;


/***/ }),

/***/ "./src/main/js/model/button.ts":
/*!*************************************!*\
  !*** ./src/main/js/model/button.ts ***!
  \*************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
/**
 * @hidden
 */
var Button = /** @class */ (function () {
    function Button(title) {
        this.emitter = new emitter_1.default();
        this.title = title;
    }
    Button.prototype.click = function () {
        this.emitter.emit('click');
    };
    return Button;
}());
exports.default = Button;


/***/ }),

/***/ "./src/main/js/model/color.ts":
/*!************************************!*\
  !*** ./src/main/js/model/color.ts ***!
  \************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
var number_util_1 = __webpack_require__(/*! ../misc/number-util */ "./src/main/js/misc/number-util.ts");
function constrainComponent(comp) {
    return number_util_1.default.constrain(comp, 0, 255);
}
/**
 * @hidden
 */
var Color = /** @class */ (function () {
    function Color(r, g, b) {
        this.emitter = new emitter_1.default();
        this.comps_ = [
            constrainComponent(r),
            constrainComponent(g),
            constrainComponent(b),
        ];
    }
    Color.prototype.getComponents = function () {
        return this.comps_;
    };
    Color.prototype.toObject = function () {
        // tslint:disable:object-literal-sort-keys
        return {
            r: this.comps_[0],
            g: this.comps_[1],
            b: this.comps_[2],
        };
        // tslint:enable:object-literal-sort-keys
    };
    return Color;
}());
exports.default = Color;


/***/ }),

/***/ "./src/main/js/model/foldable.ts":
/*!***************************************!*\
  !*** ./src/main/js/model/foldable.ts ***!
  \***************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
/**
 * @hidden
 */
var Foldable = /** @class */ (function () {
    function Foldable() {
        this.emitter = new emitter_1.default();
        this.expanded_ = false;
    }
    Object.defineProperty(Foldable.prototype, "expanded", {
        get: function () {
            return this.expanded_;
        },
        set: function (expanded) {
            var changed = this.expanded_ !== expanded;
            if (changed) {
                this.expanded_ = expanded;
                this.emitter.emit('change');
            }
        },
        enumerable: true,
        configurable: true
    });
    return Foldable;
}());
exports.default = Foldable;


/***/ }),

/***/ "./src/main/js/model/folder.ts":
/*!*************************************!*\
  !*** ./src/main/js/model/folder.ts ***!
  \*************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
/**
 * @hidden
 */
var Folder = /** @class */ (function () {
    function Folder(title, expanded) {
        this.emitter = new emitter_1.default();
        this.expanded_ = expanded;
        this.expandedHeight_ = null;
        this.title = title;
    }
    Object.defineProperty(Folder.prototype, "expanded", {
        get: function () {
            return this.expanded_;
        },
        set: function (expanded) {
            var changed = this.expanded_ !== expanded;
            if (changed) {
                this.expanded_ = expanded;
                this.emitter.emit('change');
            }
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(Folder.prototype, "expandedHeight", {
        get: function () {
            return this.expandedHeight_;
        },
        set: function (expandedHeight) {
            var changed = this.expandedHeight_ !== expandedHeight;
            if (changed) {
                this.expandedHeight_ = expandedHeight;
                this.emitter.emit('change');
            }
        },
        enumerable: true,
        configurable: true
    });
    return Folder;
}());
exports.default = Folder;


/***/ }),

/***/ "./src/main/js/model/graph-cursor.ts":
/*!*******************************************!*\
  !*** ./src/main/js/model/graph-cursor.ts ***!
  \*******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
/**
 * @hidden
 */
var GraphCursor = /** @class */ (function () {
    function GraphCursor() {
        this.emitter = new emitter_1.default();
        this.index_ = -1;
    }
    Object.defineProperty(GraphCursor.prototype, "index", {
        get: function () {
            return this.index_;
        },
        set: function (index) {
            var changed = this.index_ !== index;
            if (changed) {
                this.index_ = index;
                this.emitter.emit('change', [index]);
            }
        },
        enumerable: true,
        configurable: true
    });
    return GraphCursor;
}());
exports.default = GraphCursor;


/***/ }),

/***/ "./src/main/js/model/input-value.ts":
/*!******************************************!*\
  !*** ./src/main/js/model/input-value.ts ***!
  \******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
/**
 * @hidden
 */
var InputValue = /** @class */ (function () {
    function InputValue(initialValue, constraint) {
        this.constraint_ = constraint;
        this.emitter = new emitter_1.default();
        this.rawValue_ = initialValue;
    }
    InputValue.equalsValue = function (v1, v2) {
        return v1 === v2;
    };
    Object.defineProperty(InputValue.prototype, "constraint", {
        get: function () {
            return this.constraint_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(InputValue.prototype, "rawValue", {
        get: function () {
            return this.rawValue_;
        },
        set: function (rawValue) {
            var constrainedValue = this.constraint_
                ? this.constraint_.constrain(rawValue)
                : rawValue;
            var changed = !InputValue.equalsValue(this.rawValue_, constrainedValue);
            if (changed) {
                this.rawValue_ = constrainedValue;
                this.emitter.emit('change', [constrainedValue]);
            }
        },
        enumerable: true,
        configurable: true
    });
    return InputValue;
}());
exports.default = InputValue;


/***/ }),

/***/ "./src/main/js/model/list.ts":
/*!***********************************!*\
  !*** ./src/main/js/model/list.ts ***!
  \***********************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
/**
 * @hidden
 */
var List = /** @class */ (function () {
    function List() {
        this.emitter = new emitter_1.default();
        this.items_ = [];
    }
    Object.defineProperty(List.prototype, "items", {
        get: function () {
            return this.items_;
        },
        enumerable: true,
        configurable: true
    });
    List.prototype.append = function (item) {
        this.items_.push(item);
        this.emitter.emit('append', [item]);
    };
    return List;
}());
exports.default = List;


/***/ }),

/***/ "./src/main/js/model/monitor-value.ts":
/*!********************************************!*\
  !*** ./src/main/js/model/monitor-value.ts ***!
  \********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var emitter_1 = __webpack_require__(/*! ../misc/emitter */ "./src/main/js/misc/emitter.ts");
/**
 * @hidden
 */
var MonitorValue = /** @class */ (function () {
    function MonitorValue(totalCount) {
        this.emitter = new emitter_1.default();
        this.rawValues_ = [];
        this.totalCount_ = totalCount;
    }
    Object.defineProperty(MonitorValue.prototype, "rawValues", {
        get: function () {
            return this.rawValues_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(MonitorValue.prototype, "totalCount", {
        get: function () {
            return this.totalCount_;
        },
        enumerable: true,
        configurable: true
    });
    MonitorValue.prototype.append = function (rawValue) {
        this.rawValues_.push(rawValue);
        if (this.rawValues_.length > this.totalCount_) {
            this.rawValues_.splice(0, this.rawValues_.length - this.totalCount_);
        }
        this.emitter.emit('update', [rawValue]);
    };
    return MonitorValue;
}());
exports.default = MonitorValue;


/***/ }),

/***/ "./src/main/js/model/target.ts":
/*!*************************************!*\
  !*** ./src/main/js/model/target.ts ***!
  \*************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var type_util_1 = __webpack_require__(/*! ../misc/type-util */ "./src/main/js/misc/type-util.ts");
/**
 * @hidden
 */
var Target = /** @class */ (function () {
    function Target(object, key, opt_id) {
        this.obj_ = object;
        this.key_ = key;
        this.presetKey_ = type_util_1.default.getOrDefault(opt_id, key);
    }
    Object.defineProperty(Target.prototype, "key", {
        get: function () {
            return this.key_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(Target.prototype, "presetKey", {
        get: function () {
            return this.presetKey_;
        },
        enumerable: true,
        configurable: true
    });
    Target.prototype.read = function () {
        return this.obj_[this.key_];
    };
    Target.prototype.write = function (value) {
        this.obj_[this.key_] = value;
    };
    return Target;
}());
exports.default = Target;


/***/ }),

/***/ "./src/main/js/parser/color.ts":
/*!*************************************!*\
  !*** ./src/main/js/parser/color.ts ***!
  \*************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var color_1 = __webpack_require__(/*! ../model/color */ "./src/main/js/model/color.ts");
var SUB_PARSERS = [
    // #aabbcc
    function (text) {
        var matches = text.match(/^#?([0-9A-Fa-f]{2})([0-9A-Fa-f]{2})([0-9A-Fa-f]{2})$/);
        if (!matches) {
            return null;
        }
        return new color_1.default(parseInt(matches[1], 16), parseInt(matches[2], 16), parseInt(matches[3], 16));
    },
    // #abc
    function (text) {
        var matches = text.match(/^#?([0-9A-Fa-f])([0-9A-Fa-f])([0-9A-Fa-f])$/);
        if (!matches) {
            return null;
        }
        return new color_1.default(parseInt(matches[1] + matches[1], 16), parseInt(matches[2] + matches[2], 16), parseInt(matches[3] + matches[3], 16));
    },
];
/**
 * @hidden
 */
var ColorParser = function (text) {
    return SUB_PARSERS.reduce(function (result, subparser) {
        return result ? result : subparser(text);
    }, null);
};
exports.default = ColorParser;


/***/ }),

/***/ "./src/main/js/parser/number.ts":
/*!**************************************!*\
  !*** ./src/main/js/parser/number.ts ***!
  \**************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @hidden
 */
var NumberParser = function (text) {
    var num = parseFloat(text);
    if (isNaN(num)) {
        return null;
    }
    return num;
};
exports.default = NumberParser;


/***/ }),

/***/ "./src/main/js/tweakpane-without-style.ts":
/*!************************************************!*\
  !*** ./src/main/js/tweakpane-without-style.ts ***!
  \************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var root_1 = __webpack_require__(/*! ./api/root */ "./src/main/js/api/root.ts");
var root_2 = __webpack_require__(/*! ./controller/root */ "./src/main/js/controller/root.ts");
var class_name_1 = __webpack_require__(/*! ./misc/class-name */ "./src/main/js/misc/class-name.ts");
var DomUtil = __webpack_require__(/*! ./misc/dom-util */ "./src/main/js/misc/dom-util.ts");
var pane_error_1 = __webpack_require__(/*! ./misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var type_util_1 = __webpack_require__(/*! ./misc/type-util */ "./src/main/js/misc/type-util.ts");
function createDefaultWrapperElement(document) {
    var elem = document.createElement('div');
    elem.classList.add(class_name_1.default('dfw')());
    if (document.body) {
        document.body.appendChild(elem);
    }
    return elem;
}
var TweakpaneWithoutStyle = /** @class */ (function (_super) {
    __extends(TweakpaneWithoutStyle, _super);
    function TweakpaneWithoutStyle(opt_config) {
        var _this = this;
        var config = opt_config || {};
        var document = type_util_1.default.getOrDefault(config.document, DomUtil.getWindowDocument());
        var rootController = new root_2.default(document, {
            title: config.title,
        });
        _this = _super.call(this, rootController) || this;
        _this.containerElem_ =
            config.container || createDefaultWrapperElement(document);
        _this.containerElem_.appendChild(_this.element);
        _this.doc_ = document;
        _this.usesDefaultWrapper_ = !config.container;
        return _this;
    }
    TweakpaneWithoutStyle.prototype.dispose = function () {
        var containerElem = this.containerElem_;
        if (!containerElem) {
            throw pane_error_1.default.alreadyDisposed();
        }
        if (this.usesDefaultWrapper_) {
            var parentElem = containerElem.parentElement;
            if (parentElem) {
                parentElem.removeChild(containerElem);
            }
        }
        this.containerElem_ = null;
        this.doc_ = null;
        _super.prototype.dispose.call(this);
    };
    Object.defineProperty(TweakpaneWithoutStyle.prototype, "document", {
        get: function () {
            if (!this.doc_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.doc_;
        },
        enumerable: true,
        configurable: true
    });
    return TweakpaneWithoutStyle;
}(root_1.default));
exports.default = TweakpaneWithoutStyle;


/***/ }),

/***/ "./src/main/js/view/button.ts":
/*!************************************!*\
  !*** ./src/main/js/view/button.ts ***!
  \************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ./view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('btn');
/**
 * @hidden
 */
var ButtonView = /** @class */ (function (_super) {
    __extends(ButtonView, _super);
    function ButtonView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.button = config.button;
        _this.element.classList.add(className());
        var buttonElem = document.createElement('button');
        buttonElem.classList.add(className('b'));
        buttonElem.textContent = _this.button.title;
        _this.element.appendChild(buttonElem);
        _this.buttonElem_ = buttonElem;
        return _this;
    }
    Object.defineProperty(ButtonView.prototype, "buttonElement", {
        get: function () {
            if (!this.buttonElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.buttonElem_;
        },
        enumerable: true,
        configurable: true
    });
    ButtonView.prototype.dispose = function () {
        this.buttonElem_ = DisposingUtil.disposeElement(this.buttonElem_);
        _super.prototype.dispose.call(this);
    };
    return ButtonView;
}(view_1.default));
exports.default = ButtonView;


/***/ }),

/***/ "./src/main/js/view/folder.ts":
/*!************************************!*\
  !*** ./src/main/js/view/folder.ts ***!
  \************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var type_util_1 = __webpack_require__(/*! ../misc/type-util */ "./src/main/js/misc/type-util.ts");
var view_1 = __webpack_require__(/*! ./view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('fld');
/**
 * @hidden
 */
var FolderView = /** @class */ (function (_super) {
    __extends(FolderView, _super);
    function FolderView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onFolderChange_ = _this.onFolderChange_.bind(_this);
        _this.folder_ = config.folder;
        _this.folder_.emitter.on('change', _this.onFolderChange_);
        _this.element.classList.add(className());
        var titleElem = document.createElement('button');
        titleElem.classList.add(className('t'));
        titleElem.textContent = _this.folder_.title;
        _this.element.appendChild(titleElem);
        _this.titleElem_ = titleElem;
        var markElem = document.createElement('div');
        markElem.classList.add(className('m'));
        _this.titleElem_.appendChild(markElem);
        var containerElem = document.createElement('div');
        containerElem.classList.add(className('c'));
        _this.element.appendChild(containerElem);
        _this.containerElem_ = containerElem;
        _this.applyModel_();
        return _this;
    }
    Object.defineProperty(FolderView.prototype, "titleElement", {
        get: function () {
            if (!this.titleElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.titleElem_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(FolderView.prototype, "containerElement", {
        get: function () {
            if (!this.containerElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.containerElem_;
        },
        enumerable: true,
        configurable: true
    });
    FolderView.prototype.dispose = function () {
        this.containerElem_ = DisposingUtil.disposeElement(this.containerElem_);
        this.titleElem_ = DisposingUtil.disposeElement(this.titleElem_);
        _super.prototype.dispose.call(this);
    };
    FolderView.prototype.applyModel_ = function () {
        var containerElem = this.containerElem_;
        if (!containerElem) {
            throw pane_error_1.default.alreadyDisposed();
        }
        var expanded = this.folder_.expanded;
        var expandedClass = className(undefined, 'expanded');
        if (expanded) {
            this.element.classList.add(expandedClass);
        }
        else {
            this.element.classList.remove(expandedClass);
        }
        type_util_1.default.ifNotEmpty(this.folder_.expandedHeight, function (expandedHeight) {
            var containerHeight = expanded ? expandedHeight : 0;
            containerElem.style.height = containerHeight + "px";
        }, function () {
            containerElem.style.height = expanded ? 'auto' : '0px';
        });
    };
    FolderView.prototype.onFolderChange_ = function () {
        this.applyModel_();
    };
    return FolderView;
}(view_1.default));
exports.default = FolderView;


/***/ }),

/***/ "./src/main/js/view/input/checkbox.ts":
/*!********************************************!*\
  !*** ./src/main/js/view/input/checkbox.ts ***!
  \********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('ckb', 'input');
/**
 * @hidden
 */
var CheckboxInputView = /** @class */ (function (_super) {
    __extends(CheckboxInputView, _super);
    function CheckboxInputView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onValueChange_ = _this.onValueChange_.bind(_this);
        _this.element.classList.add(className());
        var labelElem = document.createElement('label');
        labelElem.classList.add(className('l'));
        _this.element.appendChild(labelElem);
        var inputElem = document.createElement('input');
        inputElem.classList.add(className('i'));
        inputElem.type = 'checkbox';
        labelElem.appendChild(inputElem);
        _this.inputElem_ = inputElem;
        var markElem = document.createElement('div');
        markElem.classList.add(className('m'));
        labelElem.appendChild(markElem);
        config.value.emitter.on('change', _this.onValueChange_);
        _this.value = config.value;
        _this.update();
        return _this;
    }
    Object.defineProperty(CheckboxInputView.prototype, "inputElement", {
        get: function () {
            if (!this.inputElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.inputElem_;
        },
        enumerable: true,
        configurable: true
    });
    CheckboxInputView.prototype.dispose = function () {
        this.inputElem_ = DisposingUtil.disposeElement(this.inputElem_);
        _super.prototype.dispose.call(this);
    };
    CheckboxInputView.prototype.update = function () {
        if (!this.inputElem_) {
            throw pane_error_1.default.alreadyDisposed();
        }
        this.inputElem_.checked = this.value.rawValue;
    };
    CheckboxInputView.prototype.onValueChange_ = function () {
        this.update();
    };
    return CheckboxInputView;
}(view_1.default));
exports.default = CheckboxInputView;


/***/ }),

/***/ "./src/main/js/view/input/color-picker.ts":
/*!************************************************!*\
  !*** ./src/main/js/view/input/color-picker.ts ***!
  \************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('clp', 'input');
/**
 * @hidden
 */
var ColorPickerInputView = /** @class */ (function (_super) {
    __extends(ColorPickerInputView, _super);
    function ColorPickerInputView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onFoldableChange_ = _this.onFoldableChange_.bind(_this);
        _this.onValueChange_ = _this.onValueChange_.bind(_this);
        _this.value = config.value;
        _this.value.emitter.on('change', _this.onValueChange_);
        _this.foldable = config.foldable;
        _this.foldable.emitter.on('change', _this.onFoldableChange_);
        _this.element.classList.add(className());
        var plElem = document.createElement('div');
        plElem.classList.add(className('pl'));
        var svElem = document.createElement('div');
        svElem.classList.add(className('sv'));
        _this.svPaletteView_ = config.svPaletteInputView;
        svElem.appendChild(_this.svPaletteView_.element);
        plElem.appendChild(svElem);
        var hElem = document.createElement('div');
        hElem.classList.add(className('h'));
        _this.hPaletteView_ = config.hPaletteInputView;
        hElem.appendChild(_this.hPaletteView_.element);
        plElem.appendChild(hElem);
        _this.element.appendChild(plElem);
        var inputElems = document.createElement('div');
        inputElems.classList.add(className('is'));
        _this.rgbInputViews_ = config.rgbInputViews;
        _this.rgbInputViews_.forEach(function (iv, index) {
            var elem = document.createElement('div');
            elem.classList.add(className('iw'));
            var labelElem = document.createElement('label');
            labelElem.classList.add(className('il'));
            labelElem.textContent = ['R', 'G', 'B'][index];
            elem.appendChild(labelElem);
            elem.appendChild(iv.element);
            inputElems.appendChild(elem);
        });
        _this.element.appendChild(inputElems);
        _this.update();
        return _this;
    }
    Object.defineProperty(ColorPickerInputView.prototype, "allFocusableElements", {
        get: function () {
            return [].concat(this.hPaletteView_.canvasElement, this.svPaletteView_.canvasElement, this.rgbInputViews_.map(function (iv) {
                return iv.inputElement;
            }));
        },
        enumerable: true,
        configurable: true
    });
    ColorPickerInputView.prototype.dispose = function () {
        this.hPaletteView_.dispose();
        this.rgbInputViews_ = [];
        this.svPaletteView_.dispose();
        _super.prototype.dispose.call(this);
    };
    ColorPickerInputView.prototype.update = function () {
        if (this.foldable.expanded) {
            this.element.classList.add(className(undefined, 'expanded'));
        }
        else {
            this.element.classList.remove(className(undefined, 'expanded'));
        }
        this.rgbInputViews_.forEach(function (iv) {
            iv.update();
        });
    };
    ColorPickerInputView.prototype.onValueChange_ = function () {
        this.update();
    };
    ColorPickerInputView.prototype.onFoldableChange_ = function () {
        this.update();
    };
    return ColorPickerInputView;
}(view_1.default));
exports.default = ColorPickerInputView;


/***/ }),

/***/ "./src/main/js/view/input/color-swatch-text.ts":
/*!*****************************************************!*\
  !*** ./src/main/js/view/input/color-swatch-text.ts ***!
  \*****************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('cswtxt', 'input');
/**
 * @hidden
 */
var ColorSwatchTextInputView = /** @class */ (function (_super) {
    __extends(ColorSwatchTextInputView, _super);
    function ColorSwatchTextInputView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.element.classList.add(className());
        var swatchElem = document.createElement('div');
        swatchElem.classList.add(className('s'));
        _this.swatchInputView_ = config.swatchInputView;
        swatchElem.appendChild(_this.swatchInputView_.element);
        _this.element.appendChild(swatchElem);
        var textElem = document.createElement('div');
        textElem.classList.add(className('t'));
        _this.textInputView_ = config.textInputView;
        textElem.appendChild(_this.textInputView_.element);
        _this.element.appendChild(textElem);
        return _this;
    }
    Object.defineProperty(ColorSwatchTextInputView.prototype, "value", {
        get: function () {
            return this.textInputView_.value;
        },
        enumerable: true,
        configurable: true
    });
    ColorSwatchTextInputView.prototype.dispose = function () {
        this.swatchInputView_.dispose();
        this.textInputView_.dispose();
        _super.prototype.dispose.call(this);
    };
    ColorSwatchTextInputView.prototype.update = function () {
        this.swatchInputView_.update();
        this.textInputView_.update();
    };
    return ColorSwatchTextInputView;
}(view_1.default));
exports.default = ColorSwatchTextInputView;


/***/ }),

/***/ "./src/main/js/view/input/color-swatch.ts":
/*!************************************************!*\
  !*** ./src/main/js/view/input/color-swatch.ts ***!
  \************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var ColorConverter = __webpack_require__(/*! ../../converter/color */ "./src/main/js/converter/color.ts");
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('csw', 'input');
/**
 * @hidden
 */
var ColorSwatchInputView = /** @class */ (function (_super) {
    __extends(ColorSwatchInputView, _super);
    function ColorSwatchInputView(document, config) {
        var _this = _super.call(this, document) || this;
        if (_this.element === null) {
            throw pane_error_1.default.alreadyDisposed();
        }
        _this.onValueChange_ = _this.onValueChange_.bind(_this);
        config.value.emitter.on('change', _this.onValueChange_);
        _this.value = config.value;
        _this.element.classList.add(className());
        var swatchElem = document.createElement('div');
        swatchElem.classList.add(className('sw'));
        _this.element.appendChild(swatchElem);
        _this.swatchElem_ = swatchElem;
        var buttonElem = document.createElement('button');
        buttonElem.classList.add(className('b'));
        _this.element.appendChild(buttonElem);
        _this.buttonElem_ = buttonElem;
        var pickerElem = document.createElement('div');
        pickerElem.classList.add(className('p'));
        _this.pickerView_ = config.pickerInputView;
        pickerElem.appendChild(_this.pickerView_.element);
        _this.element.appendChild(pickerElem);
        _this.update();
        return _this;
    }
    Object.defineProperty(ColorSwatchInputView.prototype, "buttonElement", {
        get: function () {
            if (this.buttonElem_ === null) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.buttonElem_;
        },
        enumerable: true,
        configurable: true
    });
    ColorSwatchInputView.prototype.dispose = function () {
        this.pickerView_.dispose();
        this.buttonElem_ = DisposingUtil.disposeElement(this.buttonElem_);
        this.swatchElem_ = DisposingUtil.disposeElement(this.swatchElem_);
        _super.prototype.dispose.call(this);
    };
    ColorSwatchInputView.prototype.update = function () {
        if (!this.swatchElem_) {
            throw pane_error_1.default.alreadyDisposed();
        }
        var value = this.value.rawValue;
        this.swatchElem_.style.backgroundColor = ColorConverter.toString(value);
    };
    ColorSwatchInputView.prototype.onValueChange_ = function () {
        this.update();
    };
    return ColorSwatchInputView;
}(view_1.default));
exports.default = ColorSwatchInputView;


/***/ }),

/***/ "./src/main/js/view/input/h-palette.ts":
/*!*********************************************!*\
  !*** ./src/main/js/view/input/h-palette.ts ***!
  \*********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var color_1 = __webpack_require__(/*! ../../formatter/color */ "./src/main/js/formatter/color.ts");
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var ColorModel = __webpack_require__(/*! ../../misc/color-model */ "./src/main/js/misc/color-model.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var DomUtil = __webpack_require__(/*! ../../misc/dom-util */ "./src/main/js/misc/dom-util.ts");
var number_util_1 = __webpack_require__(/*! ../../misc/number-util */ "./src/main/js/misc/number-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('hpl', 'input');
/**
 * @hidden
 */
var HPaletteInputView = /** @class */ (function (_super) {
    __extends(HPaletteInputView, _super);
    function HPaletteInputView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onValueChange_ = _this.onValueChange_.bind(_this);
        _this.value = config.value;
        _this.value.emitter.on('change', _this.onValueChange_);
        _this.element.classList.add(className());
        var canvasElem = document.createElement('canvas');
        canvasElem.classList.add(className('c'));
        canvasElem.tabIndex = -1;
        _this.element.appendChild(canvasElem);
        _this.canvasElem_ = canvasElem;
        var markerElem = document.createElement('div');
        markerElem.classList.add(className('m'));
        _this.element.appendChild(markerElem);
        _this.markerElem_ = markerElem;
        _this.update();
        return _this;
    }
    Object.defineProperty(HPaletteInputView.prototype, "canvasElement", {
        get: function () {
            if (!this.canvasElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.canvasElem_;
        },
        enumerable: true,
        configurable: true
    });
    HPaletteInputView.prototype.dispose = function () {
        this.canvasElem_ = DisposingUtil.disposeElement(this.canvasElem_);
        this.markerElem_ = DisposingUtil.disposeElement(this.markerElem_);
        _super.prototype.dispose.call(this);
    };
    HPaletteInputView.prototype.update = function () {
        if (!this.markerElem_) {
            throw pane_error_1.default.alreadyDisposed();
        }
        var ctx = DomUtil.getCanvasContext(this.canvasElement);
        if (!ctx) {
            return;
        }
        var width = this.canvasElement.width;
        var height = this.canvasElement.height;
        var cellCount = 64;
        var ch = Math.ceil(height / cellCount);
        for (var iy = 0; iy < cellCount; iy++) {
            var hue = number_util_1.default.map(iy, 0, cellCount - 1, 0, 360);
            var rgbComps = ColorModel.hsvToRgb(hue, 100, 100);
            ctx.fillStyle = color_1.default.rgb.apply(color_1.default, rgbComps);
            var y = Math.floor(number_util_1.default.map(iy, 0, cellCount - 1, 0, height - ch));
            ctx.fillRect(0, y, width, ch);
        }
        var c = this.value.rawValue;
        var hsvComps = ColorModel.rgbToHsv.apply(ColorModel, c.getComponents());
        var top = number_util_1.default.map(hsvComps[0], 0, 360, 0, 100);
        this.markerElem_.style.top = top + "%";
    };
    HPaletteInputView.prototype.onValueChange_ = function () {
        this.update();
    };
    return HPaletteInputView;
}(view_1.default));
exports.default = HPaletteInputView;


/***/ }),

/***/ "./src/main/js/view/input/list.ts":
/*!****************************************!*\
  !*** ./src/main/js/view/input/list.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('lst', 'input');
/**
 * @hidden
 */
var ListInputView = /** @class */ (function (_super) {
    __extends(ListInputView, _super);
    function ListInputView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onValueChange_ = _this.onValueChange_.bind(_this);
        _this.element.classList.add(className());
        _this.stringifyValue_ = config.stringifyValue;
        var selectElem = document.createElement('select');
        selectElem.classList.add(className('s'));
        config.options.forEach(function (item, index) {
            var optionElem = document.createElement('option');
            optionElem.dataset.index = String(index);
            optionElem.textContent = item.text;
            optionElem.value = _this.stringifyValue_(item.value);
            selectElem.appendChild(optionElem);
        });
        _this.element.appendChild(selectElem);
        _this.selectElem_ = selectElem;
        var markElem = document.createElement('div');
        markElem.classList.add(className('m'));
        _this.element.appendChild(markElem);
        config.value.emitter.on('change', _this.onValueChange_);
        _this.value = config.value;
        _this.update();
        return _this;
    }
    Object.defineProperty(ListInputView.prototype, "selectElement", {
        get: function () {
            if (!this.selectElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.selectElem_;
        },
        enumerable: true,
        configurable: true
    });
    ListInputView.prototype.dispose = function () {
        this.selectElem_ = DisposingUtil.disposeElement(this.selectElem_);
        _super.prototype.dispose.call(this);
    };
    ListInputView.prototype.update = function () {
        if (!this.selectElem_) {
            throw pane_error_1.default.alreadyDisposed();
        }
        this.selectElem_.value = this.stringifyValue_(this.value.rawValue);
    };
    ListInputView.prototype.onValueChange_ = function () {
        this.update();
    };
    return ListInputView;
}(view_1.default));
exports.default = ListInputView;


/***/ }),

/***/ "./src/main/js/view/input/slider-text.ts":
/*!***********************************************!*\
  !*** ./src/main/js/view/input/slider-text.ts ***!
  \***********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('sldtxt', 'input');
/**
 * @hidden
 */
var SliderTextInputView = /** @class */ (function (_super) {
    __extends(SliderTextInputView, _super);
    function SliderTextInputView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.element.classList.add(className());
        var sliderElem = document.createElement('div');
        sliderElem.classList.add(className('s'));
        _this.sliderInputView_ = config.sliderInputView;
        sliderElem.appendChild(_this.sliderInputView_.element);
        _this.element.appendChild(sliderElem);
        var textElem = document.createElement('div');
        textElem.classList.add(className('t'));
        _this.textInputView_ = config.textInputView;
        textElem.appendChild(_this.textInputView_.element);
        _this.element.appendChild(textElem);
        return _this;
    }
    Object.defineProperty(SliderTextInputView.prototype, "value", {
        get: function () {
            return this.sliderInputView_.value;
        },
        enumerable: true,
        configurable: true
    });
    SliderTextInputView.prototype.dispose = function () {
        this.sliderInputView_.dispose();
        this.textInputView_.dispose();
        _super.prototype.dispose.call(this);
    };
    SliderTextInputView.prototype.update = function () {
        this.sliderInputView_.update();
        this.textInputView_.update();
    };
    return SliderTextInputView;
}(view_1.default));
exports.default = SliderTextInputView;


/***/ }),

/***/ "./src/main/js/view/input/slider.ts":
/*!******************************************!*\
  !*** ./src/main/js/view/input/slider.ts ***!
  \******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var number_util_1 = __webpack_require__(/*! ../../misc/number-util */ "./src/main/js/misc/number-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('sld', 'input');
/**
 * @hidden
 */
var SliderInputView = /** @class */ (function (_super) {
    __extends(SliderInputView, _super);
    function SliderInputView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onValueChange_ = _this.onValueChange_.bind(_this);
        _this.minValue_ = config.minValue;
        _this.maxValue_ = config.maxValue;
        _this.element.classList.add(className());
        var outerElem = document.createElement('div');
        outerElem.classList.add(className('o'));
        _this.element.appendChild(outerElem);
        _this.outerElem_ = outerElem;
        var innerElem = document.createElement('div');
        innerElem.classList.add(className('i'));
        _this.outerElem_.appendChild(innerElem);
        _this.innerElem_ = innerElem;
        config.value.emitter.on('change', _this.onValueChange_);
        _this.value = config.value;
        _this.update();
        return _this;
    }
    Object.defineProperty(SliderInputView.prototype, "outerElement", {
        get: function () {
            if (!this.outerElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.outerElem_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SliderInputView.prototype, "innerElement", {
        get: function () {
            if (!this.innerElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.innerElem_;
        },
        enumerable: true,
        configurable: true
    });
    SliderInputView.prototype.dispose = function () {
        this.innerElem_ = DisposingUtil.disposeElement(this.innerElem_);
        this.outerElem_ = DisposingUtil.disposeElement(this.outerElem_);
        _super.prototype.dispose.call(this);
    };
    SliderInputView.prototype.update = function () {
        if (!this.innerElem_) {
            throw pane_error_1.default.alreadyDisposed();
        }
        var p = number_util_1.default.map(this.value.rawValue, this.minValue_, this.maxValue_, 0, 100);
        this.innerElem_.style.width = p + "%";
    };
    SliderInputView.prototype.onValueChange_ = function () {
        this.update();
    };
    return SliderInputView;
}(view_1.default));
exports.default = SliderInputView;


/***/ }),

/***/ "./src/main/js/view/input/sv-palette.ts":
/*!**********************************************!*\
  !*** ./src/main/js/view/input/sv-palette.ts ***!
  \**********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var color_1 = __webpack_require__(/*! ../../formatter/color */ "./src/main/js/formatter/color.ts");
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var ColorModel = __webpack_require__(/*! ../../misc/color-model */ "./src/main/js/misc/color-model.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var DomUtil = __webpack_require__(/*! ../../misc/dom-util */ "./src/main/js/misc/dom-util.ts");
var number_util_1 = __webpack_require__(/*! ../../misc/number-util */ "./src/main/js/misc/number-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('svp', 'input');
/**
 * @hidden
 */
var SvPaletteInputView = /** @class */ (function (_super) {
    __extends(SvPaletteInputView, _super);
    function SvPaletteInputView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onValueChange_ = _this.onValueChange_.bind(_this);
        _this.value = config.value;
        _this.value.emitter.on('change', _this.onValueChange_);
        _this.element.classList.add(className());
        var canvasElem = document.createElement('canvas');
        canvasElem.classList.add(className('c'));
        canvasElem.tabIndex = -1;
        _this.element.appendChild(canvasElem);
        _this.canvasElem_ = canvasElem;
        var markerElem = document.createElement('div');
        markerElem.classList.add(className('m'));
        _this.element.appendChild(markerElem);
        _this.markerElem_ = markerElem;
        _this.update();
        return _this;
    }
    Object.defineProperty(SvPaletteInputView.prototype, "canvasElement", {
        get: function () {
            if (!this.canvasElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.canvasElem_;
        },
        enumerable: true,
        configurable: true
    });
    SvPaletteInputView.prototype.dispose = function () {
        this.canvasElem_ = DisposingUtil.disposeElement(this.canvasElem_);
        this.markerElem_ = DisposingUtil.disposeElement(this.markerElem_);
        _super.prototype.dispose.call(this);
    };
    SvPaletteInputView.prototype.update = function () {
        if (!this.markerElem_) {
            throw pane_error_1.default.alreadyDisposed();
        }
        var ctx = DomUtil.getCanvasContext(this.canvasElement);
        if (!ctx) {
            return;
        }
        var c = this.value.rawValue;
        var hsvComps = ColorModel.rgbToHsv.apply(ColorModel, c.getComponents());
        var width = this.canvasElement.width;
        var height = this.canvasElement.height;
        var cellCount = 64;
        var cw = Math.ceil(width / cellCount);
        var ch = Math.ceil(height / cellCount);
        for (var iy = 0; iy < cellCount; iy++) {
            for (var ix = 0; ix < cellCount; ix++) {
                var s = number_util_1.default.map(ix, 0, cellCount - 1, 0, 100);
                var v = number_util_1.default.map(iy, 0, cellCount - 1, 100, 0);
                var rgbComps = ColorModel.hsvToRgb(hsvComps[0], s, v);
                ctx.fillStyle = color_1.default.rgb.apply(color_1.default, rgbComps);
                var x = Math.floor(number_util_1.default.map(ix, 0, cellCount - 1, 0, width - cw));
                var y = Math.floor(number_util_1.default.map(iy, 0, cellCount - 1, 0, height - ch));
                ctx.fillRect(x, y, cw, ch);
            }
        }
        var left = number_util_1.default.map(hsvComps[1], 0, 100, 0, 100);
        this.markerElem_.style.left = left + "%";
        var top = number_util_1.default.map(hsvComps[2], 0, 100, 100, 0);
        this.markerElem_.style.top = top + "%";
    };
    SvPaletteInputView.prototype.onValueChange_ = function () {
        this.update();
    };
    return SvPaletteInputView;
}(view_1.default));
exports.default = SvPaletteInputView;


/***/ }),

/***/ "./src/main/js/view/input/text.ts":
/*!****************************************!*\
  !*** ./src/main/js/view/input/text.ts ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('txt', 'input');
/**
 * @hidden
 */
var TextInputView = /** @class */ (function (_super) {
    __extends(TextInputView, _super);
    function TextInputView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onValueChange_ = _this.onValueChange_.bind(_this);
        _this.formatter_ = config.formatter;
        _this.element.classList.add(className());
        var inputElem = document.createElement('input');
        inputElem.classList.add(className('i'));
        inputElem.type = 'text';
        _this.element.appendChild(inputElem);
        _this.inputElem_ = inputElem;
        config.value.emitter.on('change', _this.onValueChange_);
        _this.value = config.value;
        _this.update();
        return _this;
    }
    Object.defineProperty(TextInputView.prototype, "inputElement", {
        get: function () {
            if (!this.inputElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.inputElem_;
        },
        enumerable: true,
        configurable: true
    });
    TextInputView.prototype.dispose = function () {
        this.inputElem_ = DisposingUtil.disposeElement(this.inputElem_);
        _super.prototype.dispose.call(this);
    };
    TextInputView.prototype.update = function () {
        if (!this.inputElem_) {
            throw pane_error_1.default.alreadyDisposed();
        }
        this.inputElem_.value = this.formatter_.format(this.value.rawValue);
    };
    TextInputView.prototype.onValueChange_ = function () {
        this.update();
    };
    return TextInputView;
}(view_1.default));
exports.default = TextInputView;


/***/ }),

/***/ "./src/main/js/view/labeled.ts":
/*!*************************************!*\
  !*** ./src/main/js/view/labeled.ts ***!
  \*************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../misc/class-name */ "./src/main/js/misc/class-name.ts");
var view_1 = __webpack_require__(/*! ./view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('lbl');
/**
 * @hidden
 */
var LabeledView = /** @class */ (function (_super) {
    __extends(LabeledView, _super);
    function LabeledView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.label = config.label;
        _this.element.classList.add(className());
        var labelElem = document.createElement('div');
        labelElem.classList.add(className('l'));
        labelElem.textContent = _this.label;
        _this.element.appendChild(labelElem);
        var viewElem = document.createElement('div');
        viewElem.classList.add(className('v'));
        viewElem.appendChild(config.view.element);
        _this.element.appendChild(viewElem);
        return _this;
    }
    return LabeledView;
}(view_1.default));
exports.default = LabeledView;


/***/ }),

/***/ "./src/main/js/view/monitor/graph.ts":
/*!*******************************************!*\
  !*** ./src/main/js/view/monitor/graph.ts ***!
  \*******************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var number_util_1 = __webpack_require__(/*! ../../misc/number-util */ "./src/main/js/misc/number-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var SVG_NS = 'http://www.w3.org/2000/svg';
var className = class_name_1.default('grp', 'monitor');
/**
 * @hidden
 */
var GraphMonitorView = /** @class */ (function (_super) {
    __extends(GraphMonitorView, _super);
    function GraphMonitorView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onCursorChange_ = _this.onCursorChange_.bind(_this);
        _this.onValueUpdate_ = _this.onValueUpdate_.bind(_this);
        _this.element.classList.add(className());
        _this.formatter_ = config.formatter;
        _this.minValue_ = config.minValue;
        _this.maxValue_ = config.maxValue;
        _this.cursor_ = config.cursor;
        _this.cursor_.emitter.on('change', _this.onCursorChange_);
        var svgElem = document.createElementNS(SVG_NS, 'svg');
        svgElem.classList.add(className('g'));
        _this.element.appendChild(svgElem);
        _this.svgElem_ = svgElem;
        var lineElem = document.createElementNS(SVG_NS, 'polyline');
        _this.svgElem_.appendChild(lineElem);
        _this.lineElem_ = lineElem;
        var tooltipElem = document.createElement('div');
        tooltipElem.classList.add(className('t'));
        _this.element.appendChild(tooltipElem);
        _this.tooltipElem_ = tooltipElem;
        config.value.emitter.on('update', _this.onValueUpdate_);
        _this.value = config.value;
        _this.update();
        return _this;
    }
    Object.defineProperty(GraphMonitorView.prototype, "graphElement", {
        get: function () {
            if (!this.svgElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.svgElem_;
        },
        enumerable: true,
        configurable: true
    });
    GraphMonitorView.prototype.dispose = function () {
        this.lineElem_ = DisposingUtil.disposeElement(this.lineElem_);
        this.svgElem_ = DisposingUtil.disposeElement(this.svgElem_);
        this.tooltipElem_ = DisposingUtil.disposeElement(this.tooltipElem_);
        _super.prototype.dispose.call(this);
    };
    GraphMonitorView.prototype.update = function () {
        var tooltipElem = this.tooltipElem_;
        if (!this.lineElem_ || !this.svgElem_ || !tooltipElem) {
            throw pane_error_1.default.alreadyDisposed();
        }
        var bounds = this.svgElem_.getBoundingClientRect();
        // Graph
        var maxIndex = this.value.totalCount - 1;
        var min = this.minValue_;
        var max = this.maxValue_;
        this.lineElem_.setAttributeNS(null, 'points', this.value.rawValues
            .map(function (v, index) {
            var x = number_util_1.default.map(index, 0, maxIndex, 0, bounds.width);
            var y = number_util_1.default.map(v, min, max, bounds.height, 0);
            return [x, y].join(',');
        })
            .join(' '));
        // Cursor
        var value = this.value.rawValues[this.cursor_.index];
        if (value === undefined) {
            tooltipElem.classList.remove(className('t', 'valid'));
            return;
        }
        tooltipElem.classList.add(className('t', 'valid'));
        var tx = number_util_1.default.map(this.cursor_.index, 0, maxIndex, 0, bounds.width);
        var ty = number_util_1.default.map(value, min, max, bounds.height, 0);
        tooltipElem.style.left = tx + "px";
        tooltipElem.style.top = ty + "px";
        tooltipElem.textContent = "" + this.formatter_.format(value);
    };
    GraphMonitorView.prototype.onValueUpdate_ = function () {
        this.update();
    };
    GraphMonitorView.prototype.onCursorChange_ = function () {
        this.update();
    };
    return GraphMonitorView;
}(view_1.default));
exports.default = GraphMonitorView;


/***/ }),

/***/ "./src/main/js/view/monitor/multi-log.ts":
/*!***********************************************!*\
  !*** ./src/main/js/view/monitor/multi-log.ts ***!
  \***********************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('mll', 'monitor');
/**
 * @hidden
 */
var MultiLogMonitorView = /** @class */ (function (_super) {
    __extends(MultiLogMonitorView, _super);
    function MultiLogMonitorView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onValueUpdate_ = _this.onValueUpdate_.bind(_this);
        _this.formatter_ = config.formatter;
        _this.element.classList.add(className());
        var textareaElem = document.createElement('textarea');
        textareaElem.classList.add(className('i'));
        textareaElem.readOnly = true;
        _this.element.appendChild(textareaElem);
        _this.textareaElem_ = textareaElem;
        config.value.emitter.on('update', _this.onValueUpdate_);
        _this.value = config.value;
        _this.update();
        return _this;
    }
    MultiLogMonitorView.prototype.dispose = function () {
        this.textareaElem_ = DisposingUtil.disposeElement(this.textareaElem_);
        _super.prototype.dispose.call(this);
    };
    MultiLogMonitorView.prototype.update = function () {
        var _this = this;
        var elem = this.textareaElem_;
        if (!elem) {
            throw pane_error_1.default.alreadyDisposed();
        }
        var shouldScroll = elem.scrollTop === elem.scrollHeight - elem.clientHeight;
        elem.textContent = this.value.rawValues
            .map(function (value) {
            return _this.formatter_.format(value);
        })
            .join('\n');
        if (shouldScroll) {
            elem.scrollTop = elem.scrollHeight;
        }
    };
    MultiLogMonitorView.prototype.onValueUpdate_ = function () {
        this.update();
    };
    return MultiLogMonitorView;
}(view_1.default));
exports.default = MultiLogMonitorView;


/***/ }),

/***/ "./src/main/js/view/monitor/single-log.ts":
/*!************************************************!*\
  !*** ./src/main/js/view/monitor/single-log.ts ***!
  \************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ../view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('sgl', 'monitor');
/**
 * @hidden
 */
var SingleLogMonitorView = /** @class */ (function (_super) {
    __extends(SingleLogMonitorView, _super);
    function SingleLogMonitorView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onValueUpdate_ = _this.onValueUpdate_.bind(_this);
        _this.formatter_ = config.formatter;
        _this.element.classList.add(className());
        var inputElem = document.createElement('input');
        inputElem.classList.add(className('i'));
        inputElem.readOnly = true;
        inputElem.type = 'text';
        _this.element.appendChild(inputElem);
        _this.inputElem_ = inputElem;
        config.value.emitter.on('update', _this.onValueUpdate_);
        _this.value = config.value;
        _this.update();
        return _this;
    }
    SingleLogMonitorView.prototype.dispose = function () {
        this.inputElem_ = DisposingUtil.disposeElement(this.inputElem_);
        _super.prototype.dispose.call(this);
    };
    SingleLogMonitorView.prototype.update = function () {
        if (!this.inputElem_) {
            throw pane_error_1.default.alreadyDisposed();
        }
        var values = this.value.rawValues;
        this.inputElem_.value =
            values.length > 0
                ? this.formatter_.format(values[values.length - 1])
                : '';
    };
    SingleLogMonitorView.prototype.onValueUpdate_ = function () {
        this.update();
    };
    return SingleLogMonitorView;
}(view_1.default));
exports.default = SingleLogMonitorView;


/***/ }),

/***/ "./src/main/js/view/root.ts":
/*!**********************************!*\
  !*** ./src/main/js/view/root.ts ***!
  \**********************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../misc/class-name */ "./src/main/js/misc/class-name.ts");
var DisposingUtil = __webpack_require__(/*! ../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
var view_1 = __webpack_require__(/*! ./view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('rot');
/**
 * @hidden
 */
var RootView = /** @class */ (function (_super) {
    __extends(RootView, _super);
    function RootView(document, config) {
        var _this = _super.call(this, document) || this;
        _this.onFolderChange_ = _this.onFolderChange_.bind(_this);
        _this.folder_ = config.folder;
        if (_this.folder_) {
            _this.folder_.emitter.on('change', _this.onFolderChange_);
        }
        _this.element.classList.add(className());
        var folder = _this.folder_;
        if (folder) {
            var titleElem = document.createElement('button');
            titleElem.classList.add(className('t'));
            titleElem.textContent = folder.title;
            _this.element.appendChild(titleElem);
            var markElem = document.createElement('div');
            markElem.classList.add(className('m'));
            titleElem.appendChild(markElem);
            _this.titleElem_ = titleElem;
        }
        var containerElem = document.createElement('div');
        containerElem.classList.add(className('c'));
        _this.element.appendChild(containerElem);
        _this.containerElem_ = containerElem;
        _this.applyModel_();
        return _this;
    }
    Object.defineProperty(RootView.prototype, "titleElement", {
        get: function () {
            return this.titleElem_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(RootView.prototype, "containerElement", {
        get: function () {
            if (!this.containerElem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.containerElem_;
        },
        enumerable: true,
        configurable: true
    });
    RootView.prototype.dispose = function () {
        this.containerElem_ = DisposingUtil.disposeElement(this.containerElem_);
        this.folder_ = null;
        this.titleElem_ = DisposingUtil.disposeElement(this.titleElem_);
        _super.prototype.dispose.call(this);
    };
    RootView.prototype.applyModel_ = function () {
        var expanded = this.folder_ ? this.folder_.expanded : true;
        var expandedClass = className(undefined, 'expanded');
        if (expanded) {
            this.element.classList.add(expandedClass);
        }
        else {
            this.element.classList.remove(expandedClass);
        }
        // TODO: Animate
    };
    RootView.prototype.onFolderChange_ = function () {
        this.applyModel_();
    };
    return RootView;
}(view_1.default));
exports.default = RootView;


/***/ }),

/***/ "./src/main/js/view/separator.ts":
/*!***************************************!*\
  !*** ./src/main/js/view/separator.ts ***!
  \***************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var class_name_1 = __webpack_require__(/*! ../misc/class-name */ "./src/main/js/misc/class-name.ts");
var view_1 = __webpack_require__(/*! ./view */ "./src/main/js/view/view.ts");
var className = class_name_1.default('spt');
/**
 * @hidden
 */
var SeparatorView = /** @class */ (function (_super) {
    __extends(SeparatorView, _super);
    function SeparatorView(document) {
        var _this = _super.call(this, document) || this;
        _this.element.classList.add(className());
        var hrElem = document.createElement('hr');
        hrElem.classList.add(className('r'));
        _this.element.appendChild(hrElem);
        return _this;
    }
    return SeparatorView;
}(view_1.default));
exports.default = SeparatorView;


/***/ }),

/***/ "./src/main/js/view/view.ts":
/*!**********************************!*\
  !*** ./src/main/js/view/view.ts ***!
  \**********************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var DisposingUtil = __webpack_require__(/*! ../misc/disposing-util */ "./src/main/js/misc/disposing-util.ts");
var pane_error_1 = __webpack_require__(/*! ../misc/pane-error */ "./src/main/js/misc/pane-error.ts");
/**
 * @hidden
 */
var View = /** @class */ (function () {
    function View(document) {
        this.disposed_ = false;
        this.doc_ = document;
        this.elem_ = this.doc_.createElement('div');
    }
    Object.defineProperty(View.prototype, "disposed", {
        get: function () {
            return this.disposed_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(View.prototype, "document", {
        get: function () {
            if (!this.doc_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.doc_;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(View.prototype, "element", {
        get: function () {
            if (!this.elem_) {
                throw pane_error_1.default.alreadyDisposed();
            }
            return this.elem_;
        },
        enumerable: true,
        configurable: true
    });
    View.prototype.dispose = function () {
        this.doc_ = null;
        this.elem_ = DisposingUtil.disposeElement(this.elem_);
        this.disposed_ = true;
    };
    return View;
}());
exports.default = View;


/***/ }),

/***/ "./src/main/sass/bundle.scss":
/*!***********************************!*\
  !*** ./src/main/sass/bundle.scss ***!
  \***********************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

exports = module.exports = __webpack_require__(/*! ../../../node_modules/css-loader/lib/css-base.js */ "./node_modules/css-loader/lib/css-base.js")(false);
// imports


// module
exports.push([module.i, ".tp-fldv_t,.tp-rotv_t{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:rgba(200,202,208,0.1);color:#c8cad0;cursor:pointer;display:block;height:24px;line-height:24px;overflow:hidden;padding-left:30px;position:relative;text-align:left;text-overflow:ellipsis;white-space:nowrap;width:100%}.tp-fldv_t:hover,.tp-rotv_t:hover{background-color:rgba(200,202,208,0.15)}.tp-fldv_t:focus,.tp-rotv_t:focus{background-color:rgba(200,202,208,0.2)}.tp-fldv_t:active,.tp-rotv_t:active{background-color:rgba(200,202,208,0.25)}.tp-fldv_m,.tp-rotv_m{background:linear-gradient(to left, #c8cad0, #c8cad0 2px, transparent 2px, transparent 4px, #c8cad0 4px, #c8cad0);border-radius:2px;bottom:0;content:'';display:block;height:6px;left:12px;margin:auto;position:absolute;top:0;-webkit-transform:rotate(90deg);transform:rotate(90deg);transition:-webkit-transform 0.2s ease-in-out;transition:transform 0.2s ease-in-out;transition:transform 0.2s ease-in-out, -webkit-transform 0.2s ease-in-out;width:6px}.tp-fldv.tp-fldv-expanded .tp-fldv_m,.tp-rotv.tp-rotv-expanded .tp-rotv_m{-webkit-transform:none;transform:none}.tp-fldv_c>.tp-fldv:first-child,.tp-rotv_c>.tp-fldv:first-child{margin-top:-4px}.tp-fldv_c>.tp-fldv:last-child,.tp-rotv_c>.tp-fldv:last-child{margin-bottom:-4px}.tp-fldv_c>*+*,.tp-rotv_c>*+*{margin-top:4px}.tp-fldv_c>.tp-fldv+.tp-fldv,.tp-rotv_c>.tp-fldv+.tp-fldv{margin-top:0}.tp-fldv_c>.tp-sptv+.tp-sptv,.tp-rotv_c>.tp-sptv+.tp-sptv{margin-top:0}.tp-btnv{padding:0 4px}.tp-btnv_b{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:#adafb8;border-radius:2px;color:#2f3137;cursor:pointer;display:block;font-weight:bold;height:20px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;width:100%}.tp-btnv_b:hover{background-color:#bbbcc4}.tp-btnv_b:focus{background-color:#c8cad0}.tp-btnv_b:active{background-color:#d6d7db}.tp-dfwv{position:absolute;top:8px;right:8px;width:256px}.tp-fldv_c{border-left:rgba(200,202,208,0.1) solid 4px;box-sizing:border-box;height:0;opacity:0;overflow:hidden;padding-bottom:0;padding-top:0;position:relative;transition:height 0.2s ease-in-out, opacity 0.2s linear, padding 0.2s ease-in-out}.tp-fldv_t:hover+.tp-fldv_c{border-left-color:rgba(200,202,208,0.15)}.tp-fldv_t:focus+.tp-fldv_c{border-left-color:rgba(200,202,208,0.2)}.tp-fldv_t:active+.tp-fldv_c{border-left-color:rgba(200,202,208,0.25)}.tp-fldv.tp-fldv-expanded .tp-fldv_c{opacity:1;overflow:visible;padding-bottom:4px;padding-top:4px;-webkit-transform:none;transform:none;transition:height 0.2s ease-in-out, opacity 0.2s linear 0.2s, padding 0.2s ease-in-out}.tp-ckbiv_l{display:block;position:relative}.tp-ckbiv_i{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background:red;left:0;opacity:0;position:absolute;top:0}.tp-ckbiv_m{background-color:rgba(200,202,208,0.15);border-radius:2px;cursor:pointer;display:block;height:20px;position:relative;width:20px}.tp-ckbiv_m::before{background-color:#c8cad0;border-radius:2px;bottom:4px;content:'';display:block;left:4px;opacity:0;position:absolute;right:4px;top:4px}.tp-ckbiv_i:hover+.tp-ckbiv_m{background-color:rgba(200,202,208,0.15)}.tp-ckbiv_i:focus+.tp-ckbiv_m{background-color:rgba(200,202,208,0.25)}.tp-ckbiv_i:active+.tp-ckbiv_m{background-color:rgba(200,202,208,0.35)}.tp-ckbiv_i:checked+.tp-ckbiv_m::before{opacity:1}.tp-clpiv{background-color:#2f3137;border-radius:6px;box-shadow:0 2px 4px rgba(0,0,0,0.2);display:none;padding:4px;position:relative;visibility:hidden;z-index:1000}.tp-clpiv.tp-clpiv-expanded{display:block;visibility:visible}.tp-clpiv_pl{display:flex}.tp-clpiv_h{margin-left:4px}.tp-clpiv_is{display:flex;margin-top:4px}.tp-clpiv_iw{align-items:center;display:flex}.tp-clpiv_iw+.tp-clpiv_iw{margin-left:4px}.tp-clpiv_il{color:rgba(200,202,208,0.8);margin-left:4px;margin-right:8px}.tp-clpiv_i{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:rgba(200,202,208,0.15);border-radius:2px;box-sizing:border-box;color:#c8cad0;font-family:inherit;height:20px;line-height:20px;width:100%;padding:0 4px;width:100%}.tp-clpiv_i:hover{background-color:rgba(200,202,208,0.15)}.tp-clpiv_i:focus{background-color:rgba(200,202,208,0.25)}.tp-clpiv_i:active{background-color:rgba(200,202,208,0.35)}.tp-hpliv{border-radius:2px;overflow:hidden;position:relative}.tp-hpliv_c{cursor:crosshair;display:block;height:80px;width:20px}.tp-hpliv_m{border-radius:100%;border:rgba(255,255,255,0.75) solid 1px;box-shadow:0 1px 2px rgba(0,0,0,0.1);height:4px;left:50%;margin-left:-3px;margin-top:-3px;pointer-events:none;position:absolute;width:4px}.tp-svpiv{border-radius:2px;overflow:hidden;position:relative}.tp-svpiv_c{cursor:crosshair;display:block;height:80px;width:100%}.tp-svpiv_m{border-radius:100%;border:rgba(255,255,255,0.75) solid 1px;box-shadow:0 1px 2px rgba(0,0,0,0.1);height:4px;margin-left:-3px;margin-top:-3px;pointer-events:none;position:absolute;width:4px}.tp-lstiv{color:#c8cad0;display:block;padding:0;position:relative}.tp-lstiv_s{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:#adafb8;border-radius:2px;color:#2f3137;cursor:pointer;display:block;height:20px;line-height:20px;padding:0 4px;width:100%}.tp-lstiv_s:hover{background-color:#bbbcc4}.tp-lstiv_s:focus{background-color:#c8cad0}.tp-lstiv_s:active{background-color:#d6d7db}.tp-lstiv_m{border-color:#2f3137 transparent transparent;border-style:solid;border-width:3px;bottom:0;box-sizing:border-box;height:6px;margin:auto;pointer-events:none;position:absolute;right:6px;top:3px;width:6px}.tp-sldiv{color:#c8cad0;display:block;padding:0}.tp-sldiv_o{box-sizing:border-box;cursor:pointer;height:20px;margin:0 6px;position:relative}.tp-sldiv_o::before{background-color:rgba(200,202,208,0.2);border-radius:1px;bottom:0;content:'';display:block;height:2px;left:0;margin:auto;position:absolute;right:0;top:0}.tp-sldiv_i{height:100%;left:0;position:absolute;top:0}.tp-sldiv_i::before{background-color:#adafb8;border-radius:2px;bottom:0;content:'';display:block;height:12px;margin:auto;position:absolute;right:-6px;top:0;width:12px}.tp-sldiv_o:hover .tp-sldiv_i::before{background-color:#bbbcc4}.tp-sldiv_o:focus .tp-sldiv_i::before{background-color:#c8cad0}.tp-sldiv_o:active .tp-sldiv_i::before{background-color:#d6d7db}.tp-txtiv{color:#c8cad0;display:block;padding:0}.tp-txtiv_i{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:rgba(200,202,208,0.15);border-radius:2px;box-sizing:border-box;color:#c8cad0;font-family:inherit;height:20px;line-height:20px;width:100%;padding:0 4px}.tp-txtiv_i:hover{background-color:rgba(200,202,208,0.15)}.tp-txtiv_i:focus{background-color:rgba(200,202,208,0.25)}.tp-txtiv_i:active{background-color:rgba(200,202,208,0.35)}.tp-cswiv_sw{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:rgba(200,202,208,0.15);border-radius:2px;box-sizing:border-box;color:#c8cad0;font-family:inherit;height:20px;line-height:20px;width:100%}.tp-cswiv_sw:hover{background-color:rgba(200,202,208,0.15)}.tp-cswiv_sw:focus{background-color:rgba(200,202,208,0.25)}.tp-cswiv_sw:active{background-color:rgba(200,202,208,0.35)}.tp-cswiv_b{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;cursor:pointer;display:block;height:20px;left:0;margin:0;outline:none;padding:0;position:absolute;top:0;width:20px}.tp-cswiv_b:focus::after{border:rgba(255,255,255,0.75) solid 2px;border-radius:2px;bottom:0;content:'';display:block;left:0;position:absolute;right:0;top:0}.tp-cswiv_p{left:-4px;position:absolute;right:-4px;top:20px}.tp-cswtxtiv{display:flex;position:relative}.tp-cswtxtiv_s{flex-grow:0;flex-shrink:0;width:20px}.tp-cswtxtiv_t{flex:1;margin-left:4px}.tp-sldtxtiv{display:flex}.tp-sldtxtiv_s{flex:2}.tp-sldtxtiv_t{flex:1;margin-left:4px}.tp-lblv{align-items:center;display:flex;padding-left:4px;padding-right:4px}.tp-lblv_l{color:rgba(200,202,208,0.8);flex:1;-webkit-hyphens:auto;-ms-hyphens:auto;hyphens:auto;padding-left:4px;padding-right:16px}.tp-lblv_v{flex-grow:0;flex-shrink:0;width:160px}.tp-grpmv{color:#c8cad0;display:block;padding:0;position:relative}.tp-grpmv_g{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:rgba(24,24,27,0.5);border-radius:2px;box-sizing:border-box;color:rgba(200,202,208,0.7);height:20px;width:100%;display:block;height:60px}.tp-grpmv_g polyline{fill:none;stroke:rgba(200,202,208,0.7);stroke-linejoin:round}.tp-grpmv_t{font-size:0.9em;left:0;pointer-events:none;position:absolute;text-indent:4px;top:0;visibility:hidden}.tp-grpmv_t.tp-grpmv_t-valid{visibility:visible}.tp-grpmv_t::before{background-color:rgba(200,202,208,0.7);border-radius:100%;content:'';display:block;height:4px;left:-2px;position:absolute;top:-2px;width:4px}.tp-sglmv_i{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:rgba(24,24,27,0.5);border-radius:2px;box-sizing:border-box;color:rgba(200,202,208,0.7);height:20px;width:100%;padding:0 4px}.tp-mllmv_i{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:rgba(24,24,27,0.5);border-radius:2px;box-sizing:border-box;color:rgba(200,202,208,0.7);height:20px;width:100%;display:block;height:60px;line-height:20px;padding:0 4px;resize:none;white-space:pre}.tp-cswmv_sw{-webkit-appearance:none;-moz-appearance:none;appearance:none;background-color:transparent;border-width:0;font-family:inherit;font-size:inherit;font-weight:inherit;margin:0;outline:none;padding:0;background-color:rgba(24,24,27,0.5);border-radius:2px;box-sizing:border-box;color:rgba(200,202,208,0.7);height:20px;width:100%}.tp-rotv{background-color:#2f3137;border-radius:6px;box-shadow:0 2px 4px rgba(0,0,0,0.2);font-family:\"Roboto Mono\",\"Source Code Pro\",Menlo,Courier,monospace;font-size:11px;font-weight:500;text-align:left}.tp-rotv_t{border-top-left-radius:6px;border-top-right-radius:6px}.tp-rotv_m{transition:none}.tp-rotv_c{box-sizing:border-box;height:0;overflow:hidden;padding-bottom:0;padding-top:0}.tp-rotv_c>.tp-fldv:first-child .tp-fldv_t{border-top-left-radius:6px;border-top-right-radius:6px}.tp-rotv.tp-rotv-expanded .tp-rotv_c{height:auto;overflow:visible;padding-bottom:4px;padding-top:4px}.tp-sptv_r{background-color:rgba(24,24,27,0.3);border-width:0;display:block;height:4px;margin:0;width:100%}\n", ""]);

// exports


/***/ })

/******/ })["default"];
});