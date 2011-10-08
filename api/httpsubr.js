// httpsubr.js
// HTTP関連のサブルーチン

var http = require('http'),
    iconv = require('iconv'),
    url = require('url');

// Bufferを連結する
function concatBuffer(src1 /* , src2, ... */) {
    var i, buf, start;
    var len = 0;

    for (i = 0; i < arguments.length; ++i) {
        len += arguments[i].length;
    }

    buf = new Buffer(len);
    start = 0;
    for (i = 0; i < arguments.length; ++i) {
        var chunk = arguments[i];
        chunk.copy(buf, start, 0);
        start += chunk.length;
    }

    return buf;
}

// HTTPレスポンスとBufferからエンコーディングを検出し
// レスポンスボディを文字列で返す
exports.convertCharset = function(response, buf) {
    var charset = null;

    var content_type = response.headers['content-type'];
    if (content_type) {
        re = content_type.match(/\bcharset=([\w\-]+)\b/i);
        if (re) {
            charset = re[1];
        }
    }

    if (!charset) {
        var bin = buf.toString('binary');
        re = bin.match(/<meta\b[^>]*charset=([\w\-]+)/i);
        if (re) {
            charset = re[1];
        } else {
            charset = 'utf-8';
        }
    }

    switch (charset) {
    case 'ascii':
    case 'utf-8':
        return buf.toString(charset);
        break;
    case 'euc-jp':
    case 'EUC-JP':
    case 'EUCJP':
    	var ic = new (iconv.Iconv)('EUCJP', 'UTF-8//TRANSLIT//IGNORE');
    	var buf2 = ic.convert(buf);
    	return buf2.toString('utf8');
    	break;

    default:
        var ic = new (iconv.Iconv)(charset, 'utf-8');
        var buf2 = ic.convert(buf);
        return buf2.toString('utf8');
        break;
    }
}

// 文字列ではなくBufferを返す版の
// requestパッケージ (https://github.com/mikeal/node-utils) のrequest関数
// とほぼ同等な関数
// request関数にある一部機能は実装していない
exports.httpRequest = function(options, callback) {
    options = options || {};
    if (typeof(options.uri) == 'string') {
        options.uri = url.parse(options.uri);
    }
    options.method = options.method || 'GET';
    options.headers = options.headers || {};
    options._nRedirect = options._nRedirect || 0;
    if (typeof(options.maxRedirects) == 'undefined') {
        options.maxRedirects = 10;
    }

    if (!options.headers.host) {
        options.headers.host = options.uri.hostname;
        if (options.uri.port) {
            options.headers.host += ':' + options.uri.port;
        }
    }

    var port = 80;
    var https = false;
    if (options.uri.protocol == 'https:') {
        port = 443;
        https = true;
    }
    if (options.uri.port) {
        port = port;
    }

    var path = (options.uri.pathname ? options.uri.pathname : '/');
    if (options.uri.search) {
        path += options.uri.search;
    }
    if (options.uri.hash) {
        path += options.uri.hash;
    }

    var client = http.createClient(port, options.uri.hostname, https);
    client.addListener('error', function(err) {
        if (callback) {
            callback(err);
        } else {
            throw err;
        }
    });

    var request = client.request(options.method, path, options.headers);
    request.addListener('response', function(response) {
        if (response.headers.location) {
            if (options._nRedirect++ >= options.maxRedirect) {
                client.emit('error', new Error('Too many redirects'));
            }
            var loc = response.headers.location;
            if (!loc.match(/^https?:/i)) {
                loc = url.resolve(options.uri.href, response.headers.location);
            }
            options.uri = loc;
            exports.httpRequest(options, callback);
        } else {
            var chunks = [];
            response.on('data', function(chunk) {
                chunks.push(chunk);
            })
            .on('end', function() {
                if (callback) {
                    var buf = concatBuffer.apply({}, chunks);
                    delete(chunks);
                    callback(null, response, buf);
                }
            });
        }
    });

    if (options.requestBody) {
        if (typeof(options.requestBody) == 'string') {
            request.write(options.requestBody);
            request.end();
        } else {
            sys.pump(options.requestBody, request);
        }
    } else {
        request.end();
    }
}

exports.get = exports.httpRequest;

exports.post = function(options, callback) {
    if (!options.requestBody) {
        options.requestBody = '';
    }
    exports.request(options, callback);
}
