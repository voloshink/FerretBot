var fs = require('fs');
var request = require('request');
var WebSocketClient = require('websocket').client;

var Ferret = function() {
    this.whitelist = [];
    this.loadWhitelist();
    this.ratelimit = 60;
    this.lastMessage = '';
    this.lastSent = new Date();
    this.socket = new WebSocketClient();

    this.socket.connect('ws://www.destiny.gg/ws', null, '*', {
        'Cookie': 'authtoken=<api token>;'
    });

    this.socket.on('connect', function(connection) {
        this.connection = connection;

        connection.on('error', function(err) {
            console.log('Socket error ' + error.toString());
        })

        connection.on('close', function() {
            console.log('Connection closed');
        })

        connection.on('message', function(message) {
            if (message.type !== 'utf8') return;
            message = message.utf8Data;

            if (message.slice(0, 3) !== 'MSG') return;
            message = message.slice(4);

            try {
                var json = JSON.parse(message);
            } catch (e) {
                return e;
            }

            this.handleMessage(json.nick, json.data);
        }.bind(this))
    }.bind(this));
}

Ferret.prototype.handleMessage = function(nick, message) {
    nick = nick.toLowerCase();
    message = message.toLowerCase();
    if (this.whitelist.indexOf(nick) === -1) return;

    if (message === '!ferret') { 
        this.getFerret(function(url) {
            this.send('FerretLOL ' + url + ' FerretLOL');
        }.bind(this));
        return;
    }

    if (nick !== 'polecat') return;

    var arr = message.split(' ');

    if (arr[0] === '!fwhitelist' && arr[1]) { 
        this.whitelist.push(arr[1]);
        this.send('FerretLOL ' + arr[1], true);
        this.saveWhiteList();
        return;
    }

    if (arr[0] === '!fratelimit' && arr[1]) this.ratelimit = parseInt(arr[1], 10);
}

Ferret.prototype.loadWhitelist = function() {
    fs.readFile('whitelist.txt', 'utf8', function(err, data) {
        this.whitelist = data.split(',');
    }.bind(this));
}

Ferret.prototype.saveWhiteList = function() {
    console.log(this.whitelist.join(','));
    fs.writeFile('whitelist.txt', this.whitelist.join(','), function(err) {
        console.log(err);
    });
}

Ferret.prototype.getFerret = function(callback) {
    var options = {
        url: 'http://polecat.me/api/ferret',
        method: 'GET',
        json: true
    }

    request(options, function(err, response, body) {
        console.log(body.url);
        if (!err && response.statusCode === 200 && body && body.url) callback(body.url);
    });
}

Ferret.prototype.send = function(msg, force) {
    var now = new Date();
    if ((msg === this.lastMessage || dateDiff(now, this.lastSent) < this.ratelimit) && !force) return;

    this.lastMessage = msg;
    if (!force) this.lastSent = now;
    this.connection.sendUTF('MSG {"data":"' + msg + '"}');
}

new Ferret();


function dateDiff(d1, d2) {
    return (Math.abs(d1.getTime() - d2.getTime()) / 1000);
}