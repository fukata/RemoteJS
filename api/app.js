
/**
 * Module dependencies.
 */

var express = require('express');
var request = require('request');
var appconfig = require('./appconfig').appconfig;

var app = module.exports = express.createServer();

// Configuration

app.configure(function(){
  app.set('views', __dirname + '/views');
  app.set('view engine', 'ejs');
  app.use(express.bodyParser());
  app.use(express.methodOverride());
  app.use(app.router);
  app.use(express.static(__dirname + '/public'));
});

app.configure('development', function(){
  app.use(express.errorHandler({ dumpExceptions: true, showStack: true })); 
});

app.configure('production', function(){
  app.use(express.errorHandler()); 
});

// Routes

app.get('/', function(req, res){
  res.render('index', {
    title: 'Demo'
  });
});

app.post('/js', function(req, res){
  var url = '';
  var js = '';
  if (typeof req.body != 'undefined') {
    url = (typeof req.body.url != 'undefined') ? req.body.url : '';
    js = (typeof req.body.js != 'undefined') ? req.body.js : '';
  }

  if (!url || !js) {
    res.json({});
  } else {
    request({
		method: "POST",
		uri: appconfig.executor_url + '/execute_js', 
		body: "url="+url+"&js="+js,
	}, function(err, response, body) {
		console.log(err);
		console.log(response);
		console.log(body);
		if (err) {
			res.json(err);
		} else {
			res.json(body);
		}
	});
  }
});

var port = process.argv[2] || appconfig.port || 3000;
app.listen(port);
console.log("Express server listening on port %d in %s mode", app.address().port, app.settings.env);
