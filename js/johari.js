

/*
        Generate site NAV BAR
*/
function DisplayNavBar() {
        $('#NavBar').html('<nav class="navbar navbar-inverse">' +
                '<div class="container-fluid">' +
                  '<div class="navbar-header">' +
                    '<button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#navbar" aria-expanded="false" aria-controls="navbar">' +
                      '<span class="sr-only">Toggle navigation</span>' +
                      '<span class="icon-bar"></span>' +
                      '<span class="icon-bar"></span>' +
                      '<span class="icon-bar"></span>' +
                    '</button>' +
                    '<a class="navbar-brand" href="/">GSS Johari Window</a>' +
                  '</div>' +
                  '<div id="navbar" class="navbar-collapse collapse">' +
                    '<ul class="nav navbar-nav">' +
                      '<li><a href="/window">My Current Window</a></li>' +
                      '<li class="dropdown">' +
                        '<a href="#" class="dropdown-toggle" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false">Previous Windows <span class="caret"></span></a>' +
                        '<ul class="dropdown-menu">' +
                          '<div id="PreviousWindowItems"></div>' +
                        '</ul>' +
                      '</li>' +
                    '</ul>' +
                    '<ul class="nav navbar-nav navbar-right">' +
                      '<li><a href="/logout"><span class="glyphicon glyphicon-log-out"></span> Logout</a></li>' +
                    '</ul>' +
                  '</div><!--/.nav-collapse --> ' +
                '</div><!--/.container-fluid --> ' +
              '</nav>');
      
      
      PopulatePreviousWindows();
      
}

var SpinOpts = {
  lines: 13, // The number of lines to draw
  length: 20, // The length of each line
  width: 10, // The line thickness
  radius: 30, // The radius of the inner circle
  corners: 1, // Corner roundness (0..1)
  rotate: 0, // The rotation offset
  direction: 1, // 1: clockwise, -1: counterclockwise
  color: '#000', // #rgb or #rrggbb or array of colors
  speed: 1, // Rounds per second
  trail: 60, // Afterglow percentage
  shadow: false, // Whether to render a shadow
  hwaccel: false, // Whether to use hardware acceleration
  className: 'spinner', // The CSS class to assign to the spinner
  zIndex: 2e9, // The z-index (defaults to 2000000000)
  top: '50%', // Top position relative to parent
  left: '50%' // Left position relative to parent
};