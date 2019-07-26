﻿//app constructor function
function app() {
    var self = this;
    self.errorState = ko.observable(false);
    self.errorMessage = ko.observable("");
    self.testSuites = ko.observableArray([]);
}

// test result for a client
function testClientResult(pass,details,name,version,instantiated,log) {
    this.pass = ko.observable(pass);
    this.details = ko.observable(details);
    this.clientName = ko.observable(name);
    this.clientVersion = ko.observable(version);
    this.clientInstantiated = ko.observable(instantiated);
    this.logfile = ko.observable(log);
}

function makeClientResults(clientResults, clientInfos) {
    $.map(clientResults, function (clientName, testResult) {
        var pass = testResult.pass;
        var details = testResult.details;
        var name = "Missing client info.";
        var version = "";
        var instantiated;
        var log="";
        if (clientInfos.hasOwnProperty(clientName)) {
            var clientInfo = clientInfos[clientName];
            name = clientInfo.name;
            version = clientInfo.versionInfo;
            instantiated = clientInfo.instantiatedAt;
            log = clientInfo.logFile;
        }
        return new testClientResult(pass, details, name, version, instantiated, log);
    });
}

// test case result 
function testResult(data) {
    this.pass = ko.observable(data.pass);
    this.details = ko.observable(data.details);

}

function testCase(data) {
    this.id = ko.observable(data.id);
    this.name = ko.observable(data.name);
    this.description = ko.observable(data.description);
    this.start = ko.observable(Date.parse(data.start));
    this.end = ko.observable(Date.parse(data.end));
    this.summaryResult = ko.observable(new testResult(data.summaryResult));
    this.clientResults = ko.observableArray(makeClientResults(data.clientResults, data.clientInfo));
}

function calcDuration(duration) {
    var ret = ""
    if (duration.hours() > 0) { ret = ret + duration.hours() + "hr "; }
    if (duration.minutes() > 0) { ret = ret + duration.minutes() + "min "; }
    ret = ret + duration.seconds() + "s "; 
    return ret;
}

function testSuite(data) {
    var self = this;
    self.id = ko.observable(data.id);
    self.suiteLabel = ko.computed(function () { return "Suite" + self.id(); })
    self.suiteDetailLabel = ko.computed(function () { return "CollapseSuite" + self.id(); })
    self.name = ko.observable(data.name);
    self.description = ko.observable(data.description);
    var testCases = $.map(data.testCases, function (item) { return new testCase(item) });
    self.testCases = ko.observableArray(testCases)
    var earliest = Math.min.apply(Math, testCases.map(function (tc) { return tc.start(); }));
    var latest = Math.max.apply(Math, testCases.map(function (tc) { return tc.end(); }));
    var fails = testCases.map(function (tc) { if (!tc.summaryResult().pass()) return 1; else return 0; }).reduce(function (a, b) { return a + b; },0);
    var successes = testCases.map(function (tc) { if (tc.summaryResult().pass()) return 1; else return 0; }).reduce(function (a, b) { return a + b; }, 0);
    self.started = ko.observable(earliest);
    self.ended = ko.observable(latest);
    self.passes = ko.observable(successes);
    self.fails = ko.observable(fails);
    var dur= moment.duration(  moment(self.ended()).diff(moment(self.started())));
    self.duration = ko.observable(calcDuration(dur));

}

app.prototype.LoadTestSuites = function (src) {
    var self = this;
    $.getJSON(src, function (allData) {
        var testSuites = $.map(allData, function (item) { return new testSuite(item) });
        self.testSuites(testSuites);
    });    
}