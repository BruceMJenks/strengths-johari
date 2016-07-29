

/*
	Generate site NAV BAR
*/
function DisplayNavBar() {
	$('#NavBar').html('<nav class="navbar navbar-inverse">' +
	' <div class="container-fluid">' +
	'   <div class="navbar-header">' +
	'     <button type="button" class="navbar-toggle" data-toggle="collapse" data-target="#myNavbar">' +
	'       <span class="icon-bar"></span>' +
	'       <span class="icon-bar"></span>' +
	'       <span class="icon-bar"></span>' +
	'     </button>' +
	'     <a class="navbar-brand" href="#"><img src="/img/red-telefon.ico"></img></a>' +
	'   </div>' +
	'   <div class="collapse navbar-collapse" id="myNavbar">' +
	'     <ul class="nav navbar-nav">' +
	'       <li><a href="/">Home</a></li>' +
	'       <li><a href="https://sites.google.com/a/pivotal.io/customer-incident" target="_blank">Process Documentation</a></li>' +
	'       <li><a href="https://sites.google.com/a/pivotal.io/customer-incident/home/training" target="_blank">Sandbox Site</a></li>' +
	'       <li><a href="/intel">Intel</a></li>' +
	'     </ul>' +
	'     <ul class="nav navbar-nav navbar-right">' +
	'       <li><a href="/logout">Logout</a></li>' +
	'     </ul>' +
	'   </div>' +
	' </div>' +
	'/nav>');
}


/*
	given a id of a select input field we populate from databse

	Sends get request to URL
	Response = []string
*/
function PopulateSelect(ID, URL, DEFAULT){

	$.ajax({
		url: URL,
		type: 'get',
		dataType: 'json',
		success: function (data) {
			 var selectOptions = "";
			 for (var i = 0; i < data.length; i++) {
			 	selectOptions += "<option>" + data[i] + "</option>";
			 }
			 $('#' + ID).html(selectOptions);

			 if ( DEFAULT ) {
			 	 $('#' + ID).val(DEFAULT);
			 }
		},
		error: function(data) {
			alert('FAILED to fetch ' + ID + ': ' + data.responseJSON.errmessage);
		}
	 });
};

function PopulateSelectPost(ID, URL, DATA, VALUE){

	var target = document.getElementById('PopulateSelectPostSpinner');
	var spinner = new Spinner(SpinOpts).spin( target );
	$.ajax({
		url: URL,
		type: 'post',
		dataType: 'json',
		success: function (data) {
			 var selectOptions = "";
			 for (var i = 0; i < data.length; i++) {
			 	selectOptions += "<option>" + data[i] + "</option>";
			 }
			 $('#' + ID).html(selectOptions);

			 if ( VALUE ) {
				 $('#'+ ID).val(VALUE);
			 }
			 spinner.stop(target);
		},
		error: function(data) {
			spinner.stop(target);
			alert('FAILED to fetch ' + ID + ': ' + data.responseJSON.errmessage);
		},
		data: JSON.stringify(DATA)
	 });
};

/*
	Return link for zendesk or CSI based on ticket number
*/
function GetTicketLink(ticket) {
	if (ticket > 70000000) {
		return "https://support.emc.com/servicecenter/srManagement/" + parseInt(ticket);
	} else {
		return "https://discuss.zendesk.com/agent/tickets/" + parseInt(ticket);
	}
}


/*
	Convert date into formated string
	Tuesday, Apr 26, 2016, 9:01 PM
*/
function DateToString(d) {
	var options = {
    weekday: "long", year: "numeric", month: "short",
    day: "numeric", hour: "2-digit", minute: "2-digit"
	};
 	return d.toLocaleTimeString("en-us", options) + " UTC";
}


function GetStoryFromOpen(ticket) {
	for ( var i = 0; i < OpenStories.length; i++ ) {
		if ( OpenStories[i].ticket == ticket ) {
			return OpenStories[i];
		}
	}
	return null
}

/*

*/
function TryToPopulateForm(ticket) {
	var DATA = {
		"ticket": ticket
	};

	$('#ProductServiceName').html('');
	var target = document.getElementById('TicketDetailsSpinner');
	var spinner = new Spinner(SpinOpts).spin( target );

	$.ajax({
		url: GetTicketDetailsURl,
		type: 'post',
		dataType: 'json',
		success: function (data) {
			$('#Customer').val(data.customername);
			var v = $('#Version').html();
			if ( $('#Product').val() != data.productname ) {
				$('#Product').val(data.productname);
				PopulateSelectPost('Version', ProductVersionsURL, {"product": data.productname}, data.productversion);
			}
			$('#Version').val(data.productversion)
			spinner.stop(target);
			GetProductServiceName(data.productname, ticket, 'ProductServiceName');
			ToggleProductEmailOverride();
		},
		error: function(data) {
			spinner.stop(target);
			alert(data.responseJSON.errmessage);
		},
		data: JSON.stringify(DATA)
	 });
}

/*
	Given a ticket number get the product service name and set it if exists
*/
function GetProductServiceName(product, ticket, id) {

	var DATA = {
		"ticket": ticket,
		"product": product
	};

	$.ajax({
		url: GetProductServiceNameURL,
		type: 'post',
		dataType: 'json',
		success: function (data) {
			if ( data.hasservicename ) {
				$('#' + id).html(data.servicename);
			}
		},
		error: function(data) {
			alert(data.responseJSON.errmessage);
		},
		data: JSON.stringify(DATA)
	 });
}

/*
	Toggle Product Email override
	
	- DATA only request email to be sent if product is in production and impact is persistent 
	- PaaS requires email to be sent for every customer incident no matter what the customer impact is 
*/
function ToggleProductEmailOverride(){
	var DATA = {
		"product": $('#Product').val()
	};
	
	$.ajax({
		url: ProductInfoURL,
		type: 'post',
		dataType: 'json',
		success: function (data) {
			if ( data.emailoverride ) {
				$('#Production').val('yes');
				$('#ProductionFormField').hide();
				$('#CurrentImpact').val('yes');
				$('#CurrentImpactFormField').hide();
			} else {
				$('#Production').val('no');
				$('#ProductionFormField').show();
				$('#CurrentImpact').val('no');
				$('#CurrentImpactFormField').show();
			}
		},
		error: function(data) {
			alert(data.responseJSON.errmessage);
		},
		data: JSON.stringify(DATA)
	 });
}
