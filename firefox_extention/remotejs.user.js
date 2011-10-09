// ==UserScript==
// @name			RemoteJs	
// @namespace		org.fukata.ff.us.remotejs
// @description		Remote execute javascript
// @author			@fukata
// @include			http://*
// @include			https://*
// @require			http://ajax.googleapis.com/ajax/libs/jquery/1.6.4/jquery.min.js
// ==/UserScript==
// ver 
// 0.1 Beta Release

// ==============================================
// Options
// ==============================================
var G_OPTIONS = {
	// 実行対象のIDのパラメーターキー
	// string: __eid
	"execute_id_param_key": "__eid",
	// js_executorサーバのAPIURL
	"executor_url": "http://localhost:1975"
};
// ==============================================

$(function(){
	if (window.opener || window != window.parent) {
		return;
	}

	var execId = getExecuteId();
	if (!execId) return;

	var js = $.ajax({
		type: "GET",
		url: G_OPTIONS.executor_url + "/internal/js?id="+execId,
		async: false,
		cache: false,
		dataType: "text"
	}).responseText;

	var result = {};
	try {
		var func = eval('(' + js + ')');
		result = func();
	} catch (e) {
		console.err(e);
	}

	console.log("UpdateJson ID=%s", execId);
	console.log(result);
	$.ajax({
		type: "POST",
		cache: false,
		url: G_OPTIONS.executor_url + "/internal/update_json",
		data: {id: execId, json: JSON.stringify(result)},
		complete: function(request, textStatus) {
			window.close();
		}
	});
});

function getexecuteid() {
	var queries = getqueryparams();
	if (queries && g_options.execute_id_param_key in queries) {
		return queries[g_options.execute_id_param_key];
	} else {
		return false;
	}
}

function getQueryParams() {
	var qs=location.search;
	if (qs) {
		var qsa=qs.substring(1).split('&');
		var params={};
		for(var i=0; i<qsa.length; i++) {
			var pair=qsa[i].split('=');
			if (pair[0]) {
				params[pair[0]]=decodeURIComponent(pair[1]);
			}
		}
		return params;
	}
	return null;
}
