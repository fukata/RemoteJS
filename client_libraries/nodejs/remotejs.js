var request = require('request');

function RemoteJS() {
}

RemoteJS.prototype.js = function(url, js, callback) {
	request.post({
		uri: 'http://api.fukata.org/remotejs/js',
		body: "url=" + url + "&js=" + js,
		headers: {
			'content-type': 'application/x-www-form-urlencoded'
		} 
	}, function(err, response, body) {
		var parsedJson = {};
		try { parsedJson = JSON.parse(body); } catch (e) {}
		callback(err, parsedJson);
	});
}

exports.RemoteJS = new RemoteJS();
