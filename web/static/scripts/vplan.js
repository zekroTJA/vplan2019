// 'use strict';
var months = ["Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"];
var days = ["Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Sonnabend"];

function formatDate(date) {
	var selectedDate = new Date(date);
	var dateFormat = date.split('-');
	return dateFormat[2].split('T')[0] + '. ' + months[selectedDate.getMonth()] + ' <span class="badge badge-dark" id="day1c">' + days[selectedDate.getDay()] + '</span>';
}

function createVplanEntry(id, entry) {
	var list_item = document.createElement("li");
	list_item.setAttribute('class', 'list-group-item d-flex justify-content-between lh-condensed');

	var desc = document.createElement("div");
	var head = document.createElement("h6");
	head.setAttribute('class', 'my-0');
	head.textContent = entry.class;
	desc.appendChild(head);
	var tiny = document.createElement("small");
	tiny.textContent = entry.time + ' - ' + entry.responsible;
	desc.appendChild(tiny);
	list_item.appendChild(desc);

	var span = document.createElement("span");
	span.setAttribute('class', 'text-muted');
	span.textContent = entry.messures;
	list_item.appendChild(span);

	_(id).appendChild(list_item);
}

function getDataForVplan(method, url, args) {
	getJson(method, url, args, function() {
		console.log(this);
		if(this.data != undefined) {
			_("day0").innerHTML = formatDate(this.data[0].date_for);
			_("day0e").innerHTML = '';
			(this.data[0].entries == null) ? console.log(this.data[0].entries) : this.data[0].entries.forEach( function(entry) { createVplanEntry("day0e", entry); });
			_("day1").innerHTML = formatDate(this.data[1].date_for);
			_("day1e").innerHTML = '';
			(this.data[1].entries == null) ? console.log(this.data[1].entries) : this.data[1].entries.forEach( function(entry) { createVplanEntry("day1e", entry); });
			_("day2").innerHTML = formatDate(this.data[2].date_for);
			_("day2e").innerHTML = '';
			(this.data[2].entries == null) ? console.log(this.data[2].entries) : this.data[2].entries.forEach( function(entry) { createVplanEntry("day2e", entry); });
		} else if(this.error != undefined) {
			console.log(this.error.code)
		} else {
			// TODO
		}
	});
	setTimeout(function() {getDataForVplan(method, url, args);}, 20000);
}

function createNewsEntry(id, entry) {
	var list_item = document.createElement("div");
	list_item.setAttribute('class', 'alert alert-secondary');
	list_item.setAttribute('role', 'alert');

	var head = document.createElement("h6");
	head.setAttribute('class', 'alert-heading');
	head.innerHTML = '<b>' + formatDate(entry.date) + ': ' + entry.headline + '</b>';
	list_item.appendChild(head);
	var desc = document.createElement("div");
	desc.textContent = entry.short + ' - ' + entry.story;
	list_item.appendChild(desc);

	_(id).appendChild(list_item);
}

function getDataForNews(method, url, args) {
	getJson(method, url, args, function() {
		console.log(this);
		if(this.data != undefined) {
			_("news").innerHTML = '';
			this.data.forEach( function(entry) { createNewsEntry("news", entry); });
		} else if(this.error != undefined) {
			console.log(this.error.code)
		} else {
			// TODO
		}
	});
	setTimeout(function() {getDataForNews(method, url, args);}, 20000);
}