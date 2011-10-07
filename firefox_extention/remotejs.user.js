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
// 1.0 初回リリース

// ==============================================
// Options
// ==============================================
var G_OPTIONS = {
	// 実行対象のIDのパラメーターキー
	// string: __eid
	"execute_id_param_key": "__eid",
	"api_url": "http://localhost:1975"
};
// ==============================================
jQuery(function(){
	var execId = getExecuteId();
	console.log("ExecuteID=%s",execId);
	if (!execId) return;

	var js = jQuery.ajax({
		type: "GET",
		url: G_OPTIONS.api_url + "/internal/js?id="+execId,
		async: false,
		cache: false,
		dataType: "text"
	}).responseText;
	console.log(js);

	var result = {};
	try {
		var func = eval('(' + js + ')');
		result = func();
	} catch (e) {
		console.err(e);
	}

	console.log("update json");
	console.log(result);
	jQuery.post(
		G_OPTIONS.api_url + "/internal/update_json",
		{id: execId, json: JSON.stringify(result)},
		function(data) {
			//window.close();
		}
	);
});

function getExecuteId() {
	var queries = getQueryParams();
	if (queries && G_OPTIONS.execute_id_param_key in queries) {
		return queries[G_OPTIONS.execute_id_param_key];
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
